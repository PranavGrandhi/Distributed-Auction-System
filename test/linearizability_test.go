package test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"sort"
	"sync"
	"testing"
	"time"

	"github.com/PranavGrandhi/Distributed-Auction-System/pkg/auction"
)

// Server URLs for the three different servers
var serverURLs = []string{
	"http://localhost:8080",
	"http://localhost:8081",
	"http://localhost:8082",
}

// Helper function to get bid history for an auction
func getBidHistory(t *testing.T, serverURL, auctionID string) ([]auction.Bid, error) {
	// Create a unique request to prevent caching
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/auctions/%s/history", serverURL, auctionID), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	// Add a unique header to prevent request caching
	req.Header.Add("Cache-Control", "no-cache")
	req.Header.Add("X-Test-ID", fmt.Sprintf("%d", time.Now().UnixNano()))

	// Execute the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get bid history: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("unexpected status code: %d, body: %s", resp.StatusCode, string(bodyBytes))
	}

	var bids []auction.Bid
	if err := json.NewDecoder(resp.Body).Decode(&bids); err != nil {
		return nil, fmt.Errorf("failed to decode response: %v", err)
	}

	// Sort bids by timestamp for consistent comparison
	sort.Slice(bids, func(i, j int) bool {
		return bids[i].Timestamp.Before(bids[j].Timestamp)
	})

	return bids, nil
}

// Helper function to get a random server URL
func getRandomServerURL() string {
	randomIdx := rand.Intn(len(serverURLs))
	return serverURLs[randomIdx]
}

// Helper function to get all auctions
func getAuctions(t *testing.T, serverURL string) (*AuctionList, error) {
	// Create a unique request to prevent caching
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/auctions", serverURL), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	// Add a unique header to prevent request caching
	req.Header.Add("Cache-Control", "no-cache")
	req.Header.Add("X-Test-ID", fmt.Sprintf("%d", time.Now().UnixNano()))

	// Execute the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get auctions: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("unexpected status code: %d, body: %s", resp.StatusCode, string(bodyBytes))
	}

	// Read the response body
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %v", err)
	}

	// Direct list of auctions
	var directAuctions []auction.AuctionItem
	if err := json.Unmarshal(bodyBytes, &directAuctions); err == nil {
		t.Log("Direct auction list parsed successfully")
		return &AuctionList{Auctions: directAuctions}, nil
	}

	return nil, fmt.Errorf("failed to parse response as auction data: %s", string(bodyBytes))
}

// Helper function to get a specific auction
func getAuctionStatus(t *testing.T, serverURL, auctionID string) (*AuctionStatus, error) {
	// Create a unique request to prevent caching
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/auctions/%s/status", serverURL, auctionID), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	// Add a unique header to prevent request caching
	req.Header.Add("Cache-Control", "no-cache")
	req.Header.Add("X-Test-ID", fmt.Sprintf("%d", time.Now().UnixNano()))

	// Execute the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get auction: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("unexpected status code: %d, body: %s", resp.StatusCode, string(bodyBytes))
	}

	var auctionItem AuctionStatus
	if err := json.NewDecoder(resp.Body).Decode(&auctionItem); err != nil {
		return nil, fmt.Errorf("failed to decode response: %v", err)
	}

	return &auctionItem, nil
}

// Helper function to place a bid
func placeBid(t *testing.T, serverURL, auctionID string, amount float64) (*auction.Bid, error) {
	bidRequest := PlaceBidRequest{
		BidPrice:      amount,
		AuctionItemId: auctionID,
		ParticipantId: fmt.Sprintf("test-participant-%d", time.Now().UnixNano()),
		Timestamp:     time.Now(),
	}

	requestBody, err := json.Marshal(bidRequest)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal bid request: %v", err)
	}

	// Create a unique request to prevent caching
	req, err := http.NewRequest("POST",
		fmt.Sprintf("%s/auctions/%s/bids", serverURL, auctionID),
		bytes.NewBuffer(requestBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	// Add content-type and cache-busting headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Add("Cache-Control", "no-cache")
	req.Header.Add("X-Test-ID", fmt.Sprintf("%d", time.Now().UnixNano()))

	// Execute the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to place bid: %v", err)
	}
	defer resp.Body.Close()

	// Read response body for error messages
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %v", err)
	}

	if resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("bid failed with status %d: %s", resp.StatusCode, string(body))
	}

	var bid auction.Bid
	if err := json.Unmarshal(body, &bid); err != nil {
		// If we can't parse the response as a Bid, it might be a success message
		return nil, nil
	}

	return &bid, nil
}

