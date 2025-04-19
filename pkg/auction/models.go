package auction

import (
	"time"
)

// AuctionItem represents an item up for auction
type AuctionItem struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	MinimumBid  float64   `json:"minimum_bid"`
	ExpiryTime  time.Time `json:"expiry_time"`
	CreatedAt   time.Time `json:"created_at"`
}

// Bid represents a bid placed on an auction item
type Bid struct {
	ID            string    `json:"id"`
	ParticipantID string    `json:"participant_id"`
	AuctionItemID string    `json:"auction_item_id"`
	BidPrice      float64   `json:"bid_price"`
	Timestamp     time.Time `json:"timestamp"`
}
