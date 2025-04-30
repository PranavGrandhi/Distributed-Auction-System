package test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"os/exec"
	"testing"
	"time"

	"github.com/PranavGrandhi/Distributed-Auction-System/pkg/auction"
	"github.com/stretchr/testify/assert"
)

// AuctionList represents multiple auction items
type AuctionList struct {
	Auctions []auction.AuctionItem `json:"auctions"`
}

// PlaceBidRequest represents the request body for placing a bid
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

// ErrorResponse represents the API error response format
type ErrorResponse struct {
	Error struct {
		Code    string      `json:"code"`
		Message string      `json:"message"`
		Details interface{} `json:"details"`
	} `json:"error"`
}

// Server endpoints
var serverPorts = []string{"8080", "8081", "8082"}
var zookeepers = []string{"auction-zoo3-1", "auction-zoo2-1", "auction-zoo1-1"}

// Helper function to get a random server URL
func getRandomServerURL() string {
	randomIdx := rand.Intn(len(serverPorts))
	return fmt.Sprintf("http://localhost:%s", serverPorts[randomIdx])
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
        ParticipantId: "aditya", // Assuming participant ID is 1 for this test
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
        var errorResp ErrorResponse
        if err := json.Unmarshal(body, &errorResp); err == nil {
            return nil, fmt.Errorf("bid failed with status %d: %s", resp.StatusCode, errorResp.Error.Message)
        }
        return nil, fmt.Errorf("bid failed with status %d", resp.StatusCode)
    }

    var bid auction.Bid
    if err := json.Unmarshal(body, &bid); err != nil {
        return nil, fmt.Errorf("failed to decode response: %v", err)
    }

    return &bid, nil
}

// Helper function to kill a Zookeeper node
func killZookeeperNode(t *testing.T, nodeIndex int) error {
	if nodeIndex < 0 || nodeIndex >= len(zookeepers) {
		return fmt.Errorf("invalid zookeeper node index: %d", nodeIndex)
	}

	cmd := exec.Command("docker", "stop", zookeepers[nodeIndex])
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to stop zookeeper node: %v, output: %s", err, output)
	}

	t.Logf("Successfully stopped zookeeper node: %s", zookeepers[nodeIndex])
	return nil
}

// Helper function to restart a Zookeeper node
func restartZookeeperNode(t *testing.T, nodeIndex int) error {
	if nodeIndex < 0 || nodeIndex >= len(zookeepers) {
		return fmt.Errorf("invalid zookeeper node index: %d", nodeIndex)
	}

	cmd := exec.Command("docker", "start", zookeepers[nodeIndex])
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to restart zookeeper node: %v, output: %s", err, output)
	}

	t.Logf("Successfully restarted zookeeper node: %s", zookeepers[nodeIndex])
	return nil
}

