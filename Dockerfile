
# Stage 1: Build the Go binary
FROM golang:1.24-alpine AS builder

WORKDIR /app

# Copy go mod and sum files to leverage Docker cache
COPY go.mod go.sum ./
RUN go mod download

# Copy the source code
COPY . .

# Build the application
# CGO_ENABLED=0 is important for creating a static binary for Alpine
# -ldflags="-w -s" strips debugging information, making the binary smaller
RUN CGO_ENABLED=0 GOOS=linux go build -a -ldflags="-w -s" -o /mediaguard-api ./cmd/mediaguard-api

# Stage 2: Create the final, minimal image
FROM ubuntu:24.10

# Add certificates for potential HTTPS calls
RUN apt-get update && apt-get install -y ffmpeg ca-certificates && rm -rf /var/lib/apt/lists/*

WORKDIR /app/

# Copy the pre-built binary from the builder stage
COPY --from=builder /mediaguard-api .
COPY ./bin/desc_tools .

# Expose port 8080 to the outside world
EXPOSE 8080

# Command to run the executable
CMD ["./mediaguard-api"]
