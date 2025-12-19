FROM golang:1.24-alpine AS builder

WORKDIR /app

RUN apk add --no-cache make protobuf

RUN go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.36.10 && \
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.4.0

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN make proto

RUN CGO_ENABLED=0 GOOS=linux go build -o groups-service ./cmd/main.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates

WORKDIR /root/
COPY --from=builder /app/groups-service ./groups-service

EXPOSE 50058

CMD ["./groups-service"]
