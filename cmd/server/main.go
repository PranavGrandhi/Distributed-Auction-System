package main

import (
	"flag"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/PranavGrandhi/Distributed-Auction-System/pkg/api"
)

func main() {
	// Command line flags
	zkHosts := flag.String("zk", "localhost:2181,localhost:2182,localhost:2183", "ZooKeeper hosts, comma separated")
	port := flag.String("port", "", "HTTP server port")
	useZK := flag.Bool("use-zk", false, "Use ZooKeeper for distributed storage")
	flag.Parse()

	// Get port from environment variable or flag or use default
	if *port == "" {
		*port = os.Getenv("PORT")
		if *port == "" {
			*port = "8080"
		}
	}

	var server *api.Server
	var err error

	if *useZK {
		// Using ZooKeeper
		zkHostsList := strings.Split(*zkHosts, ",")
		server, err = api.NewZooKeeperServer(zkHostsList)
		if err != nil {
			log.Fatalf("Failed to create ZooKeeper server: %v", err)
		}
		log.Printf("Starting distributed auction server with ZooKeeper on port %s...", *port)
	} else {
		// Using memory storage (for backward compatibility)
		server = api.NewServer()
		log.Printf("Starting standalone auction server on port %s...", *port)
	}

	log.Fatal(http.ListenAndServe(":"+*port, server.Router))
}
