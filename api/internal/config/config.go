package config

import (
	"errors"
	"net/url"
	"os"
	"strconv"
	"time"
)

type Config struct {
	AppEnv            string
	AppPort           string
	SessionSecret     string
	SessionTTLMinutes int
	MongoURI          string
	MongoDatabase     string
	MongoMaxPoolSize  uint64
	MongoTimeout      time.Duration
}

func Load() (Config, error) {
	cfg := Config{
		AppEnv:            getEnv("APP_ENV", "development"),
		AppPort:           getEnv("APP_PORT", "8080"),
		SessionSecret:     getEnv("SESSION_SECRET", ""),
		SessionTTLMinutes: 60 * 24,
		MongoURI:          getEnv("MONGODB_URI", ""),
		MongoDatabase:     getEnv("MONGODB_DATABASE", ""),
		MongoMaxPoolSize:  100,
		MongoTimeout:      10 * time.Second,
	}

	if cfg.SessionSecret == "" {
		return Config{}, errors.New("SESSION_SECRET is required")
	}
	if cfg.MongoURI == "" {
		return Config{}, errors.New("MONGODB_URI is required")
	}
	parsedURI, err := url.Parse(cfg.MongoURI)
	if err != nil || (parsedURI.Scheme != "mongodb" && parsedURI.Scheme != "mongodb+srv") || parsedURI.Host == "" {
		return Config{}, errors.New("MONGODB_URI must be a valid MongoDB connection string")
	}
	if cfg.MongoDatabase == "" {
		return Config{}, errors.New("MONGODB_DATABASE is required")
	}

	if raw := os.Getenv("SESSION_TTL_MINUTES"); raw != "" {
		minutes, err := strconv.Atoi(raw)
		if err != nil || minutes <= 0 {
			return Config{}, errors.New("SESSION_TTL_MINUTES must be a positive integer")
		}
		cfg.SessionTTLMinutes = minutes
	}

	if raw := os.Getenv("MONGODB_MAX_POOL_SIZE"); raw != "" {
		poolSize, err := strconv.ParseUint(raw, 10, 64)
		if err != nil || poolSize == 0 {
			return Config{}, errors.New("MONGODB_MAX_POOL_SIZE must be a positive integer")
		}
		cfg.MongoMaxPoolSize = poolSize
	}

	if raw := os.Getenv("MONGODB_TIMEOUT_SECONDS"); raw != "" {
		seconds, err := strconv.Atoi(raw)
		if err != nil || seconds <= 0 {
			return Config{}, errors.New("MONGODB_TIMEOUT_SECONDS must be a positive integer")
		}
		cfg.MongoTimeout = time.Duration(seconds) * time.Second
	}

	return cfg, nil
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok && value != "" {
		return value
	}
	return fallback
}
