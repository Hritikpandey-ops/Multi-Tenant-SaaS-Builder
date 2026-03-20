# Multi-stage Dockerfile for Go services

# Build stage
FROM golang:1.21-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git make

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod go.work* ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the services
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o bin/gateway ./cmd/gateway
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o bin/auth ./cmd/auth

# Runtime stage
FROM alpine:latest

# Install ca-certificates for HTTPS
RUN apk --no-cache add ca-certificates curl

WORKDIR /root/

# Copy binaries from builder
COPY --from=builder /app/bin/ /app/bin/

# Copy migration files
COPY --from=builder /app/db/ /app/db/

# Expose ports
EXPOSE 3000 3001

# Run the gateway by default
CMD ["/app/bin/gateway"]
