# Use the official Go image as base
FROM golang:1.23.4-alpine AS builder

# Set working directory inside the container
WORKDIR /app

# Install build dependencies for CGO (required for SQLite)
RUN apk add --no-cache gcc musl-dev sqlite-dev

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy the source code
COPY . .

# Build the application
RUN CGO_ENABLED=1 GOOS=linux go build -a -installsuffix cgo -o main ./cmd/main.go

# Use a minimal base image for the final stage
FROM alpine:latest

# Install SQLite and CA certificates
RUN apk --no-cache add ca-certificates sqlite

# Set working directory
WORKDIR /root/

# Copy the binary from builder stage
COPY --from=builder /app/main .

# Copy static files and templates
COPY --from=builder /app/web ./web

# Copy internal directory for migrations
COPY --from=builder /app/internal ./internal

# Create directory for the database
RUN mkdir -p /root/data

# Expose port 8080
EXPOSE 8080

# Set environment variable for database path (optional)
ENV DB_PATH=/root/data/forum.db

# Run the application
CMD ["./main"]
