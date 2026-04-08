package config

import (
	"os"
	"testing"
)

func TestLoad_Defaults(t *testing.T) {
	// Clear env to test defaults
	os.Unsetenv("API_PORT")
	os.Unsetenv("API_HOST")
	os.Unsetenv("API_ENV")
	os.Unsetenv("API_SECRET_KEY")
	os.Unsetenv("CORS_ALLOWED_ORIGINS")
	os.Unsetenv("RATE_LIMIT_RPM")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.Port != "8080" {
		t.Errorf("expected port 8080, got %s", cfg.Port)
	}
	if cfg.Host != "0.0.0.0" {
		t.Errorf("expected host 0.0.0.0, got %s", cfg.Host)
	}
	if cfg.Env != "development" {
		t.Errorf("expected env development, got %s", cfg.Env)
	}
	if cfg.RateLimitRPM != 60 {
		t.Errorf("expected rate limit 60, got %d", cfg.RateLimitRPM)
	}
	if len(cfg.CORSAllowedOrigins) != 3 || cfg.CORSAllowedOrigins[0] != "http://localhost:3000" {
		t.Errorf("unexpected CORS origins: %v", cfg.CORSAllowedOrigins)
	}
}

func TestLoad_CustomValues(t *testing.T) {
	t.Setenv("API_PORT", "9090")
	t.Setenv("API_HOST", "127.0.0.1")
	t.Setenv("API_ENV", "staging")
	t.Setenv("RATE_LIMIT_RPM", "120")
	t.Setenv("CORS_ALLOWED_ORIGINS", "https://example.com, https://app.example.com")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.Port != "9090" {
		t.Errorf("expected port 9090, got %s", cfg.Port)
	}
	if cfg.Host != "127.0.0.1" {
		t.Errorf("expected host 127.0.0.1, got %s", cfg.Host)
	}
	if cfg.RateLimitRPM != 120 {
		t.Errorf("expected rate limit 120, got %d", cfg.RateLimitRPM)
	}
	if len(cfg.CORSAllowedOrigins) != 2 {
		t.Fatalf("expected 2 CORS origins, got %d", len(cfg.CORSAllowedOrigins))
	}
	if cfg.CORSAllowedOrigins[0] != "https://example.com" {
		t.Errorf("expected first origin https://example.com, got %s", cfg.CORSAllowedOrigins[0])
	}
	if cfg.CORSAllowedOrigins[1] != "https://app.example.com" {
		t.Errorf("expected second origin https://app.example.com, got %s", cfg.CORSAllowedOrigins[1])
	}
}

func TestLoad_ProductionRequiresSecrets(t *testing.T) {
	t.Setenv("API_ENV", "production")
	t.Setenv("API_SECRET_KEY", "")
	t.Setenv("MEMORY_ENCRYPTION_KEY", "")

	_, err := Load()
	if err == nil {
		t.Fatal("expected error for production without secrets")
	}
}

func TestLoad_ProductionShortSecret(t *testing.T) {
	t.Setenv("API_ENV", "production")
	t.Setenv("API_SECRET_KEY", "short")
	t.Setenv("MEMORY_ENCRYPTION_KEY", "")

	_, err := Load()
	if err == nil {
		t.Fatal("expected error for short secret key")
	}
}

func TestLoad_ProductionValid(t *testing.T) {
	t.Setenv("API_ENV", "production")
	t.Setenv("API_SECRET_KEY", "abcdefghijklmnopqrstuvwxyz123456")
	t.Setenv("MEMORY_ENCRYPTION_KEY", "abcdefghijklmnopqrstuvwxyz1234567890abcdefghijklmnopqrstuvwxyz1234")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !cfg.IsProduction() {
		t.Error("expected IsProduction to return true")
	}
}

func TestIsProduction(t *testing.T) {
	tests := []struct {
		env  string
		want bool
	}{
		{"production", true},
		{"development", false},
		{"staging", false},
		{"", false},
	}

	for _, tt := range tests {
		c := &Config{Env: tt.env}
		if got := c.IsProduction(); got != tt.want {
			t.Errorf("IsProduction(%q) = %v, want %v", tt.env, got, tt.want)
		}
	}
}

func TestGetEnvAsInt_Invalid(t *testing.T) {
	t.Setenv("RATE_LIMIT_RPM", "not-a-number")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should fall back to default
	if cfg.RateLimitRPM != 60 {
		t.Errorf("expected default 60 for invalid int, got %d", cfg.RateLimitRPM)
	}
}
