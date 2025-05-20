# Start from the official Go image (as a build stage)
FROM golang:1.24-alpine AS builder

# Enable Go modules and set up working directory
WORKDIR /app

# Install git (needed for some dependencies)
RUN apk add --no-cache git

# Copy go.mod and go.sum first, then download deps (leverages Docker layer caching)
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of your source code
COPY . .

# Build the binary
RUN go build -o server .

# Final lightweight image
FROM alpine:latest

WORKDIR /root/

# Copy the binary from builder stage
COPY --from=builder /app/server .

# Set environment variables
ENV PORT=8080

# Expose the port
EXPOSE 8080

# Start the server
CMD ["./server"]
