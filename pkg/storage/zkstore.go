package storage

import (
	"encoding/json"
	"errors"
	"path"
	"sort"
	"time"

	"github.com/PranavGrandhi/Distributed-Auction-System/pkg/auction"
	"github.com/go-zookeeper/zk"
	"github.com/google/uuid"
)

// ZKStore provides a ZooKeeper-backed implementation of auction storage
type ZKStore struct {
	conn     *zk.Conn
	basePath string
}

// NewZKStore creates a new ZooKeeper-backed store
func NewZKStore(zkHosts []string, basePath string) (*ZKStore, error) {
	conn, _, err := zk.Connect(zkHosts, time.Second*10)
	if err != nil {
		return nil, err
	}

	store := &ZKStore{
		conn:     conn,
		basePath: basePath,
	}

	// Ensure base paths exist
	paths := []string{
		basePath,
		path.Join(basePath, "auctions"),
		path.Join(basePath, "bids"),
		path.Join(basePath, "locks"),
	}

	for _, p := range paths {
		exists, _, err := conn.Exists(p)
		if err != nil {
			return nil, err
		}

		if !exists {
			_, err := conn.Create(p, []byte{}, 0, zk.WorldACL(zk.PermAll))
			if err != nil && err != zk.ErrNodeExists {
				return nil, err
			}
		}
	}

	return store, nil
}

// Close closes the ZooKeeper connection
func (z *ZKStore) Close() {
	if z.conn != nil {
		z.conn.Close()
	}
}

// CreateAuction adds a new auction item to the store
func (z *ZKStore) CreateAuction(item auction.AuctionItem) (auction.AuctionItem, error) {
	if item.ID == "" {
		item.ID = uuid.New().String()
	}

	item.CreatedAt = time.Now()

	data, err := json.Marshal(item)
	if err != nil {
		return auction.AuctionItem{}, err
	}

	auctionPath := path.Join(z.basePath, "auctions", item.ID)
	_, err = z.conn.Create(auctionPath, data, 0, zk.WorldACL(zk.PermAll))
	if err != nil {
		return auction.AuctionItem{}, err
	}

	// Create the bids path for this auction
	bidsPath := path.Join(z.basePath, "bids", item.ID)
	_, err = z.conn.Create(bidsPath, []byte{}, 0, zk.WorldACL(zk.PermAll))
	if err != nil {
		// Try to clean up the auction node
		z.conn.Delete(auctionPath, 0)
		return auction.AuctionItem{}, err
	}

	return item, nil
}

// ListAuctions returns all auction items in the store
func (z *ZKStore) ListAuctions() ([]auction.AuctionItem, error) {
	auctionsPath := path.Join(z.basePath, "auctions")
	children, _, err := z.conn.Children(auctionsPath)
	if err != nil {
		return nil, err
	}

	auctions := make([]auction.AuctionItem, 0, len(children))
	for _, child := range children {
		itemPath := path.Join(auctionsPath, child)
		data, _, err := z.conn.Get(itemPath)
		if err != nil {
			continue // Skip items we can't retrieve
		}

		var item auction.AuctionItem
		if err := json.Unmarshal(data, &item); err != nil {
			continue // Skip items we can't unmarshal
		}

		auctions = append(auctions, item)
	}

	return auctions, nil
}

// GetAuction retrieves an auction by ID
func (z *ZKStore) GetAuction(id string) (auction.AuctionItem, error) {
	auctionPath := path.Join(z.basePath, "auctions", id)
	data, _, err := z.conn.Get(auctionPath)
	if err != nil {
		return auction.AuctionItem{}, errors.New("auction not found")
	}

	var item auction.AuctionItem
	if err := json.Unmarshal(data, &item); err != nil {
		return auction.AuctionItem{}, err
	}

	return item, nil
}

