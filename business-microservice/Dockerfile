# Stage 1: Build the Go binary
FROM golang:1.22 AS builder

# Set the working directory inside the container
WORKDIR /app

# Copy dependency files first
COPY go.mod go.sum ./
RUN go mod download

# Copy the entire project source
COPY . .

# Build the Go binary
RUN go build -o business-microservice main.go

# Stage 2: Create a minimal image
FROM alpine:latest

# Add SSL certificates (for secure DB/API calls)
RUN apk --no-cache add ca-certificates

# Set working directory
WORKDIR /root/

# Copy the binary from the builder stage
COPY --from=builder /app/business-microservice .

# Expose the application's port
EXPOSE 8088

# Command to run the app
CMD ["./business-microservice"]
