# ===== Stage 1: Build =====
FROM golang:1.26.1-alpine AS builder

WORKDIR /app

# Copy go.mod and go.sum first (for dependency caching)
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy the rest of the source code
COPY . .

# Build the binary
RUN go build -o healthcheck .

# ===== Stage 2: Run =====
FROM alpine:latest

WORKDIR /app

# Copy the binary from builder stage
COPY --from=builder /app/healthcheck .

# Expose port
EXPOSE 8080

# Run the binary
CMD ["./healthcheck"]
