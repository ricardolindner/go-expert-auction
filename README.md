# go-expert-auction

This project is part of the "Go Expert" course and implements an auction system with features for creating auctions and bids, along with automatic auction closing functionality.

---

## Table of Contents
- [Project Structure](#project-structure)
- [How It Works](#how-it-works)
- [Getting Started](#getting-started)
- [Configuration](#configuration)
  - [Environment Variables (.env)](#environment-variables-env)
  - [Example .env File](#example-env-file)
- [Running the Project](#running-the-project)
- [How to Test](#how-to-test)
---

## Project Structure

```text
go-expert-auction/
|-- cmd/
|   |-- auction/
|   |   |-- main.go
|   |   |-- .env
|-- configuration/
|   |-- database/mongodb/connection.go
|   |-- logger/logger.go
|   |-- rest_err/rest_err.go
|-- internal/
|   |-- entity/
|   |   |-- auction_entity/auction_entity.go
|   |   |-- bid_entity/bid_entity.go
|   |   |-- user_entity/user_entity.go
|   |-- infra/
|   |   |-- api/web/controller/
|   |   |   |-- auction_controller/
|   |   |   |   |-- create_auction_controller.go
|   |   |   |   |-- find_auction_controller.go
|   |   |   |-- bid_controller/
|   |   |   |   |-- create_bid_controller.go
|   |   |   |   |-- find_bid_controller.go
|   |   |   |-- user_controller/
|   |   |   |   |-- find_user_controller.go
|   |   |-- validation/validation.go
|   |   |-- database/
|   |   |   |-- auction/
|   |   |   |   |-- create_auction.go
|   |   |   |   |-- find_auction.go
|   |   |   |   |-- create_auction_test.go
|   |   |   |-- bid/
|   |   |   |   |-- create_bid.go
|   |   |   |   |-- find_bid.go
|   |   |   |-- user/
|   |   |   |   |-- find_user.go
|   |-- internal_error/internal_error.go
|   |-- usecase/
|   |   |-- auction_usecase/
|   |   |   |-- create_auction_usecase.go
|   |   |   |-- find_auction_usecase.go
|   |   |-- bid_usecase/
|   |   |   |-- create_bid_usecase.go
|   |   |   |-- find_bid_usecase.go
|   |   |-- user_usecase/
|   |   |   |-- find_user_usecase.go
|-- .gitignore
|-- docker-compose.yml
|-- go.mod
|-- go.sum
|-- README.md
```

## How It Works

The project follows a standard layered architecture. The automatic auction closing feature is implemented using a goroutine that runs in the background when a new auction is created. It waits for the duration specified in `AUCTION_INTERVAL` and updates the auction's status to `Completed`.

The project uses **Docker Compose** to orchestrate both the application and a MongoDB database, making setup and execution straightforward.

## Getting Started
Prerequisites
-   Go (version 1.20 or higher)
-   Docker
-   **Docker Compose V2** (version 2.x.x or higher)

Clone the repository
```bash
git clone https://github.com/ricardolindner/go-expert-auction.git
cd go-expert-auction
```

## Configuration
All configuration is managed via environment variables. For local development, create a `.env` file in the `cmd/auction` directory.

### Environment Variables (`.env`)
* `PBATCH_INSERT_INTERVAL`: Interval for batch-inserting bids into the database.
* `MAX_BATCH_SIZE`: Maximum number of bids per batch insert.
* `AUCTION_INTERVAL`: Duration after which an auction automatically closes (e.g., `20s`, `5m`, `1h`).

* `MONGO_INITDB_ROOT_USERNAME`: The root username for MongoDB container.
* `MONGO_INITDB_ROOT_PASSWORD`: The root password for MongoDB container.
* `MONGODB_URL`: Connection string to MongoDB.
* `MONGODB_DB`: Database name.

### Example `.env` File
```.env
PBATCH_INSERT_INTERVAL=20s
MAX_BATCH_SIZE=4
AUCTION_INTERVAL=20s

MONGO_INITDB_ROOT_USERNAME=admin
MONGO_INITDB_ROOT_PASSWORD=admin
MONGODB_URL=mongodb://admin:admin@mongodb:27017/auctions?authSource=admin
MONGODB_DB=auctions
```

## Running the Project

### 1.Start the Containers:
```bash
docker compose up --build
```
This will build and start both the application and the MongoDB database.

## How to Test

### 1. Create an Auction
Send a `POST` request to create a new auction. For testing automatic closing, set `AUCTION_INTERVAL` to a short duration (e.g., 10s).

```bash
curl -X POST http://localhost:8080/auction \
-H 'Content-Type: application/json' \
-d '{
  "product_name": "Vintage Watch",
  "category": "Accessories",
  "description": "A beautiful vintage watch.",
  "condition": 1
}'
```

### 2. Verify Automatic Closing
After the auction duration, send a `GET` request to verify the auction status:
```bash
curl -X GET http://localhost:8080/auction/{{YOUR_AUCTION_ID}}
```
Expected response:
```json
{
  "id":"{{YOUR_AUCTION_ID}}",
  "product_name":"Vintage Watch",
  "category":"Accessories",
  "description":"A beautiful vintage watch.",
  "condition":1,
  "status":1,
  "timestamp":"2025-08-24T00:49:20Z"
}
```

### 3. Run Automated Tests
The project includes a unit test that validates automatic auction closing.

In the project root run:
```bash
go test ./internal/infra/database/auction/ -v
```
