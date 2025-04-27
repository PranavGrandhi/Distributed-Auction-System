package storage

import (
	"errors"
	"sync"
	"time"

	"github.com/PranavGrandhi/Distributed-Auction-System/pkg/auction"
	"github.com/google/uuid"
)

// MemoryStore provides an in-memory implementation of auction storage
type MemoryStore struct {
	auctionsMutex sync.RWMutex
	auctions      map[string]auction.AuctionItem

	bidsMutex sync.RWMutex
	bids      map[string][]auction.Bid // Map auction ID to its bids
}

// NewMemoryStore creates a new in-memory store
func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		auctions: make(map[string]auction.AuctionItem),
		bids:     make(map[string][]auction.Bid),
	}
}

// CreateAuction adds a new auction item to the store
func (m *MemoryStore) CreateAuction(item auction.AuctionItem) (auction.AuctionItem, error) {
	m.auctionsMutex.Lock()
	defer m.auctionsMutex.Unlock()

	// Generate a UUID if not provided
	if item.ID == "" {
		item.ID = uuid.New().String()
	}

	item.CreatedAt = time.Now()
	m.auctions[item.ID] = item

	// Initialize an empty bid list for this auction
	m.bidsMutex.Lock()
	m.bids[item.ID] = []auction.Bid{}
	m.bidsMutex.Unlock()

	return item, nil
}

// ListAuctions returns all auction items in the store
func (m *MemoryStore) ListAuctions() ([]auction.AuctionItem, error) {
	m.auctionsMutex.RLock()
	defer m.auctionsMutex.RUnlock()

	auctions := make([]auction.AuctionItem, 0, len(m.auctions))
	for _, item := range m.auctions {
		auctions = append(auctions, item)
	}
	return auctions, nil
}

// GetAuction retrieves an auction by ID
func (m *MemoryStore) GetAuction(id string) (auction.AuctionItem, error) {
	m.auctionsMutex.RLock()
	defer m.auctionsMutex.RUnlock()

	item, exists := m.auctions[id]
	if !exists {
		return auction.AuctionItem{}, errors.New("auction not found")
	}

	return item, nil
}

// PlaceBid adds a new bid to an auction item
func (m *MemoryStore) PlaceBid(bid auction.Bid) error {
	// Check if auction exists
	m.auctionsMutex.RLock()
	auctionItem, exists := m.auctions[bid.AuctionItemID]
	m.auctionsMutex.RUnlock()

	if !exists {
		return errors.New("auction not found")
	}

	// Check if auction has expired
	if time.Now().After(auctionItem.ExpiryTime) {
		return errors.New("auction has expired")
	}

	m.bidsMutex.Lock()
	defer m.bidsMutex.Unlock()

	// Check if the bid is higher than the minimum bid
	if bid.BidPrice < auctionItem.MinimumBid {
		return errors.New("bid price is lower than minimum bid")
	}

	// Check if there are existing bids and if current bid is higher
	bids := m.bids[bid.AuctionItemID]
	if len(bids) > 0 {
		highestBid := bids[len(bids)-1]
		if bid.BidPrice <= highestBid.BidPrice {
			return errors.New("bid price is not higher than current highest bid")
		}
	}

	// Generate a UUID if not provided
	if bid.ID == "" {
		bid.ID = uuid.New().String()
	}

	// Set timestamp if not provided
	if bid.Timestamp.IsZero() {
		bid.Timestamp = time.Now()
	}

	// Add bid to the list (acting as a queue where newest bid is at the end)
	m.bids[bid.AuctionItemID] = append(m.bids[bid.AuctionItemID], bid)

	return nil
}

// GetHighestBid returns the highest bid for an auction
func (m *MemoryStore) GetHighestBid(auctionID string) (auction.Bid, error) {
	m.bidsMutex.RLock()
	defer m.bidsMutex.RUnlock()

	bids, exists := m.bids[auctionID]
	if !exists {
		return auction.Bid{}, errors.New("auction not found")
	}

	if len(bids) == 0 {
		return auction.Bid{}, errors.New("no bids found for this auction")
	}

	// Return the last bid (highest bid)
	return bids[len(bids)-1], nil
}

// GetBidHistory returns all bids for an auction
func (m *MemoryStore) GetBidHistory(auctionID string) ([]auction.Bid, error) {
	m.bidsMutex.RLock()
	defer m.bidsMutex.RUnlock()

	bids, exists := m.bids[auctionID]
	if !exists {
		return nil, errors.New("auction not found")
	}

	// Return a copy of the bids slice to prevent modification
	result := make([]auction.Bid, len(bids))
	copy(result, bids)

	return result, nil
}
