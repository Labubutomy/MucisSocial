# Build stage
FROM golang:1.23-alpine AS builder

# Install necessary tools
RUN apk add --no-cache git protobuf-dev protoc

# Set working directory
WORKDIR /app

# Copy go mod files
COPY gateway/go.mod ./

# Download dependencies
RUN go mod download && go mod tidy

# Install protobuf Go plugins and swag
ENV PATH="/root/go/bin:${PATH}"
RUN go install google.golang.org/protobuf/cmd/protoc-gen-go@latest && \
    go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest && \
    go install github.com/swaggo/swag/cmd/swag@latest

# Copy gateway source code
COPY gateway/ .

# Copy proto files from users service
COPY users/proto/users proto/users

# Generate protobuf files
RUN mkdir -p proto/users/v1 proto/artists/v1 && \
    protoc --proto_path=proto \
        --go_out=proto --go_opt=paths=source_relative \
        --go-grpc_out=proto --go-grpc_opt=paths=source_relative \
        proto/users/v1/user_service.proto proto/artists/v1/artist_service.proto

# Generate Swagger documentation
RUN swag init || true

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