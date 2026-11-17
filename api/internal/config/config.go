package config

import (
	"errors"
	"net/mail"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	AppEnv            string
	AppPort           string
	AppBaseURL        string
	AllowedOrigins    []string
	SessionSecret     string
	SessionTTLMinutes int
	MongoURI          string
	MongoDatabase     string
	MongoMaxPoolSize  uint64
	MongoTimeout      time.Duration
	SMTPHost          string
	SMTPPort          string
	SMTPUsername      string
	SMTPPassword      string
	SMTPFrom          string
}

func Load() (Config, error) {
	cfg := Config{
		AppEnv:            getEnv("APP_ENV", "development"),
		AppPort:           getEnv("APP_PORT", "8080"),
		AppBaseURL:        getEnv("APP_BASE_URL", "http://localhost:5173"),
		SessionSecret:     getEnv("SESSION_SECRET", ""),
		SessionTTLMinutes: 60 * 24,
		MongoURI:          getEnv("MONGODB_URI", ""),
		MongoDatabase:     getEnv("MONGODB_DATABASE", ""),
		MongoMaxPoolSize:  100,
		MongoTimeout:      10 * time.Second,
		SMTPHost:          strings.TrimSpace(os.Getenv("SMTP_HOST")),
		SMTPPort:          strings.TrimSpace(os.Getenv("SMTP_PORT")),
		SMTPUsername:      os.Getenv("SMTP_USERNAME"),
		SMTPPassword:      os.Getenv("SMTP_PASSWORD"),
		SMTPFrom:          strings.TrimSpace(os.Getenv("SMTP_FROM")),
	}

	if cfg.SessionSecret == "" {
		return Config{}, errors.New("SESSION_SECRET is required")
	}
	if cfg.AppEnv == "production" && len([]byte(strings.TrimSpace(cfg.SessionSecret))) < 32 {
		return Config{}, errors.New("SESSION_SECRET must contain at least 32 bytes in production")
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
	if !isOrigin(cfg.AppBaseURL) {
		return Config{}, errors.New("APP_BASE_URL must be a valid origin without a path")
	}
	if cfg.AppEnv == "production" && !strings.HasPrefix(cfg.AppBaseURL, "https://") {
		return Config{}, errors.New("APP_BASE_URL must use HTTPS in production")
	}
	if err := setAllowedOrigins(&cfg); err != nil {
		return Config{}, err
	}
	if err := validateSMTP(&cfg); err != nil {
		return Config{}, err
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

func validateSMTP(cfg *Config) error {
	hasSMTPSetting := cfg.SMTPHost != "" || cfg.SMTPPort != "" ||
		cfg.SMTPUsername != "" || cfg.SMTPPassword != "" || cfg.SMTPFrom != ""
	if cfg.AppEnv == "production" && !hasSMTPSetting {
		return errors.New("SMTP_HOST, SMTP_PORT, and SMTP_FROM are required in production")
	}
	if !hasSMTPSetting {
		return nil
	}
	if cfg.SMTPHost == "" || cfg.SMTPPort == "" || cfg.SMTPFrom == "" {
		return errors.New("SMTP_HOST, SMTP_PORT, and SMTP_FROM must be configured together")
	}
	port, err := strconv.Atoi(cfg.SMTPPort)
	if err != nil || port < 1 || port > 65535 {
		return errors.New("SMTP_PORT must be an integer between 1 and 65535")
	}
	if (cfg.SMTPUsername == "") != (cfg.SMTPPassword == "") {
		return errors.New("SMTP_USERNAME and SMTP_PASSWORD must be configured together")
	}
	if _, err := mail.ParseAddress(cfg.SMTPFrom); err != nil {
		return errors.New("SMTP_FROM must be a valid email address")
	}
	return nil
}

func setAllowedOrigins(cfg *Config) error {
	defaultOrigins := cfg.AppBaseURL
	if cfg.AppEnv != "production" {
		defaultOrigins += ",http://localhost:" + cfg.AppPort
	}
	rawOrigins := getEnv("ALLOWED_ORIGINS", defaultOrigins)
	seen := make(map[string]struct{})
	for _, rawOrigin := range strings.Split(rawOrigins, ",") {
		origin := strings.TrimSpace(rawOrigin)
		if !isOrigin(origin) {
			return errors.New("ALLOWED_ORIGINS must contain valid origins without paths")
		}
		if _, exists := seen[origin]; !exists {
			cfg.AllowedOrigins = append(cfg.AllowedOrigins, origin)
			seen[origin] = struct{}{}
		}
	}
	return nil
}

func isOrigin(origin string) bool {
	parsed, err := url.Parse(origin)
	return err == nil &&
		(parsed.Scheme == "http" || parsed.Scheme == "https") &&
		parsed.Host != "" &&
		parsed.User == nil &&
		parsed.Path == "" &&
		parsed.RawQuery == "" &&
		parsed.Fragment == ""
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok && value != "" {
		return value
	}
	return fallback
}
