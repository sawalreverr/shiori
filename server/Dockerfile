# Build stage
FROM golang:1.25-alpine AS builder

WORKDIR /app

# Install git for go mod download if needed
RUN apk add --no-cache git

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 go build -o server ./cmd/server

# Final stage
FROM alpine:latest

WORKDIR /app

# Install CA certs for HTTPS
RUN apk add --no-cache ca-certificates

# Copy binary
COPY --from=builder /app/server /usr/local/bin/server

# Create data directory
RUN mkdir -p /data

# Set environment
ENV PORT=8080
ENV DATABASE_PATH=/data/shiori.db

EXPOSE 8080

# Run server directly
CMD ["/usr/local/bin/server"]
