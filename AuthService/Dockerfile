# Stage 1: Build the Go binary
FROM golang:1.22 AS builder

WORKDIR /app

# Copy go.mod and go.sum, then download dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the code
COPY . .

# Build the Go app
RUN go build -o auth-service main.go

# Stage 2: Create a minimal runtime image
FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Copy the binary from the builder
COPY --from=builder /app/auth-service .

# Expose the service port
EXPOSE 8080

# Run the binary
CMD ["./auth-service"]
