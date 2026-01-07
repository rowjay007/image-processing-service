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

# Common runtime stage
FROM alpine:latest AS runtime
RUN apk add --no-cache vips ca-certificates
WORKDIR /app

# API stage
FROM runtime AS api
COPY --from=builder /app/api /app/api
EXPOSE 8080
CMD ["/app/api"]

# Worker stage
FROM runtime AS worker
COPY --from=builder /app/worker /app/worker
CMD ["/app/worker"]
