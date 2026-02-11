# Product Inventory CLI

A production-grade CLI application for managing a product inventory system in Go.

## Features

*   **CRUD Operations**: Create, Read, Update, Delete products.
*   **Storage Backends**:
    *   In-Memory (default, thread-safe)
    *   JSON File Persistence
*   **Advanced Features**:
    *   Concurrent Bulk Import
    *   Export to JSON
    *   Filtering and Sorting
*   **Configuration**: Supports config file, environment variables, and flags (Viper).
*   **Observability**: Structured logging (slog).
*   **Dockerized**: Multi-stage Dockerfile included.

## Prerequisites

*   Go 1.21 or higher
*   Docker (optional, for containerized execution)

## Build

```bash
go build -o inventory-cli ./cmd/inventory-cli
```

## Usage

### Global Flags

*   `--config`: Config file (default is $HOME/.inventory-cli.yaml)
*   `--store`: Storage type (`memory` or `json`) (default "memory")
*   `--db-file`: File path for json store (default "products.json")
*   `--log-level`: Log level (`debug`, `info`, `warn`, `error`) (default "info")

### Commands

#### Create a Product
```bash
./inventory-cli create --name "Laptop" --price 999.99 --quantity 10 --category "Electronics"
```

#### List Products
```bash
./inventory-cli list --category "Electronics" --min-price 500
./inventory-cli list --output json
```

#### Get a Product
```bash
./inventory-cli get <product-id>
```

#### Update a Product
```bash
./inventory-cli update <product-id> --price 899.99
```

#### Delete a Product
```bash
./inventory-cli delete <product-id>
```

#### Import Products
```bash
./inventory-cli import --file data.json
```

#### Export Products
```bash
./inventory-cli export --file backup.json --category "Electronics"
```

## Testing

Run unit tests:
```bash
go test -v ./...
# For race detection (requires CGO):
# go test -v -race ./...
```

## Docker

Build the image:
```bash
docker build -t inventory-cli:latest .
```

Run container:
```bash
docker run --rm inventory-cli:latest list
```

Persist data with JSON store:
```bash
docker run --rm -v $(pwd)/data:/data inventory-cli:latest create \
  --name "Docker Product" --price 10.0 --quantity 1 \
  --store json --db-file /data/products.json
```

## Design Decisions

*   **Architecture**: Follows standard Go layout (`cmd`, `internal`, `pkg`). Domain logic is isolated in `internal/domain`.
*   **Dependency Injection**: `StoreFactory` creates specific store implementations satisfying `ProductStore` interface.
*   **Concurrency**: `InMemoryStore` uses `sync.RWMutex` for thread safety. Bulk import uses worker pool pattern.
*   **Configuration**: Viper handles configuration precedence (Flag > Env > Config File > Default).
