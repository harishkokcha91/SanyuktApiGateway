# Stage 1: Build the Go binary
FROM golang:1.22 AS builder

# Set the working directory
WORKDIR /app

# Copy Go module files and download dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the Go binary
RUN go build -o user-service main.go

# Stage 2: Create minimal runtime image
FROM alpine:latest

# Install root certificates for HTTPS
RUN apk --no-cache add ca-certificates

# Set working directory
WORKDIR /root/

# Copy built binary from the builder stage
COPY --from=builder /app/user-service .

# Expose the app's port
EXPOSE 8083

# Start the service
CMD ["./user-service"]
