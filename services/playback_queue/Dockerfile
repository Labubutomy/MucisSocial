FROM golang:1.24-alpine AS builder

# Install protoc and dependencies
RUN apk add --no-cache protobuf protobuf-dev

# Install Go plugins for protoc
RUN go install google.golang.org/protobuf/cmd/protoc-gen-go@latest && \
    go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

WORKDIR /src

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy proto files
COPY api/queue.proto ./api/

# Generate code from proto files
RUN export PATH=$PATH:/root/go/bin && \
    protoc --go_out=. --go_opt=paths=source_relative \
           --go-grpc_out=. --go-grpc_opt=paths=source_relative \
           api/queue.proto

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o playback-queue ./cmd/main.go

FROM alpine:3.20
RUN apk --no-cache add ca-certificates curl \
    && curl -fsSL -o /usr/local/bin/grpc_health_probe https://github.com/grpc-ecosystem/grpc-health-probe/releases/download/v0.4.24/grpc_health_probe-linux-amd64 \
    && chmod +x /usr/local/bin/grpc_health_probe \
    && apk del curl

WORKDIR /app
COPY --from=builder /src/playback-queue ./playback-queue

EXPOSE 50056

ENTRYPOINT ["/app/playback-queue"]
