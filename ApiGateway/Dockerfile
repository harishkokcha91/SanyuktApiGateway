# Stage 1: Build
FROM golang:1.22.11-alpine AS builder

ENV CGO_ENABLED=0 GOOS=linux GOARCH=amd64

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Build the binary from the correct path
RUN go build -o api-gateway .


# Stage 2: Runtime
FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/api-gateway .

ENTRYPOINT ["/app/api-gateway"]
