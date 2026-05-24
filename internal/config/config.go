package config

import (
	"os"
)

type Config struct {
	Port    string
	DBUrl   string
	RedisURL string
}

func Load() *Config {
	return &Config{
		Port:     getEnv("PORT", "8080"),
		DBUrl:    getEnv("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/nustdevs?sslmode=disable"),
		RedisURL: getEnv("REDIS_URL", "redis://localhost:6379"),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
