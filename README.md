# Labs Auction

## Project Overview

Labs Auction is an HTTP auction service written in Go. It allows clients to create and query auctions, place bids, retrieve bids and winners, and find users. Auctions are automatically closed in the background after a configurable duration without blocking the request that created them.

## Architecture

```mermaid
flowchart LR
    Client --> Gin[HTTP API - Gin]
    Gin --> UseCases[Use cases]
    UseCases --> Entities[Domain entities]
    UseCases --> Repositories[MongoDB repositories]
    Repositories --> MongoDB[(MongoDB)]
    Repositories -->|Goroutine after AUCTION_DURATION| Closure[Automatic auction closure]
    Closure --> MongoDB
```

The application follows a layered architecture with domain entities, use cases, HTTP controllers, and MongoDB repository implementations.

## Technologies

- Go 1.20
- Gin
- MongoDB
- MongoDB Go Driver
- Zap
- Docker and Docker Compose

## Requirements

- Go 1.20+ for local execution
- Docker and Docker Compose for containerized execution
- Ports `8080` and `27017` available

## Environment Variables

The project reads its configuration from `cmd/auction/.env`.

| Variable | Example | Description |
| --- | --- | --- |
| `BATCH_INSERT_INTERVAL` | `20s` | Maximum time before a pending batch of bids is written to MongoDB |
| `MAX_BATCH_SIZE` | `4` | Maximum number of bids in one batch |
| `AUCTION_INTERVAL` | `5s` | Time window used to reject bids after an auction expires |
| `AUCTION_DURATION` | `5s` | Time after which a newly created auction is automatically marked as completed |
| `MONGO_INITDB_ROOT_USERNAME` | `admin` | MongoDB root username |
| `MONGO_INITDB_ROOT_PASSWORD` | `admin` | MongoDB root password |
| `MONGODB_URL` | `mongodb://admin:admin@mongodb:27017/auctions?authSource=admin` | MongoDB connection URL |
| `MONGODB_DB` | `auctions` | MongoDB database name |

Go duration values accept units such as `ms`, `s`, and `m`. For example, use `AUCTION_DURATION=30s` to close auctions after 30 seconds or `AUCTION_DURATION=5m` to close them after five minutes. Keep `AUCTION_DURATION` and `AUCTION_INTERVAL` equal so automatic closure and bid acceptance use the same expiration window. If `AUCTION_DURATION` is absent or invalid, the application uses five minutes.

## How to Run Locally

Start MongoDB:

```bash
docker compose up -d mongodb
```

For an application running directly on the host, change the MongoDB hostname in `cmd/auction/.env` from `mongodb` to `localhost`:

```env
MONGODB_URL=mongodb://admin:admin@localhost:27017/auctions?authSource=admin
```

Then start the API from the repository root:

```bash
go run ./cmd/auction
```

Create an auction:

```bash
curl -X POST http://localhost:8080/auction \
  -H "Content-Type: application/json" \
  -d '{"product_name":"Notebook","category":"Electronics","description":"Notebook in excellent condition","condition":1}'
```

The request returns HTTP `201`. The auction starts with status `0` (`Active`) and is automatically changed to status `1` (`Completed`) after `AUCTION_DURATION`.

## How to Execute Tests

Run the automated tests from the repository root:

```bash
go test ./internal/infra/database/auction
```

The automatic-closure test uses the MongoDB driver's mock deployment and therefore does not require a running MongoDB instance.

## Docker

Build and start the API and MongoDB in the background:

```bash
docker compose up --build -d
```

Check container status:

```bash
docker compose ps
```

Follow application logs:

```bash
docker compose logs -f app
```

Stop the stack:

```bash
docker compose down
```

To also remove the MongoDB volume:

```bash
docker compose down -v
```

## Project Structure

```text
.
|-- cmd/auction                  # Application entry point and environment file
|-- configuration               # MongoDB, logging, and REST error configuration
|-- internal
|   |-- entity                  # Auction, bid, and user domain entities
|   |-- infra
|   |   |-- api/web             # Gin controllers and request validation
|   |   `-- database            # MongoDB repositories and auction closure routine
|   |-- internal_error          # Application error model
|   `-- usecase                 # Application use cases and DTOs
|-- Dockerfile
|-- docker-compose.yml
|-- go.mod
`-- README.md
```
