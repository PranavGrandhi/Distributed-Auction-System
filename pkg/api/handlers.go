package api

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/PranavGrandhi/Distributed-Auction-System/pkg/auction"
	"github.com/PranavGrandhi/Distributed-Auction-System/pkg/storage"
	"github.com/gorilla/mux"
)

// Server represents the API server
type Server struct {
	Router *mux.Router
	Store  storage.Store
}

// NewZooKeeperServer creates a new API server with ZooKeeper storage
func NewZooKeeperServer(zkHosts []string) (*Server, error) {
	store, err := storage.NewZKStore(zkHosts, "/auction-system")
	if err != nil {
		return nil, err
	}

	server := &Server{
		Router: mux.NewRouter(),
		Store:  store,
	}
	server.setupRoutes()
	return server, nil
}

// NewServer creates a new API server
func NewServer() *Server {
	server := &Server{
		Router: mux.NewRouter(),
		Store:  storage.NewMemoryStore(),
	}
	server.setupRoutes()
	return server
}

// setupRoutes configures the API routes
func (s *Server) setupRoutes() {

	// Add CORS middleware
	s.Router.Use(corsMiddleware)

	// Static file handling
	s.Router.PathPrefix("/frontend/").Handler(http.StripPrefix("/frontend/", http.FileServer(http.Dir("./frontend"))))

	s.Router.HandleFunc("/", serveFrontend).Methods("GET")
	s.Router.HandleFunc("/auctions", s.CreateAuction).Methods("POST")
	s.Router.HandleFunc("/auctions", s.ListAuctions).Methods("GET")
	s.Router.HandleFunc("/auctions/{id}", s.GetAuction).Methods("GET")
	s.Router.HandleFunc("/auctions/{id}/bids", s.PlaceBid).Methods("POST")
	s.Router.HandleFunc("/auctions/{id}/status", s.QueryAuctionStatus).Methods("GET")
	s.Router.HandleFunc("/auctions/{id}/history", s.GetBidHistory).Methods("GET")
}

// corsMiddleware adds CORS headers to enable cross-origin requests

func corsMiddleware(next http.Handler) http.Handler {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		// Set CORS headers
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		// Handle preflight requests
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		next.ServeHTTP(w, r)
	})

}

func serveFrontend(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	http.ServeFile(w, r, "./frontend/index.html")
}

// CreateAuction handles requests to create a new auction item
func (s *Server) CreateAuction(w http.ResponseWriter, r *http.Request) {
	var item auction.AuctionItem
	if err := json.NewDecoder(r.Body).Decode(&item); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	// Validate required fields
	if item.Name == "" || item.MinimumBid <= 0 || item.ExpiryTime.IsZero() {
		http.Error(w, "Missing required fields: name, minimum_bid, expiry_time", http.StatusBadRequest)
		return
	}

	createdItem, err := s.Store.CreateAuction(item)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(createdItem)
}

// ListAuctions handles GET /auctions
func (s *Server) ListAuctions(w http.ResponseWriter, r *http.Request) {
	auctions, err := s.Store.ListAuctions()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(auctions)
}

// GetAuction handles requests to get an auction item by ID
func (s *Server) GetAuction(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	item, err := s.Store.GetAuction(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(item)
}

// PlaceBid handles requests to place a bid on an auction item
func (s *Server) PlaceBid(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	auctionID := vars["id"]

	var bid auction.Bid
	if err := json.NewDecoder(r.Body).Decode(&bid); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	// Validate required fields
	if bid.ParticipantID == "" || bid.BidPrice <= 0 {
		http.Error(w, "Missing required fields: participant_id, bid_price", http.StatusBadRequest)
		return
	}

	// Set the auction ID from the URL
	bid.AuctionItemID = auctionID

	// Set timestamp if not provided
	if bid.Timestamp.IsZero() {
		bid.Timestamp = time.Now()
	}

	if err := s.Store.PlaceBid(bid); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"message": "Bid placed successfully"})
}

// QueryAuctionStatus handles requests to get the current status of an auction
func (s *Server) QueryAuctionStatus(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	auctionID := vars["id"]

	// Get the auction
	auctionItem, err := s.Store.GetAuction(auctionID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	// Get the highest bid
	highestBid, err := s.Store.GetHighestBid(auctionID)

	// Prepare the response
	type AuctionStatus struct {
		Auction       auction.AuctionItem `json:"auction"`
		HighestBid    *auction.Bid        `json:"highest_bid,omitempty"`
		Status        string              `json:"status"`
		TimeRemaining string              `json:"time_remaining,omitempty"`
	}

	status := AuctionStatus{
		Auction: auctionItem,
	}

	// Set status based on auction expiry
	if time.Now().After(auctionItem.ExpiryTime) {
		status.Status = "expired"
	} else {
		status.Status = "active"
		status.TimeRemaining = auctionItem.ExpiryTime.Sub(time.Now()).String()
	}

	// Include highest bid if available
	if err == nil {
		status.HighestBid = &highestBid
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(status)
}

// GetBidHistory handles requests to get the bid history for an auction
func (s *Server) GetBidHistory(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	auctionID := vars["id"]

	bids, err := s.Store.GetBidHistory(auctionID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(bids)
}
