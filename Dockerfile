# Build Stage
FROM golang:1.21-alpine AS builder

WORKDIR /app

# Install dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build statically linked binary
RUN CGO_ENABLED=0 GOOS=linux go build -o inventory-cli ./cmd/inventory-cli

# Runtime Stage
FROM alpine:latest

WORKDIR /app

# Copy binary from builder
COPY --from=builder /app/inventory-cli .

# Create non-root user
RUN adduser -D nonroot
USER nonroot

# Set environment variables
ENV STORE_TYPE=memory
ENV LOG_LEVEL=info

# Default command (can be overridden)
ENTRYPOINT ["./inventory-cli"]
