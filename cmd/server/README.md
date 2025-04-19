# Server

## API Specifications

This document outlines the API specifications for the auction system web server.

### Auctions

#### Create Auction
- **Method**: POST
- **Endpoint**: `/api/auctions`
- **Auth**: Required
- **Request Body**:
  ```json
  {
    "title": "string",
    "description": "string",
    "starting_price": "number",
    "reserve_price": "number",
    "end_time": "timestamp",
    "images": ["string"]
  }
  ```
- **Response**:
  ```json
  {
    "id": "string",
    "title": "string",
    "description": "string",
    "starting_price": "number",
    "current_price": "number",
    "reserve_price": "number",
    "seller_id": "string",
    "created_at": "timestamp",
    "end_time": "timestamp",
    "status": "string",
    "images": ["string"]
  }
  ```
- **Status Codes**:
  - `201 Created`: Auction created
  - `400 Bad Request`: Invalid input
  - `401 Unauthorized`: Not authenticated

#### Query Auction Status
- **Method**: GET
- **Endpoint**: `/api/auctions/{id}`
- **Response**: Auction object with current status
  ```json
  {
    "id": "string",
    "title": "string",
    "description": "string",
    "starting_price": "number",
    "current_price": "number",
    "reserve_price": "number",
    "seller_id": "string",
    "created_at": "timestamp",
    "end_time": "timestamp",
    "status": "string",
    "bids_count": "number",
    "images": ["string"]
  }
  ```
- **Status Codes**:
  - `200 OK`: Success
  - `404 Not Found`: Auction not found

### Bidding

#### Place Bid
- **Method**: POST
- **Endpoint**: `/api/auctions/{id}/bids`
- **Auth**: Required
- **Request Body**:
  ```json
  {
    "amount": "number"
  }
  ```
- **Response**:
  ```json
  {
    "id": "string",
    "auction_id": "string",
    "bidder_id": "string",
    "amount": "number",
    "created_at": "timestamp",
    "status": "string"
  }
  ```
- **Status Codes**:
  - `201 Created`: Bid placed
  - `400 Bad Request`: Invalid bid (too low)
  - `401 Unauthorized`: Not authenticated
  - `403 Forbidden`: Cannot bid (e.g., own auction)
  - `404 Not Found`: Auction not found
  - `409 Conflict`: Auction ended

#### Get Bid History
- **Method**: GET
- **Endpoint**: `/api/auctions/{id}/bids`
- **Query Parameters**:
  - `page`: Page number
  - `limit`: Items per page
- **Response**:
  ```json
  {
    "bids": [
      {
        "id": "string",
        "bidder_username": "string",
        "amount": "number",
        "created_at": "timestamp"
      }
    ],
    "total": "number",
    "page": "number",
    "limit": "number"
  }
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
