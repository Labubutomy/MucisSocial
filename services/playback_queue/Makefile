.PHONY: proto clean build

proto:
	@echo "Generating playback queue proto..."
	protoc --go_out=. --go_opt=paths=source_relative \
		--go-grpc_out=. --go-grpc_opt=paths=source_relative \
		api/queue.proto
	@echo "Playback queue proto generated"

build:
	@mkdir -p bin
	go build -o bin/playback-queue cmd/main.go

clean:
	@rm -f api/*.pb.go
	@rm -rf bin
