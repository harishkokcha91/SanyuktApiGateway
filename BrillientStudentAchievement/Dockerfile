# Stage 1: Build the Go app
FROM golang:1.22 AS builder

# Set the Current Working Directory inside the container
WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download all dependencies. Dependencies will be cached if the go.mod and go.sum files are not changed
RUN go mod download

# Copy the source code into the container
COPY . .

# Build the Go app
RUN GOOS=linux GOARCH=amd64 go build -o brilliant-student main.go

# Stage 2: Minimal runtime container
FROM alpine:latest

# Install certificates for HTTPS requests
RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Copy the Pre-built binary file from the previous stage
COPY --from=builder /app/brilliant-student .

# Expose port
EXPOSE 8085

# Command to run the executable
CMD ["./brilliant-student"]
