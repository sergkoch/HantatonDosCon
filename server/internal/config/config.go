package config

import "os"

type Config struct {
	Listen      string
	DatabaseURL string
	AIBaseURL   string
	AIKey       string
	AIVision    string
	AIText      string
}

func Load() Config {
	return Config{
		Listen:      env("LISTEN", ":8080"),
		DatabaseURL: env("DATABASE_URL", ""),
		AIBaseURL:   env("AI_BASE_URL", ""),
		AIKey:       env("AI_API_KEY", ""),
		AIVision:    env("AI_VISION_MODEL", ""),
		AIText:      env("AI_TEXT_MODEL", ""),
	}
}

func env(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
