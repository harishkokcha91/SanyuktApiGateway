# Stage 1: Build the Go binary
FROM golang:1.22 AS builder

# Set working directory
WORKDIR /app

# Copy module files and download dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy the source code
COPY . .

# Build the binary
RUN go build -o namdev-events main.go

# Stage 2: Minimal runtime image
FROM alpine:latest

# Install certificates for HTTPS, if needed
RUN apk --no-cache add ca-certificates

# Set working directory
WORKDIR /root/

# Copy binary from builder
COPY --from=builder /app/namdev-events .

# Expose the port used in the Go service
EXPOSE 8087

# Run the binary
CMD ["./namdev-events"]
