FROM golang:1.23-alpine AS build

RUN apk add --no-cache git

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o transcoder ./cmd/main.go

FROM alpine:3.20

RUN apk add --no-cache ca-certificates ffmpeg

WORKDIR /app

COPY --from=build /src/transcoder .

ENV KAFKA_BROKERS=redpanda:9092 \
    TRANSCODER_TOPIC=transcoder-tasks \
    TRANSCODER_GROUP_ID=transcoder-service

ENTRYPOINT ["./transcoder"]

