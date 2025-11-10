package config

import (
	"os"
)

type Config struct {
	Server   ServerConfig
	MinIO    MinIOConfig
	Redpanda RedpandaConfig
	Tracks   TrackServiceConfig
}

type ServerConfig struct {
	GRPCPort string
}

type MinIOConfig struct {
	Endpoint        string
	AccessKeyID     string
	SecretAccessKey string
	BucketName      string
	UseSSL          bool
}

type RedpandaConfig struct {
	Brokers         []string
	TranscoderTopic string
}

type TrackServiceConfig struct {
	Address string
}

func Load() *Config {
	return &Config{
		Server: ServerConfig{
			GRPCPort: getEnv("GRPC_PORT", "50051"),
		},
		MinIO: MinIOConfig{
			Endpoint:        getEnv("MINIO_ENDPOINT", "minio:9000"),
			AccessKeyID:     getEnv("MINIO_ACCESS_KEY", "minioadmin"),
			SecretAccessKey: getEnv("MINIO_SECRET_KEY", "minioadmin"),
			BucketName:      getEnv("MINIO_BUCKET", "tracks"),
			UseSSL:          false,
		},
		Redpanda: RedpandaConfig{
			Brokers:         []string{getEnv("REDPANDA_BROKERS", "redpanda:9092")},
			TranscoderTopic: getEnv("TRANSCODER_TOPIC", "transcoder-tasks"),
		},
		Tracks: TrackServiceConfig{
			Address: getEnv("TRACK_SERVICE_ADDR", "tracks-service:50051"),
		},
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
