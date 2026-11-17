package config

import (
	"strings"
	"testing"
	"time"
)

func TestLoadMongoSettings(t *testing.T) {
	setRequiredEnvironment(t)
	t.Setenv("APP_ENV", "development")
	t.Setenv("MONGODB_MAX_POOL_SIZE", "")
	t.Setenv("MONGODB_TIMEOUT_SECONDS", "")
	t.Setenv("SESSION_TTL_MINUTES", "")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.MongoMaxPoolSize != 100 {
		t.Fatalf("MongoMaxPoolSize = %d, want 100", cfg.MongoMaxPoolSize)
	}
	if cfg.MongoTimeout != 10*time.Second {
		t.Fatalf("MongoTimeout = %s, want 10s", cfg.MongoTimeout)
	}
}

func TestLoadRejectsInvalidMongoURIWithoutEchoingIt(t *testing.T) {
	setRequiredEnvironment(t)
	secretURI := "not-a-mongodb-uri"
	t.Setenv("MONGODB_URI", secretURI)
	t.Setenv("APP_ENV", "development")

	_, err := Load()
	if err == nil {
		t.Fatal("Load() accepted an invalid MongoDB URI")
	}
	if strings.Contains(err.Error(), secretURI) {
		t.Fatal("configuration error must not echo the connection string")
	}
}

func TestLoadRejectsPartialSMTPConfiguration(t *testing.T) {
	setRequiredEnvironment(t)
	t.Setenv("APP_ENV", "development")
	t.Setenv("SMTP_HOST", "smtp.example.com")
	t.Setenv("SMTP_PORT", "")
	t.Setenv("SMTP_USERNAME", "")
	t.Setenv("SMTP_PASSWORD", "")
	t.Setenv("SMTP_FROM", "")

	if _, err := Load(); err == nil {
		t.Fatal("Load() accepted incomplete SMTP configuration")
	}
}

func TestProductionRequiresHTTPSAndSMTPAndDoesNotAllowLocalOrigin(t *testing.T) {
	setRequiredEnvironment(t)
	t.Setenv("APP_ENV", "production")
	t.Setenv("SESSION_SECRET", "0123456789abcdef0123456789abcdef")
	t.Setenv("APP_BASE_URL", "https://planner.example.com")
	t.Setenv("ALLOWED_ORIGINS", "")
	t.Setenv("SMTP_HOST", "smtp.example.com")
	t.Setenv("SMTP_PORT", "587")
	t.Setenv("SMTP_USERNAME", "mailer")
	t.Setenv("SMTP_PASSWORD", "secret")
	t.Setenv("SMTP_FROM", "Year Planner <no-reply@example.com>")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if len(cfg.AllowedOrigins) != 1 || cfg.AllowedOrigins[0] != "https://planner.example.com" {
		t.Fatalf("production origins = %#v; want only configured HTTPS origin", cfg.AllowedOrigins)
	}

	t.Setenv("APP_BASE_URL", "http://planner.example.com")
	if _, err := Load(); err == nil {
		t.Fatal("Load() accepted an HTTP APP_BASE_URL in production")
	}
}

func setRequiredEnvironment(t *testing.T) {
	t.Helper()
	t.Setenv("SESSION_SECRET", "test-secret")
	t.Setenv("MONGODB_URI", "mongodb://localhost:27018/?authSource=admin")
	t.Setenv("MONGODB_DATABASE", "year_planner_test")
	t.Setenv("ALLOWED_ORIGINS", "")
	t.Setenv("APP_BASE_URL", "http://localhost:5173")
	t.Setenv("SMTP_HOST", "")
	t.Setenv("SMTP_PORT", "")
	t.Setenv("SMTP_USERNAME", "")
	t.Setenv("SMTP_PASSWORD", "")
	t.Setenv("SMTP_FROM", "")
}
