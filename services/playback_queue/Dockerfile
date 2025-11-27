FROM golang:1.24-alpine AS builder

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .

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
