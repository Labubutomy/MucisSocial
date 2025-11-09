.PHONY: help build run stop clean proto migrate test

# Default target
help: ## Show this help message
	@echo "Available commands:"
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-15s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

# Build the user service
build: ## Build user service Docker image
	cd services/users && docker build -t music-social-users:latest .

# Generate protobuf files
proto: ## Generate protobuf Go files
	protoc --proto_path=proto \
		--go_out=proto --go_opt=paths=source_relative \
		--go-grpc_out=proto --go-grpc_opt=paths=source_relative \
		proto/users/v1/user_service.proto

# Run development environment
run: ## Start all services
	docker-compose up -d
	
# Run only services (no apps)
run-services: ## Start only backend services
	docker-compose -f services/docker-compose.yml up -d

# Stop all services
stop: ## Stop all services
	docker-compose down
	docker-compose -f services/docker-compose.yml down

# Clean up
clean: ## Clean up containers, images and volumes
	docker-compose down -v
	docker-compose -f services/docker-compose.yml down -v
	docker system prune -f

# Run database migrations
migrate: ## Run database migrations
	cd services/users && \
	migrate -path migrations -database "postgres://postgres:password@localhost:5432/music_social_users?sslmode=disable" up

# Create a new migration
migrate-create: ## Create new migration file (usage: make migrate-create name=migration_name)
	cd services/users && \
	migrate create -ext sql -dir migrations -seq $(name)

# Run tests
test: ## Run tests
	cd services/users && go test -v ./...

# Install dependencies
deps: ## Install Go dependencies
	cd services/users && go mod tidy && go mod download

# Format code
fmt: ## Format Go code
	cd services/users && go fmt ./...

# Lint code
lint: ## Lint Go code
	cd services/users && golangci-lint run

# Generate mocks
mocks: ## Generate mocks for testing
	cd services/users && go generate ./...

# View logs
logs: ## View service logs
	docker-compose logs -f

logs-users: ## View users service logs
	docker-compose -f services/docker-compose.yml logs -f users-service

# Enter user service container
shell-users: ## Enter users service container shell
	docker-compose -f services/docker-compose.yml exec users-service sh

# Connect to database
db-shell: ## Connect to PostgreSQL database
	docker-compose -f services/docker-compose.yml exec postgres psql -U postgres -d music_social_users