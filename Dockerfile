# Build stage
FROM golang:1.26-alpine AS builder

WORKDIR /app

# Install build dependencies
RUN apk add --no-cache git

# Copy dependencies and download
COPY go.mod go.sum ./
RUN go mod download

# Copy source code and build
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o bazinga main.go

# Run stage
FROM alpine:latest

# Install necessary runtime dependencies
RUN apk add --no-cache ca-certificates openssh-client

WORKDIR /app

# Copy binary from builder
COPY --from=builder /app/bazinga .

# Create a directory for the host key
RUN mkdir -p /data

# Default configuration for SSH
ENV BAZINGA__SSH__ENABLED=true
ENV BAZINGA__SSH__ADDR=:2222

# Expose the SSH port
EXPOSE 2222

# Run the application
ENTRYPOINT ["./bazinga"]
