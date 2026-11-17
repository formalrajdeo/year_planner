package config

import (
	"strings"
	"testing"
	"time"
)

func TestLoadMongoSettings(t *testing.T) {
	t.Setenv("SESSION_SECRET", "test-secret")
	t.Setenv("MONGODB_URI", "mongodb://localhost:27018/?authSource=admin")
	t.Setenv("MONGODB_DATABASE", "year_planner_test")
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
	secretURI := "not-a-mongodb-uri"
	t.Setenv("SESSION_SECRET", "test-secret")
	t.Setenv("MONGODB_URI", secretURI)
	t.Setenv("MONGODB_DATABASE", "year_planner_test")

	_, err := Load()
	if err == nil {
		t.Fatal("Load() accepted an invalid MongoDB URI")
	}
	if strings.Contains(err.Error(), secretURI) {
		t.Fatal("configuration error must not echo the connection string")
	}
}
