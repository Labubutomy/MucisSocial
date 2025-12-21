package internal

import (
	"os"
	"strconv"
	"time"
)

type Config struct {
	SuggestionLimit    int
	SuggestionCooldown time.Duration
	LinkBaseURL        string
	QueueServiceAddr   string
}

func LoadConfig() Config {
	limit := parseIntEnv("SUGGESTION_LIMIT", 10)
	cooldownSeconds := parseIntEnv("SUGGESTION_COOLDOWN_SECONDS", 1800)

	return Config{
		SuggestionLimit:    limit,
		SuggestionCooldown: time.Duration(cooldownSeconds) * time.Second,
		LinkBaseURL:        getEnv("GROUP_LINK_BASE", "https://music.local/groups"),
		QueueServiceAddr:   getEnv("QUEUE_SERVICE_ADDR", "playback-queue-service:50056"),
	}
}

func parseIntEnv(key string, fallback int) int {
	if value := os.Getenv(key); value != "" {
		if parsed, err := strconv.Atoi(value); err == nil && parsed > 0 {
			return parsed
		}
	}
	return fallback
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
