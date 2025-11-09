# Базовый образ для сборки
FROM golang:1.23-alpine AS builder

# Устанавливаем необходимые инструменты
RUN apk add --no-cache git protobuf-dev

# Устанавливаем protoc-gen-go и protoc-gen-go-grpc
RUN go install google.golang.org/protobuf/cmd/protoc-gen-go@latest && \
    go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

WORKDIR /app

# Копируем proto файлы
COPY proto ./proto

# Генерируем proto файлы
RUN protoc --go_out=. --go_opt=paths=source_relative \
    --go-grpc_out=. --go-grpc_opt=paths=source_relative \
    proto/upload.proto proto/track.proto

# Копируем go.mod и весь код
COPY go.mod ./
COPY . .

# Обновляем зависимости (создаст новый go.sum)
RUN go mod tidy

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

