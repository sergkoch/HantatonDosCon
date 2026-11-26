package config

import (
	"os"
	"strconv"
	"strings"
)

type Config struct {
	Listen             string
	DatabaseURL        string
	AIBaseURL          string
	AIKey              string
	AIVision           string
	AIText             string
	AllowedHosts       []string
	RateLimitPerMinute int
}

func Load() Config {
	return Config{
		Listen:             env("LISTEN", ":8080"),
		DatabaseURL:        env("DATABASE_URL", ""),
		AIBaseURL:          env("AI_BASE_URL", ""),
		AIKey:              env("AI_API_KEY", ""),
		AIVision:           env("AI_VISION_MODEL", ""),
		AIText:             env("AI_TEXT_MODEL", ""),
		AllowedHosts:       list("SCAN_ALLOWED_HOSTS"),
		RateLimitPerMinute: envInt("RATE_LIMIT_PER_MINUTE", 20),
	}
}

func env(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func envInt(key string, fallback int) int {
	value, err := strconv.Atoi(env(key, ""))
	if err != nil || value <= 0 {
		return fallback
	}
	return value
}

func list(key string) []string {
	raw := strings.TrimSpace(os.Getenv(key))
	if raw == "" {
		return nil
	}

	parts := strings.Split(raw, ",")
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			result = append(result, part)
		}
	}
	return result
}
