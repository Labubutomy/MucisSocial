# Build stage
FROM golang:1.23-alpine AS builder

# Install necessary tools
RUN apk add --no-cache git

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod ./

# Download dependencies
RUN go mod download && go mod tidy

# Copy source code and proto files
COPY . .

# Update modules after copying source code
RUN go mod tidy

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o gateway .

# Final stage
FROM alpine:latest

# Install ca-certificates and wget for healthcheck
RUN apk --no-cache add ca-certificates wget

WORKDIR /root/

# Copy the binary from builder stage
COPY --from=builder /app/gateway .

# Expose port
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=30s --retries=3 \
  CMD wget -q --spider http://localhost:8080/__health || exit 1

# Run the gateway
CMD ["./gateway"]