package env

import (
	"fmt"
	"os"
)

func GetConfig() Config {
	return Config{
		Port:       envWithFallback("PORT", "3000"),
		BackendURL: require("BACKEND_URL"),
		RedisURL:   envWithFallback("REDIS_URL", "redis://localhost"),
	}
}

type Config struct {
	Port       string
	BackendURL string
	RedisURL   string
}

func envWithFallback(key, fallback string) string {
	if fromEnv := os.Getenv(key); fromEnv != "" {
		return fromEnv
	}

	return fallback
}

func require(key string) string {
	if fromEnv := os.Getenv(key); fromEnv != "" {
		return fromEnv
	}

	panic(fmt.Sprintf("Required environment variable %v is not set", key))
}
