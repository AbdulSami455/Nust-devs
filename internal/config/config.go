package config

import (
	"os"
	"strings"
)

type Config struct {
	Port     string
	DBUrl    string
	RedisURL string
	JWTSecret string
}

func Load() *Config {
	return &Config{
		Port:      getEnv("PORT", "8080"),
		DBUrl:     getEnv("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/nustdevs?sslmode=disable"),
		RedisURL:  getEnv("REDIS_URL", "redis://localhost:6379"),
		JWTSecret: getEnv("JWT_SECRET", "change_me_in_production"),
	}
}

// RedisAddr strips the scheme from RedisURL for libraries that want host:port.
func (c *Config) RedisAddr() string {
	addr := c.RedisURL
	if after, ok := strings.CutPrefix(addr, "redis://"); ok {
		addr = after
	}
	return addr
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