// TestDataReplicationAfterLeaderFailure tests that data is replicated correctly
// even when the ZooKeeper leader node fails.
func TestDataReplicationAfterLeaderFailure(t *testing.T) {
	// More aggressive prevent test caching - this forces the test to run every time
    testID := time.Now().UnixNano()
    rand.Seed(testID) // Randomize behavior further
    t.Logf("Running test with unique ID: %d", testID)

    // Add a completely random delay to ensure tests don't run in a predictable pattern
    randomDelay := time.Duration(rand.Intn(1000)) * time.Millisecond
    time.Sleep(randomDelay)
    t.Logf("Used random delay of %v", randomDelay)

    t.Log("Starting data replication test...")

	// Get a random server to work with initially
	initialServerURL := getRandomServerURL()
	t.Logf("Using initial server at %s", initialServerURL)

	// 1. List all available auction items
	auctions, err := getAuctions(t, initialServerURL)
	if err != nil {
		t.Fatalf("Failed to get auctions: %v", err)
	}

	// 2. Pick a random auction item
	randomIdx := rand.Intn(len(auctions.Auctions))
	selectedAuction := auctions.Auctions[randomIdx]
	t.Logf("Selected auction: ID=%s, Name=%s",
		selectedAuction.ID, selectedAuction.Name)

	// Get detailed info about the selected auction
	auctionDetail, err := getAuctionStatus(t, initialServerURL, selectedAuction.ID)
	if err != nil {
		t.Fatalf("Failed to get auction details: %v", err)
	}

	// Calculate a new bid amount higher than the current minimum bid
	// Determine the current price from highest bid or minimum bid
	var currentPrice float64
	if auctionDetail.HighestBid != nil {
		currentPrice = auctionDetail.HighestBid.BidPrice
	} else {
		currentPrice = auctionDetail.Auction.MinimumBid
	}
	newBidAmount := currentPrice + 10.0 // Add $10 to current price
	t.Logf("Current minimum bid: %.2f, Placing new bid: %.2f", currentPrice, newBidAmount)

	// 3. Place a bid
	_, err = placeBid(t, initialServerURL, selectedAuction.ID, newBidAmount)
	if err != nil {
		t.Fatalf("Failed to place bid: %v", err)
	}

	// Rest of the test with ZooKeeper leader failure...
	// Note that you'll need to update field references:
	// - bid.Amount becomes bid.BidPrice
	// - updatedAuction.CurrentPrice becomes updatedAuction.HighestBid.BidPrice

	// 4. Kill the leader ZooKeeper node (we'll kill zoo1)
	leaderNodeIndex := 0 // Assuming zoo1 is the leader for this test
	t.Logf("Killing ZooKeeper node: %s", zookeepers[leaderNodeIndex])
	if err := killZookeeperNode(t, leaderNodeIndex); err != nil {
		t.Fatalf("Failed to kill ZooKeeper node: %v", err)
	}

	// Sleep to allow time for ZooKeeper to elect a new leader
	time.Sleep(10 * time.Second)
	t.Log("Waiting for ZooKeeper cluster to stabilize after leader failure...")

	// 5. Get auction details from a different server
	var differentServerURL string
	for _, port := range serverPorts {
		url := fmt.Sprintf("http://localhost:%s", port)
		if url != initialServerURL {
			differentServerURL = url
			break
		}
	}
	t.Logf("Using different server at %s to verify data replication", differentServerURL)

	// Retry logic for getting auction info
	var updatedAuction *AuctionStatus
	var getAuctionErr error
	maxRetries := 5
	for i := 0; i < maxRetries; i++ {
		time.Sleep(1 * time.Second)

		updatedAuction, getAuctionErr = getAuctionStatus(t, differentServerURL, selectedAuction.ID)
		// Check if the HighestBid was updated properly, instead of CurrentPrice
		if getAuctionErr == nil && updatedAuction.HighestBid.BidPrice == newBidAmount {
			break
		}

		t.Logf("Retry %d: Getting updated auction info...", i+1)
	}

	if getAuctionErr != nil {
		t.Fatalf("Failed to get updated auction details: %v", getAuctionErr)
	}

	// Restart the ZooKeeper node we stopped
	time.Sleep(10 * time.Second)
	defer func() {
		if err := restartZookeeperNode(t, leaderNodeIndex); err != nil {
			t.Logf("Warning: Failed to restart ZooKeeper node: %v", err)
		}
	}()

	// Assert that the bid was properly replicated
	// Use HighestBid instead of CurrentPrice
	assert.Equal(t, newBidAmount, updatedAuction.HighestBid.BidPrice,
		"The highest bid should be replicated across servers despite ZooKeeper leader failure")

	if updatedAuction.HighestBid.BidPrice == newBidAmount {
		t.Log("✅ TEST PASSED: Data was successfully replicated despite ZooKeeper leader failure")
	} else {
		t.Errorf("❌ TEST FAILED: Expected highest bid to be %.2f, but got %.2f",
			newBidAmount, updatedAuction.HighestBid.BidPrice)
	}
}
