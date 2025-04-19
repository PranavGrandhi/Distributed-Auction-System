package main

import (
	"log"
	"net/http"
	"os"

	"github.com/PranavGrandhi/Distributed-Auction-System/pkg/api"
)

func main() {
	// Get port from environment variable or use default
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	server := api.NewServer()

	log.Printf("Starting auction server on port %s...", port)
	log.Fatal(http.ListenAndServe(":"+port, server.Router))
}
