.PHONY: build run docker-build docker-up docker-down help

# Сборка приложения
build:
	@echo "Building application..."
	@go build -o bin/playlist-service ./cmd/main.go
	@echo "Build completed!"

# Запуск приложения локально
run:
	@echo "Running application..."
	@go run cmd/main.go

# Сборка Docker образа
docker-build:
	@echo "Building Docker image..."
	@docker build -t playlist-service:latest .
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
	@echo "  make build        - Build the application"
	@echo "  make run          - Run the application locally"
	@echo "  make docker-build - Build Docker image"
	@echo "  make docker-up    - Start services with docker-compose"
	@echo "  make docker-down  - Stop services"

