package config

import "os"

type Config struct {
	HTTPPort             string
	KafkaBrokers         []string
	TrackEventsTopic     string
	ListeningEventsTopic string
	TracksServiceURL     string // URL for tracks service API (for bootstrap)
	BootstrapEnabled     bool   // Enable loading tracks on startup

	// MinIO backup configuration
	MinIOEndpoint   string
	MinIOAccessKey  string
	MinIOSecretKey  string
	MinIOBucketName string
	BackupInterval  string // e.g., "5m", "1h"
}

func Load() *Config {
	return &Config{
		HTTPPort:             getEnv("HTTP_PORT", "8080"),
		KafkaBrokers:         []string{getEnv("KAFKA_BROKERS", "redpanda:9092")},
		TrackEventsTopic:     getEnv("TRACK_EVENTS_TOPIC", "track-events"),
		ListeningEventsTopic: getEnv("LISTENING_EVENTS_TOPIC", "listening-events"),
		TracksServiceURL:     getEnv("TRACKS_SERVICE_URL", ""),
		BootstrapEnabled:     getEnv("BOOTSTRAP_ENABLED", "true") == "true",

		// MinIO backup config
		MinIOEndpoint:   getEnv("MINIO_ENDPOINT", "minio:9000"),
		MinIOAccessKey:  getEnv("MINIO_ACCESS_KEY", "minioadmin"),
		MinIOSecretKey:  getEnv("MINIO_SECRET_KEY", "minioadmin"),
		MinIOBucketName: getEnv("MINIO_BUCKET_NAME", "recommendations"),
		BackupInterval:  getEnv("BACKUP_INTERVAL", "5m"),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
