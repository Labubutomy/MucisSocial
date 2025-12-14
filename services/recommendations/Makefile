.PHONY: build run test clean docker

# Build the service
build:
	go build -o bin/recommendation-service ./cmd/main.go

# Run the service locally
run:
	go run ./cmd/main.go

# Run tests
test:
	go test -v ./...

# Clean build artifacts
clean:
	rm -rf bin/

# Build Docker image
docker:
	docker build -t recommendation-service .

# Run with Docker
docker-run:
	docker run -p 8080:8080 \
		-e KAFKA_BROKERS=host.docker.internal:19092 \
		recommendation-service

# Tidy dependencies
tidy:
	go mod tidy

# Format code
fmt:
	go fmt ./...

# Lint code
lint:
	golangci-lint run
