# ===== Stage 1: Build =====
FROM golang:1.22-alpine AS builder

# Install git and ca-certificates (needed for go modules + HTTPS)
RUN apk add --no-cache git ca-certificates tzdata

WORKDIR /app

# Copy go module files first (layer caching)
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build a statically linked binary
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-w -s" \
    -o server .

# ===== Stage 2: Runtime =====
FROM scratch

# Carry over CA certs and timezone data from builder
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo

# Copy the compiled binary
COPY --from=builder /app/server /server

EXPOSE 8080

ENTRYPOINT ["/server"]