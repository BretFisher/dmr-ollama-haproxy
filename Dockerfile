# Build stage
FROM golang:1.24-alpine AS builder

# Install git and ca-certificates (needed for go mod download)
RUN apk add --no-cache git ca-certificates

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application with optimizations
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags="-w -s" -o dmr-models-convert .

# Final stage
FROM alpine:latest

# Set working directory
WORKDIR /app

# Copy binary from builder stage
COPY --from=builder /app/dmr-models-convert .

# download DMR models and convert to Ollama format
# NOTE: you'll want to bind mount the output directory so you can access the models.json file
CMD ["./dmr-models-convert", "--dmr", "http://model-runner.docker.internal/models", "--output", "/output/models.json"] 