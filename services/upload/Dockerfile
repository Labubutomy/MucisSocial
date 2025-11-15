# Базовый образ для сборки
FROM golang:1.23-alpine AS builder
ENV GOPROXY=https://proxy.golang.org,direct

RUN apk add --no-cache git protobuf-dev protoc
RUN go install google.golang.org/protobuf/cmd/protoc-gen-go@latest && \
    go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

WORKDIR /src

COPY upload/go.mod upload/go.sum ./
RUN go mod download

COPY upload/proto ./proto

# Обновляем proto треков из исходного сервиса и приводим структуру к v1
RUN rm -rf proto/tracks && mkdir -p proto/tracks/v1

# Копируем proto файл из tracks service
COPY tracks/api/tracks.proto ./proto/tracks/v1/tracks.proto

# Исправляем go_package в tracks.proto для upload модуля
RUN sed -i 's|option go_package = "github.com/Labubutomy/MucisSocial/services/tracks/api;tracks";|option go_package = "github.com/MusicSocial/upload/proto/tracks/v1;trackspb";|' proto/tracks/v1/tracks.proto

# Генерируем proto файлы
RUN mkdir -p proto/tracks/v1 && \
    protoc --go_out=. --go_opt=paths=source_relative \
    --go-grpc_out=. --go-grpc_opt=paths=source_relative \
    proto/upload.proto && \
    protoc --go_out=. --go_opt=paths=source_relative \
    --go-grpc_out=. --go-grpc_opt=paths=source_relative \
    proto/tracks/v1/tracks.proto

# Копируем исходный код
COPY upload/cmd ./cmd
COPY upload/internal ./internal
COPY upload/Makefile ./Makefile
COPY upload/README.md ./README.md

# Собираем приложение
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o upload cmd/main.go

# Финальный образ
FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /app

# Копируем бинарник из builder
COPY --from=builder /src/upload .

# Открываем порт gRPC
EXPOSE 50055

# Запускаем приложение
CMD ["./upload"]

