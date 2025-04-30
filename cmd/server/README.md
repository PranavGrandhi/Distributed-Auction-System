# Server

## API Specifications

This document outlines the API specifications for the auction system web server.

### Auctions

#### Create Auction
- **Method**: POST
- **Endpoint**: `/auctions`
- **Auth**: Required
- **Request Body**:
  ```json
  {
    "name": "string",
    "description": "string",
    "minimum_bid": "number",
    "expiry_time": "timestamp"
  }
  ```
- **Response**:
  ```json
  {
    "id": "string",
    "name": "string",
    "description": "string",
    "minimum_bid": "number",
    "expiry_time": "timestamp",
    "created_at": "timestamp",
  }
  ```
- **Status Codes**:
  - `201 Created`: Auction created
  - `400 Bad Request`: Invalid input
  - `401 Unauthorized`: Not authenticated

#### List All Auctions
- **Method**: GET
- **Endpoint**: `/auctions`
- **Response**:
  ```json
  [
    {
      "id": "string",
      "name": "string",
      "description": "string",
      "minimum_bid": "number",
      "expiry_time": "timestamp",
      "created_at": "timestamp",
    }
  ]
  ```
- **Status Codes**:
  - `200 OK`: Success
  - `500 Internal Server Error`: Server error

#### Get Auction Details
- **Method**: GET
- **Endpoint**: `/auctions/{id}/status`
- **Response**:
  ```json
  {
    "auction": {
      "id": "string",
      "name": "string",
      "description": "string",
      "minimum_bid": "number",
      "expiry_time": "timestamp",
      "created_at": "timestamp",
    },
    "highest_bid": {
      "id": "string",
      "participant_id": "string",
      "auction_item_id": "string",
      "bid_price": "number",
      "timestamp": "timestamp"
    },
    "status": "string",
    "time_remaining": "string"
  }
  ```
- **Status Codes**:
  - `200 OK`: Success
  - `404 Not Found`: Auction not found

### Bidding

#### Place Bid
- **Method**: POST
- **Endpoint**: `/auctions/{id}/bids`
- **Auth**: Required
- **Request Body**:
  ```json
  {
    "participant_id": "string",
    "bid_price": "number",
    "auction_item_id": "string",
    "timestamp": "timestamp",
  }
  ```
- **Response**:
  ```json
  {
    "message": "Bid placed successfully"
  }
  ```
- **Status Codes**:
  - `201 Created`: Bid placed
  - `400 Bad Request`: Invalid bid (too low)
  - `401 Unauthorized`: Not authenticated
  - `404 Not Found`: Auction not found

#### Get Bid History
- **Method**: GET
- **Endpoint**: `/auctions/{id}/history`
- **Response**:
  ```json
  [
    {
      "auction_item_id": "string",
      "participant_id": "string",
      "bid_price": "number",
      "timestamp": "timestamp"
    }
  ]
  ```
- **Status Codes**:
  - `200 OK`: Success
  - `404 Not Found`: Auction not found

## Error Responses

All API endpoints return errors in the following format:

```json
{
  "error": {
    "code": "string",
    "message": "string",
    "details": {}
  }
}
```

## Static Content

The server also serves static content:

- `/` - Serves the frontend application
- `/frontend/*` - Serves static frontend files
