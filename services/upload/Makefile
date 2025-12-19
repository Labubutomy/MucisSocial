.PHONY: proto build clean

proto:
	@echo "Generating proto files..."
	@mkdir -p proto/tracks/v1
	@cp ../tracks/api/tracks.proto proto/tracks/v1/
	protoc --go_out=. --go_opt=paths=source_relative \
		--go-grpc_out=. --go-grpc_opt=paths=source_relative \
		proto/upload.proto proto/tracks/v1/tracks.proto
	@echo "Proto files generated successfully"

build:
	@echo "Building upload service..."
	@mkdir -p bin
	go build -o bin/upload cmd/main.go
	@echo "Build completed"

clean:
	@echo "Removing generated proto files..."
	@rm -f proto/*.pb.go
	@rm -rf proto/tracks
	@echo "Removing build artifacts..."
	@rm -rf bin/
	@echo "Done"
