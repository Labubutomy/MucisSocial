package server

import (
	"testing"

	"github.com/MucisSocial/upload/internal/config"
)

func TestNewUploadServer(t *testing.T) {
	cfg := &config.Config{
		Server: config.ServerConfig{
			GRPCPort: "50051",
		},
		MinIO: config.MinIOConfig{
			Endpoint:        "localhost:9000",
			AccessKeyID:     "minioadmin",
			SecretAccessKey: "minioadmin",
			BucketName:      "tracks",
			UseSSL:          false,
		},
		Redpanda: config.RedpandaConfig{
			Brokers:         []string{"localhost:9092"},
			TranscoderTopic: "transcoder-tasks",
		},
		Tracks: config.TrackServiceConfig{
			Address: "localhost:50052",
		},
	}

	server := NewUploadServer(cfg, nil, nil, nil)

	if server == nil {
		t.Fatal("Expected non-nil server")
	}

	if server.config != cfg {
		t.Error("Config not set properly")
	}
}
