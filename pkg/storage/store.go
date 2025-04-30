package storage

import (
	"github.com/PranavGrandhi/Distributed-Auction-System/pkg/auction"
)

// Store defines the interface for auction storage implementations
type Store interface {
	CreateAuction(auction.AuctionItem) (auction.AuctionItem, error)
	ListAuctions() ([]auction.AuctionItem, error)
	GetAuction(id string) (auction.AuctionItem, error)
	PlaceBid(bid auction.Bid) error
	GetHighestBid(auctionID string) (auction.Bid, error)
	GetBidHistory(auctionID string) ([]auction.Bid, error)
}
