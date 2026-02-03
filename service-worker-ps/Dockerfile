# Build stage
FROM golang:1.21-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git ca-certificates

# Set working directory
WORKDIR /build

# Copy go.mod and go.sum for dependency caching
COPY service-worker-ps/go.mod service-worker-ps/go.sum ./
RUN go mod download

# Copy source code
COPY service-worker-ps/ ./

# Build the binary with optimizations
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -ldflags="-w -s" \
    -o /service-worker-ps \
    ./cmd/service-worker-ps

# Final stage - minimal runtime image
FROM scratch

# Copy CA certificates for HTTPS connections
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# Copy the binary
COPY --from=builder /service-worker-ps /service-worker-ps

# Run the service worker
ENTRYPOINT ["/service-worker-ps"]
