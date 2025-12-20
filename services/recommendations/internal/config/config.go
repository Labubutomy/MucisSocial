package config

import (
	"os"
	"strconv"
	"time"
)

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

	// Geo Top configuration
	EnableGeoTop         bool
	GeohashPrecision     int
	GeoTopCacheTTL       time.Duration
	GeoTopMaxPerGeo      int
	GeoTopDefaultRadiusM int
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

		// Geo Top config
		EnableGeoTop:         getEnv("ENABLE_GEO_TOP", "false") == "true",
		GeohashPrecision:     getEnvInt("GEOHASH_PRECISION", 6),
		GeoTopCacheTTL:       getEnvDuration("GEO_TOP_CACHE_TTL", 60*time.Second),
		GeoTopMaxPerGeo:      getEnvInt("GEO_TOP_MAX_PER_GEO", 1000),
		GeoTopDefaultRadiusM: getEnvInt("GEO_TOP_DEFAULT_RADIUS_M", 1000),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intVal, err := strconv.Atoi(value); err == nil {
			return intVal
		}
	}
	return defaultValue
}

func getEnvDuration(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return defaultValue
}
