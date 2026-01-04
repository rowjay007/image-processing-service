# Build stage
FROM golang:1.23-alpine AS builder

# Install build dependencies
RUN apk add --no-cache gcc musl-dev vips-dev

WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build API
RUN go build -o /app/api cmd/api/main.go

# Build Worker
RUN go build -o /app/worker cmd/worker/main.go

# Final stage
FROM alpine:latest

# Install runtime dependencies for vips
RUN apk add --no-cache vips ca-certificates

WORKDIR /app

# Copy binaries from builder
COPY --from=builder /app/api /app/api
COPY --from=builder /app/worker /app/worker

# Default to API
EXPOSE 8080
CMD ["/app/api"]
