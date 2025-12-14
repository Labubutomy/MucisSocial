package config

import "os"

type Config struct {
	HTTPPort             string
	KafkaBrokers         []string
	TrackEventsTopic     string
	ListeningEventsTopic string
	TracksServiceURL     string // URL for tracks service API (for bootstrap)
	BootstrapEnabled     bool   // Enable loading tracks on startup
}

func Load() *Config {
	return &Config{
		HTTPPort:             getEnv("HTTP_PORT", "8080"),
		KafkaBrokers:         []string{getEnv("KAFKA_BROKERS", "redpanda:9092")},
		TrackEventsTopic:     getEnv("TRACK_EVENTS_TOPIC", "track-events"),
		ListeningEventsTopic: getEnv("LISTENING_EVENTS_TOPIC", "listening-events"),
		TracksServiceURL:     getEnv("TRACKS_SERVICE_URL", ""),
		BootstrapEnabled:     getEnv("BOOTSTRAP_ENABLED", "true") == "true",
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
