// Package config loads and validates all configuration from environment variables.
// SECURITY: No default values for secrets. The app refuses to start without them.
package config

import (
	"errors"
	"os"
	"strconv"
	"strings"
)

// Config holds all application configuration loaded from environment variables.
type Config struct {
	// Server
	Port string
	Host string
	Env  string // development | staging | production

	// Security
	SecretKey           string
	MemoryEncryptionKey string

	// CORS
	CORSAllowedOrigins []string

	// Rate limiting
	RateLimitRPM int

	// Providers
	OpenAIAPIKey    string
	AnthropicAPIKey string
	OllamaHost      string

	// Database
	DatabaseURL string
}

// Load reads configuration from environment variables and validates required fields.
func Load() (*Config, error) {
	cfg := &Config{
		Port:         getEnvOrDefault("API_PORT", "8080"),
		Host:         getEnvOrDefault("API_HOST", "0.0.0.0"),
		Env:          getEnvOrDefault("API_ENV", "development"),
		SecretKey:    os.Getenv("API_SECRET_KEY"),
		MemoryEncryptionKey: os.Getenv("MEMORY_ENCRYPTION_KEY"),
		OpenAIAPIKey:    os.Getenv("OPENAI_API_KEY"),
		AnthropicAPIKey: os.Getenv("ANTHROPIC_API_KEY"),
		OllamaHost:      getEnvOrDefault("OLLAMA_HOST", "http://localhost:11434"),
		DatabaseURL:     os.Getenv("DATABASE_URL"),
		RateLimitRPM:    getEnvAsInt("RATE_LIMIT_RPM", 60),
	}

	// Parse CORS origins
	corsRaw := getEnvOrDefault("CORS_ALLOWED_ORIGINS", "http://localhost:3000")
	cfg.CORSAllowedOrigins = strings.Split(corsRaw, ",")
	for i := range cfg.CORSAllowedOrigins {
		cfg.CORSAllowedOrigins[i] = strings.TrimSpace(cfg.CORSAllowedOrigins[i])
	}

	if err := cfg.validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

// validate checks that required fields are present.
func (c *Config) validate() error {
	if c.Env == "production" {
		if c.SecretKey == "" {
			return errors.New("API_SECRET_KEY is required in production")
		}
		if len(c.SecretKey) < 32 {
			return errors.New("API_SECRET_KEY must be at least 32 characters")
		}
		if c.MemoryEncryptionKey == "" {
			return errors.New("MEMORY_ENCRYPTION_KEY is required in production")
		}
		if len(c.MemoryEncryptionKey) < 32 {
			return errors.New("MEMORY_ENCRYPTION_KEY must be at least 32 characters")
		}
	}
	return nil
}

// IsProduction returns true if the app is running in production mode.
func (c *Config) IsProduction() bool {
	return c.Env == "production"
}

func getEnvOrDefault(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}

func getEnvAsInt(key string, defaultVal int) int {
	val := os.Getenv(key)
	if val == "" {
		return defaultVal
	}
	n, err := strconv.Atoi(val)
	if err != nil {
		return defaultVal
	}
	return n
}
