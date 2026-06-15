package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	Port               string
	DBUrl              string
	RedisURL           string
	JWTSecret          string
	AdminEmail         string
	AdminPassword      string
	AllowedCORSOrigins []string
	SecureCookies      bool
	PublicRateLimit    int
	PublicRateWindow   time.Duration
	OpenRouterKey      string
	AIModel            string
}

func Load() *Config {
	return &Config{
		Port:          getEnv("PORT", "8080"),
		DBUrl:         getEnv("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/nustdevs?sslmode=disable"),
		RedisURL:      getEnv("REDIS_URL", "redis://127.0.0.1:6379"),
		JWTSecret:     strings.TrimSpace(os.Getenv("JWT_SECRET")),
		AdminEmail:    strings.ToLower(strings.TrimSpace(os.Getenv("ADMIN_EMAIL"))),
		AdminPassword: os.Getenv("ADMIN_PASSWORD"),
		AllowedCORSOrigins: splitCSV(getEnv(
			"CORS_ALLOWED_ORIGINS",
			"http://localhost:3000,http://127.0.0.1:3000",
		)),
		SecureCookies: getEnv("SECURE_COOKIES", "true") != "false",
		PublicRateLimit: getEnvInt(
			"PUBLIC_RATE_LIMIT_REQUESTS",
			600,
		),
		PublicRateWindow: getEnvDuration(
			"PUBLIC_RATE_LIMIT_WINDOW",
			time.Minute,
		),
		OpenRouterKey: getEnv("OPENROUTER_API_KEY", ""),
		AIModel:       getEnv("AI_MODEL", "openai/gpt-oss-120b:free"),
	}
}

func (c *Config) ValidateServer() error {
	if len(c.JWTSecret) < 32 {
		return fmt.Errorf("JWT_SECRET must be set to at least 32 characters")
	}
	switch c.JWTSecret {
	case "change_me", "change_me_in_production":
		return fmt.Errorf("JWT_SECRET uses an unsafe default value")
	}
	if (c.AdminEmail == "") != (c.AdminPassword == "") {
		return fmt.Errorf("ADMIN_EMAIL and ADMIN_PASSWORD must be set together")
	}
	if c.AdminPassword != "" && len(c.AdminPassword) < 12 {
		return fmt.Errorf("ADMIN_PASSWORD must be at least 12 characters")
	}
	if c.AdminPassword == "admin123" {
		return fmt.Errorf("ADMIN_PASSWORD uses an unsafe default value")
	}
	if len(c.AllowedCORSOrigins) == 0 {
		return fmt.Errorf("CORS_ALLOWED_ORIGINS must include at least one frontend origin")
	}
	for _, origin := range c.AllowedCORSOrigins {
		if origin == "*" {
			return fmt.Errorf("CORS_ALLOWED_ORIGINS cannot include wildcard origin")
		}
	}
	if c.PublicRateLimit < 1 {
		return fmt.Errorf("PUBLIC_RATE_LIMIT_REQUESTS must be greater than 0")
	}
	if c.PublicRateWindow <= 0 {
		return fmt.Errorf("PUBLIC_RATE_LIMIT_WINDOW must be greater than 0")
	}
	return nil
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

func getEnvInt(key string, fallback int) int {
	v := strings.TrimSpace(os.Getenv(key))
	if v == "" {
		return fallback
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return fallback
	}
	return n
}

func getEnvDuration(key string, fallback time.Duration) time.Duration {
	v := strings.TrimSpace(os.Getenv(key))
	if v == "" {
		return fallback
	}
	d, err := time.ParseDuration(v)
	if err != nil {
		return fallback
	}
	return d
}

func splitCSV(s string) []string {
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			out = append(out, part)
		}
	}
	return out
}
