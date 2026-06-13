package config

import (
	"os"
	"strings"
)

type Config struct {
	Port          string
	DBUrl         string
	RedisURL      string
	JWTSecret     string
	OpenRouterKey string
	AIModel       string
}

func Load() *Config {
	return &Config{
		Port:          getEnv("PORT", "8080"),
		DBUrl:         getEnv("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/nustdevs?sslmode=disable"),
		RedisURL:      getEnv("REDIS_URL", "redis://127.0.0.1:6379"),
		JWTSecret:     getEnv("JWT_SECRET", "change_me_in_production"),
		OpenRouterKey: getEnv("OPENROUTER_API_KEY", ""),
		AIModel:       getEnv("AI_MODEL", "openai/gpt-oss-120b:free"),
	}
}

func (c *Config) RedisAddr() string {
	addr := strings.TrimPrefix(c.RedisURL, "redis://")
	addr = strings.Replace(addr, "localhost", "127.0.0.1", 1)
	return addr
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
