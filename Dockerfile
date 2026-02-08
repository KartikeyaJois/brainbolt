# Build stage
FROM golang:alpine AS builder

# Set working directory
WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy the source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o brainbolt ./cmd/brainbolt

# Run stage
FROM alpine:latest

# Add ca-certificates for potential HTTPS requests
RUN apk --no-cache add ca-certificates

# Set working directory
WORKDIR /root/

# Copy the binary from the builder stage
COPY --from=builder /app/brainbolt .

# Expose the port the app runs on
EXPOSE 3001

# Command to run the application
CMD ["./brainbolt"]
