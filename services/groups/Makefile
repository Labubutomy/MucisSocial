APP_NAME=groups-service

.PHONY: proto build clean

proto:
	@echo "Syncing playback-queue proto..."
	@mkdir -p proto/queue
	@if [ -f ../playback_queue/api/queue.proto ]; then \
		sed 's#github.com/Labubutomy/MucisSocial/services/playback_queue/api;queuepb#github.com/Labubutomy/MucisSocial/services/groups/proto/queue;queuepb#' \
			../playback_queue/api/queue.proto > proto/queue/queue.proto; \
	fi
	@echo "Generating playback-queue client stubs..."
	mkdir -p proto/queue
	protoc --go_out=. --go_opt=paths=source_relative \
		--go-grpc_out=. --go-grpc_opt=paths=source_relative \
		proto/queue/queue.proto
	@echo "Generating groups proto..."
	protoc --go_out=. --go_opt=paths=source_relative \
		--go-grpc_out=. --go-grpc_opt=paths=source_relative \
		api/groups.proto
	@echo "Proto files generated successfully"

build:
	@echo "Building $(APP_NAME)..."
	@mkdir -p bin
	go build -o bin/$(APP_NAME) cmd/main.go
	@echo "Build completed"

clean:
	@echo "Removing generated proto files..."
	@rm -f api/*.pb.go
	@rm -rf proto/
	@echo "Removing build artifacts..."
	@rm -rf bin/
	@echo "Done"