// Also need to add these missing struct types
type AuctionList struct {
	Auctions []auction.AuctionItem `json:"auctions"`
}

type PlaceBidRequest struct {
	BidPrice      float64   `json:"bid_price"`
	AuctionItemId string    `json:"auction_item_id"`
	ParticipantId string    `json:"participant_id"`
	Timestamp     time.Time `json:"timestamp"`
}

type AuctionStatus struct {
	Auction       auction.AuctionItem `json:"auction"`
	HighestBid    *auction.Bid        `json:"highest_bid,omitempty"`
	Status        string              `json:"status"`
	TimeRemaining string              `json:"time_remaining,omitempty"`
}

// TestLinearizabilityAcrossServers tests that all servers maintain consistent bid histories
// with concurrent clients placing bids
func TestLinearizabilityAcrossServers(t *testing.T) {
	// Generate a unique test ID to prevent test caching
	testID := time.Now().UnixNano()
	rand.Seed(testID)
	t.Logf("Running linearizability test with unique ID: %d", testID)

	// Add a completely random delay to ensure tests don't run in a predictable pattern
	randomDelay := time.Duration(rand.Intn(1000)) * time.Millisecond
	time.Sleep(randomDelay)
	t.Logf("Used random delay of %v", randomDelay)

	t.Log("Starting linearizability test...")

	// First, get all auction items from a random server
	randomServerURL := getRandomServerURL()
	t.Logf("Using server at %s for initial auctions list", randomServerURL)

	auctions, err := getAuctions(t, randomServerURL)
	if err != nil {
		t.Fatalf("Failed to get auctions: %v", err)
	}

	if len(auctions.Auctions) == 0 {
		t.Fatalf("No auctions available for testing")
	}

	t.Logf("Found %d auctions for testing", len(auctions.Auctions))

	// Create a map to track which auctions were used
	usedAuctions := make(map[string]bool)
	var usedAuctionsMutex sync.Mutex

	// Create a wait group for our goroutines
	var wg sync.WaitGroup
	wg.Add(len(serverURLs))

	// Create 3 goroutines, one for each server
	for i, serverURL := range serverURLs {
		go func(serverIndex int, url string) {
			defer wg.Done()
			clientID := fmt.Sprintf("client-%d-%d", serverIndex, testID)
			t.Logf("Starting client %s connecting to %s", clientID, url)

			// Perform 10 bids for each client
			for j := 0; j < 10; j++ {
				// Generate a small random delay to increase concurrency variability
				time.Sleep(time.Duration(rand.Intn(500)) * time.Millisecond)

				// Pick a random auction
				randomIdx := rand.Intn(len(auctions.Auctions))
				selectedAuction := auctions.Auctions[randomIdx]

				// Lock to update the used auctions map
				usedAuctionsMutex.Lock()
				usedAuctions[selectedAuction.ID] = true
				usedAuctionsMutex.Unlock()

				// Get current highest bid
				auctionStatus, err := getAuctionStatus(t, url, selectedAuction.ID)
				if err != nil {
					t.Logf("Client %s: Failed to get auction status: %v", clientID, err)
					continue
				}

				// Determine bid amount
				var currentPrice float64
				if auctionStatus.HighestBid != nil {
					currentPrice = auctionStatus.HighestBid.BidPrice
				} else {
					currentPrice = auctionStatus.Auction.MinimumBid
				}

				// Add a random amount (between 1 and 10) to current price
				newBidAmount := currentPrice + float64(1+rand.Intn(10))

				// Place bid
				participantID := fmt.Sprintf("%s-iter-%d", clientID, j)
				_, err = placeBid(t, url, selectedAuction.ID, newBidAmount)
				if err != nil {
					t.Logf("Client %s: Failed to place bid: %v", participantID, err)
					continue
				}
				t.Logf("Client %s: Placed bid of %.2f on auction %s",
					participantID, newBidAmount, selectedAuction.ID)
			}
			t.Logf("Client on server %s completed all bids", url)
		}(i, serverURL)
	}

	// Wait for all goroutines to finish
	t.Log("Waiting for all clients to complete their bids...")
	wg.Wait()
	t.Log("All bidding clients have completed")

	// Wait for 20 seconds to allow replication to complete
	t.Log("Waiting 20 seconds for replication to complete across servers...")
	time.Sleep(20 * time.Second)

	// Extract used auction IDs
	var usedAuctionIDs []string
	for id := range usedAuctions {
		usedAuctionIDs = append(usedAuctionIDs, id)
	}

	// Check bid histories from all servers for each auction
	var testFailed bool
	for _, auctionID := range usedAuctionIDs {
		t.Logf("Checking bid history consistency for auction %s", auctionID)

		// Get bid history from first server
		firstServerBids, err := getBidHistory(t, serverURLs[0], auctionID)
		if err != nil {
			t.Errorf("Failed to get bid history from server %s: %v", serverURLs[0], err)
			testFailed = true
			continue
		}

		// Compare with other servers
		for i := 1; i < len(serverURLs); i++ {
			otherServerBids, err := getBidHistory(t, serverURLs[i], auctionID)
			if err != nil {
				t.Errorf("Failed to get bid history from server %s: %v", serverURLs[i], err)
				testFailed = true
				continue
			}

			// Format the bid histories for logging
			firstServerInfo := formatBidHistoryForLogging(firstServerBids)
			otherServerInfo := formatBidHistoryForLogging(otherServerBids)

			// Check if the number of bids match
			if len(firstServerBids) != len(otherServerBids) {
				t.Errorf("Linearizability violation: Server %s has %d bids while server %s has %d bids for auction %s",
					serverURLs[0], len(firstServerBids), serverURLs[i], len(otherServerBids), auctionID)
				t.Errorf("Server %s bids: %s", serverURLs[0], firstServerInfo)
				t.Errorf("Server %s bids: %s", serverURLs[i], otherServerInfo)
				testFailed = true
				continue
			}

			// Compare each bid in detail
			for j := 0; j < len(firstServerBids); j++ {
				bid1 := firstServerBids[j]
				bid2 := otherServerBids[j]

				// Compare bid prices, which must match exactly
				if bid1.BidPrice != bid2.BidPrice {
					t.Errorf("Linearizability violation: Bid prices don't match between servers for auction %s", auctionID)
					t.Errorf("Server %s bids: %s", serverURLs[0], firstServerInfo)
					t.Errorf("Server %s bids: %s", serverURLs[i], otherServerInfo)
					testFailed = true
					break
				}

				// Compare participant IDs, which must match exactly
				if bid1.ParticipantID != bid2.ParticipantID {
					t.Errorf("Linearizability violation: Participant IDs don't match between servers for auction %s", auctionID)
					t.Errorf("Server %s bids: %s", serverURLs[0], firstServerInfo)
					t.Errorf("Server %s bids: %s", serverURLs[i], otherServerInfo)
					testFailed = true
					break
				}
			}
		}
	}

	if testFailed {
		t.Fatal("❌ TEST FAILED: Linearizability violation detected between servers")
	} else {
		t.Log("✅ TEST PASSED: Bid histories are consistent across all servers (linearizability maintained)")
	}
}

// Helper function to format bid history for logging
func formatBidHistoryForLogging(bids []auction.Bid) string {
	result := ""
	for i, bid := range bids {
		result += fmt.Sprintf("[%d] %.2f by %s at %s",
			i, bid.BidPrice, bid.ParticipantID, bid.Timestamp.Format(time.RFC3339))
		if i < len(bids)-1 {
			result += ", "
		}
	}
	return result
}