// PlaceBid adds a new bid to an auction item with distributed locking
func (z *ZKStore) PlaceBid(bid auction.Bid) error {
	// Get the auction to check if it exists and hasn't expired
	auctionItem, err := z.GetAuction(bid.AuctionItemID)
	if err != nil {
		return err
	}

	// Check if auction has expired
	if time.Now().After(auctionItem.ExpiryTime) {
		return errors.New("auction has expired")
	}

	// Check if the bid is higher than the minimum bid
	if bid.BidPrice < auctionItem.MinimumBid {
		return errors.New("bid price is lower than minimum bid")
	}

	// Create lock path
	lockPath := path.Join(z.basePath, "locks", bid.AuctionItemID)

	// Ensure parent lock path exists
	lockParentPath := path.Join(z.basePath, "locks")
	exists, _, err := z.conn.Exists(lockParentPath)
	if err != nil {
		return err
	}

	if !exists {
		_, err = z.conn.Create(lockParentPath, []byte{}, 0, zk.WorldACL(zk.PermAll))
		if err != nil && err != zk.ErrNodeExists {
			return err
		}
	}

	// Create a distributed lock using the proper API
	lock := zk.NewLock(z.conn, lockPath, zk.WorldACL(zk.PermAll))

	// Acquire the lock (this will block until lock is acquired)
	if err := lock.Lock(); err != nil {
		return err
	}

	// Make sure we release the lock when done
	defer lock.Unlock()

	// Check if there are existing bids and if the current bid is higher
	highestBid, err := z.GetHighestBid(bid.AuctionItemID)
	if err == nil {
		// There is a highest bid, check if the new bid is higher
		if bid.BidPrice <= highestBid.BidPrice {
			return errors.New("bid price is not higher than current highest bid")
		}
	} else if err.Error() != "no bids found for this auction" {
		// An error occurred that's not just "no bids"
		return err
	}

	// Generate a UUID if not provided
	if bid.ID == "" {
		bid.ID = uuid.New().String()
	}

	// Set timestamp if not provided
	if bid.Timestamp.IsZero() {
		bid.Timestamp = time.Now()
	}

	// Serialize bid data
	bidData, err := json.Marshal(bid)
	if err != nil {
		return err
	}

	// Create a sequential node for this bid
	bidPath := path.Join(z.basePath, "bids", bid.AuctionItemID, "bid-")
	_, err = z.conn.Create(bidPath, bidData, zk.FlagSequence, zk.WorldACL(zk.PermAll))

	return err
}

// GetHighestBid returns the highest bid for an auction
func (z *ZKStore) GetHighestBid(auctionID string) (auction.Bid, error) {
	bids, err := z.GetBidHistory(auctionID)
	if err != nil {
		return auction.Bid{}, err
	}

	if len(bids) == 0 {
		return auction.Bid{}, errors.New("no bids found for this auction")
	}

	// Return the highest bid by price
	sort.Slice(bids, func(i, j int) bool {
		return bids[i].BidPrice < bids[j].BidPrice
	})

	return bids[len(bids)-1], nil
}

// GetBidHistory returns all bids for an auction
func (z *ZKStore) GetBidHistory(auctionID string) ([]auction.Bid, error) {
	bidsPath := path.Join(z.basePath, "bids", auctionID)

	// Check if the auction exists
	exists, _, err := z.conn.Exists(bidsPath)
	if err != nil {
		return nil, err
	}

	if !exists {
		return nil, errors.New("auction not found")
	}

	children, _, err := z.conn.Children(bidsPath)
	if err != nil {
		return nil, err
	}

	// Sort the children by sequence number to get chronological order
	sort.Strings(children)

	bids := make([]auction.Bid, 0, len(children))
	for _, child := range children {
		bidPath := path.Join(bidsPath, child)
		data, _, err := z.conn.Get(bidPath)
		if err != nil {
			continue // Skip bids we can't retrieve
		}

		var bid auction.Bid
		if err := json.Unmarshal(data, &bid); err != nil {
			continue // Skip bids we can't unmarshal
		}

		bids = append(bids, bid)
	}

	return bids, nil
}
