package config

import (
	"os"
	"strings"
	"time"
)

type Config struct {
	Kafka        KafkaConfig
	MinIO        MinIOConfig
	TrackService TrackServiceConfig
	WorkDir      string
}

type KafkaConfig struct {
	Brokers        []string
	Topic          string
	GroupID        string
	MinBytes       int
	MaxBytes       int
	CommitInterval time.Duration
}

type MinIOConfig struct {
	Endpoint        string
	AccessKeyID     string
	SecretAccessKey string
	BucketName      string
}

type TrackServiceConfig struct {
	Address string
}

func Load() Config {
	cfg := Config{
		Kafka: KafkaConfig{
			Brokers:        splitAndTrim(getEnv("KAFKA_BROKERS", "localhost:9092")),
			Topic:          getEnv("TRANSCODER_TOPIC", "transcoder-tasks"),
			GroupID:        getEnv("TRANSCODER_GROUP_ID", "transcoder-service"),
			MinBytes:       1,
			MaxBytes:       10 * 1024 * 1024,
			CommitInterval: time.Second,
		},
		MinIO: MinIOConfig{
			Endpoint:        getEnv("MINIO_ENDPOINT", "minio:9000"),
			AccessKeyID:     getEnv("MINIO_ACCESS_KEY", "minioadmin"),
			SecretAccessKey: getEnv("MINIO_SECRET_KEY", "minioadmin"),
			BucketName:      getEnv("MINIO_BUCKET", "tracks"),
		},
		TrackService: TrackServiceConfig{
			Address: getEnv("TRACK_SERVICE_ADDR", "track-service:50052"),
		},
		WorkDir: getEnv("TRANSCODER_WORKDIR", os.TempDir()),
	}

	return cfg
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func splitAndTrim(value string) []string {
	parts := strings.Split(value, ",")
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		p := strings.TrimSpace(part)
		if p != "" {
			result = append(result, p)
		}
	}
	return result
}
