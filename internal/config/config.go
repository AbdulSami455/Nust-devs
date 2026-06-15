package config

import (
	"fmt"
	"os"
	"strings"
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
