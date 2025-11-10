FROM golang:1.23-alpine AS build

# Install protoc and dependencies
RUN apk add --no-cache git protobuf-dev protoc

# Install Go plugins for protoc
RUN go install google.golang.org/protobuf/cmd/protoc-gen-go@latest && \
    go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

WORKDIR /src

# Copy go mod files
COPY transcoder/go.mod transcoder/go.sum ./
RUN go mod download

# Copy proto file from tracks service
COPY tracks/api/tracks.proto ./proto/tracks/v1/tracks.proto

# Fix go_package in tracks.proto for transcoder module
RUN sed -i 's|option go_package = "github.com/Labubutomy/MucisSocial/services/tracks/api;tracks";|option go_package = "github.com/MusicSocial/transcoder/proto/tracks/v1;trackspb";|' proto/tracks/v1/tracks.proto

# Generate proto files into proto directory
RUN export PATH=$PATH:/root/go/bin && \
    mkdir -p proto/tracks/v1 && \
    protoc --proto_path=proto \
        --go_out=proto --go_opt=paths=source_relative \
        --go-grpc_out=proto --go-grpc_opt=paths=source_relative \
        proto/tracks/v1/tracks.proto

# Copy source code
COPY transcoder/cmd ./cmd
COPY transcoder/internal ./internal

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o transcoder ./cmd/main.go

FROM alpine:3.20

RUN apk add --no-cache ca-certificates ffmpeg

WORKDIR /app

COPY --from=build /src/transcoder .

ENV KAFKA_BROKERS=redpanda:9092 \
    TRANSCODER_TOPIC=transcoder-tasks \
    TRANSCODER_GROUP_ID=transcoder-service

ENTRYPOINT ["./transcoder"]

