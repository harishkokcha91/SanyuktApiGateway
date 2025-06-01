# Stage 1: Build the Go binary
FROM golang:1.22 AS builder

# Set the working directory
WORKDIR /app

# Copy Go mod and sum files and download dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN go build -o upload-service main.go

# Stage 2: Create a minimal image
FROM alpine:latest

# Install required packages
RUN apk --no-cache add ca-certificates

# Set the working directory
WORKDIR /root/

# Create uploads directory for storing uploaded files
RUN mkdir uploads

# Copy binary from builder stage
COPY --from=builder /app/upload-service .

# Expose the upload directory and port
VOLUME ["/root/uploads"]
EXPOSE 8086

# Command to run the app
CMD ["./upload-service"]
