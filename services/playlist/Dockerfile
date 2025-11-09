# Build stage
FROM golang:1.24-alpine AS builder

# Install protoc and dependencies
RUN apk add --no-cache protobuf protobuf-dev

# Install Go plugins for protoc
RUN go install google.golang.org/protobuf/cmd/protoc-gen-go@latest && \
    go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy proto files
COPY api/playlist.proto ./api/

# Generate code from proto files
RUN export PATH=$PATH:/root/go/bin && \
    protoc --go_out=. --go_opt=paths=source_relative \
           --go-grpc_out=. --go-grpc_opt=paths=source_relative \
           api/playlist.proto

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/bin/playlist-service ./cmd/main.go

# Runtime stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Copy the binary from builder
COPY --from=builder /app/bin/playlist-service .

EXPOSE 8080 50054

CMD ["./playlist-service"]

