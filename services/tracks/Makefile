.PHONY: proto generate build run docker-build docker-up docker-down help

# Генерация кода из proto файлов
proto: generate

generate:
	@echo "Generating gRPC code from proto files..."
	@mkdir -p api
	@export PATH=$$PATH:$$(go env GOPATH)/bin && \
	protoc --go_out=. --go_opt=paths=source_relative \
		--go-grpc_out=. --go-grpc_opt=paths=source_relative \
		api/tracks.proto
	@echo "Proto code generated successfully!"

# Сборка приложения
build: generate
	@echo "Building application..."
	@go build -o bin/tracks-service ./cmd/main.go
	@echo "Build completed!"

# Запуск приложения локально
run: generate
	@echo "Running application..."
	@go run cmd/main.go

# Сборка Docker образа
docker-build:
	@echo "Building Docker image..."
	@docker build -t tracks-service:latest .
	@echo "Docker image built successfully!"

# Запуск через docker-compose
docker-up:
	@echo "Starting services with docker-compose..."
	@docker-compose up -d
	@echo "Services started!"

# Остановка docker-compose
docker-down:
	@echo "Stopping services..."
	@docker-compose down
	@echo "Services stopped!"

# Показать помощь
help:
	@echo "Available commands:"
	@echo "  make generate     - Generate code from proto files"
	@echo "  make build        - Build the application"
	@echo "  make run          - Run the application locally"
	@echo "  make docker-build - Build Docker image"
	@echo "  make docker-up    - Start services with docker-compose"
	@echo "  make docker-down  - Stop services"

