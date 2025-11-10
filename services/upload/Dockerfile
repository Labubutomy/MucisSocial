# Базовый образ для сборки
FROM golang:1.23-alpine AS builder

# Устанавливаем необходимые инструменты
RUN apk add --no-cache git protobuf-dev protoc

# Устанавливаем protoc-gen-go и protoc-gen-go-grpc
RUN go install google.golang.org/protobuf/cmd/protoc-gen-go@latest && \
    go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

WORKDIR /app

# Копируем proto файлы upload
COPY upload/proto ./proto

# Копируем proto файл из tracks service
COPY tracks/api/tracks.proto ./proto/tracks/tracks.proto

# Исправляем go_package в tracks.proto для upload модуля
RUN sed -i 's|option go_package = "github.com/Labubutomy/MucisSocial/services/tracks/api;tracks";|option go_package = "github.com/MusicSocial/upload/proto/tracks;tracks";|' proto/tracks/tracks.proto

# Генерируем proto файлы
RUN mkdir -p proto/tracks && \
    protoc --go_out=. --go_opt=paths=source_relative \
    --go-grpc_out=. --go-grpc_opt=paths=source_relative \
    proto/upload.proto && \
    protoc --go_out=. --go_opt=paths=source_relative \
    --go-grpc_out=. --go-grpc_opt=paths=source_relative \
    proto/tracks/tracks.proto

# Копируем go.mod и go.sum
COPY upload/go.mod upload/go.sum ./
RUN go mod download

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

WORKDIR /root/

# Копируем бинарник из builder
COPY --from=builder /app/upload .

# Открываем порт gRPC
EXPOSE 50051

# Запускаем приложение
CMD ["./upload"]

