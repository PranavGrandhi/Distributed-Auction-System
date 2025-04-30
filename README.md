# Distributed Auction System

A fault-tolerant, distributed auction system built with Go and Apache ZooKeeper for coordination between multiple server instances.

## Overview

This project implements a distributed auction platform that allows users to:
- Create auctions with descriptions, minimum bids, and expiry times
- Place bids on active auctions
- View auction status and bid histories
- Access the system through multiple server instances

The system ensures consistency across multiple servers using ZooKeeper for distributed coordination and locking, making it resilient to individual server failures.

## Features

- **Distributed Architecture**: Multiple auction server instances sharing the same state
- **Fault Tolerance**: Uses a ZooKeeper ensemble for coordination, maintaining functionality even if individual servers fail
- **Distributed Locking**: Ensures bid consistency and prevents race conditions
- **User-Friendly Interface**: Simple web UI for interacting with the auction system
- **Real-Time Updates**: Auction status updated across all servers in near real-time

## Architecture

The system follows a tiered architecture:

- **Frontend**: HTML/CSS/JavaScript web interface
- **Backend API Server**: Go-based REST API endpoints
- **Storage Layer**: Two implementations:
  - `MemoryStore`: In-memory storage for single-server deployment
  - `ZKStore`: ZooKeeper-backed distributed storage

## Project Structure

```
DISTRIBUTED-AUCTION-SYSTEM/
├── cmd/
│   ├── client/       # Client application
│   │   └── main.go
│   └── server/       # Server application
│       └── main.go
├── frontend/
│   └── index.html    # Web UI
├── internal/
│   └── state/        # Internal state management
├── pkg/
│   ├── api/          # API handlers
│   │   └── handlers.go
│   ├── auction/      # Auction models
│   │   └── models.go
│   ├── consensus/    # Consensus mechanisms
│   └── storage/      # Storage implementations
│       ├── memory.go # In-memory storage
│       ├── store.go  # Storage interface
│       └── zkstore.go # ZooKeeper storage
├── test/             # Test suite
├── go.mod            # Go module definition
└── go.sum            # Go module checksums
```

## Requirements

- Go 1.16+
- Apache ZooKeeper 3.7+ (for distributed mode)

## Getting Started

### Single-Server Mode

1. Clone the repository:
   ```bash
   git clone https://github.com/PranavGrandhi/Distributed-Auction-System
   cd Distributed-Auction-System
   ```

2. Build and run the server:
   ```bash
   go run cmd/server/main.go
   ```

3. Access the web interface at http://localhost:8080

### Distributed Mode

#### Setting Up ZooKeeper with Docker

1. Create a file named `docker-compose.yml` with the following content:

```yaml
version: '3.1'

services:
  zoo1:
    image: zookeeper
    restart: always
    hostname: zoo1
    ports:
      - 2181:2181
    environment:
      ZOO_MY_ID: 1
      ZOO_SERVERS: server.1=zoo1:2888:3888;2181 server.2=zoo2:2888:3888;2181 server.3=zoo3:2888:3888;2181

  zoo2:
    image: zookeeper
    restart: always
    hostname: zoo2
    ports:
      - 2182:2181
    environment:
      ZOO_MY_ID: 2
      ZOO_SERVERS: server.1=zoo1:2888:3888;2181 server.2=zoo2:2888:3888;2181 server.3=zoo3:2888:3888;2181

  zoo3:
    image: zookeeper
    restart: always
    hostname: zoo3
    ports:
      - 2183:2181
    environment:
      ZOO_MY_ID: 3
      ZOO_SERVERS: server.1=zoo1:2888:3888;2181 server.2=zoo2:2888:3888;2181 server.3=zoo3:2888:3888;2181
```

2. Start the ZooKeeper ensemble:
   ```bash
   docker-compose up -d
   ```

3. Initialize the required ZooKeeper paths:
   ```bash
   # Connect to ZooKeeper CLI
   docker exec -it zookeeper-docker-zoo1-1 zkCli.sh
   
   # Create required paths
   ls /
   create /auction-system ""
   create /auction-system/auctions ""
   create /auction-system/bids ""
   create /auction-system/locks ""
   create /auction-system/services ""
   
   # Verify the paths
   ls /auction-system
   
   # Exit the CLI
   quit
   ```

4. Start multiple auction server instances:
   ```bash
   # Terminal 1
   go run cmd/server/main.go --port=8080 --use-zk=true --zk=localhost:2181,localhost:2182,localhost:2183
   
   # Terminal 2
   go run cmd/server/main.go --port=8081 --use-zk=true --zk=localhost:2181,localhost:2182,localhost:2183
   
   # Terminal 3
   go run cmd/server/main.go --port=8082 --use-zk=true --zk=localhost:2181,localhost:2182,localhost:2183
   ```

3. Access any server's web interface:
   - Server 1: http://localhost:8080
   - Server 2: http://localhost:8081
   - Server 3: http://localhost:8082

## API Endpoints

- `GET /auctions` - List all auctions
- `POST /auctions` - Create a new auction
- `GET /auctions/{id}` - Get auction details
- `POST /auctions/{id}/bids` - Place a bid on an auction
- `GET /auctions/{id}/status` - Get current auction status
- `GET /auctions/{id}/history` - Get bid history for an auction

## Acknowledgments

- Apache ZooKeeper team for the distributed coordination service
- Go community for the excellent libraries and tools
- Prof. Aurojit Panda whose Distributed Systems course this was a final project for.
