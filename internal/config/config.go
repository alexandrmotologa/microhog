package config

import (
	"crypto/rand"
	"encoding/hex"
	"flag"
	"os"
	"strconv"
)

// Config holds runtime configuration for the MicroHog server.
type Config struct {
	Port          int
	DBPath        string
	SaltSecret    string
	RetentionDays int
}

// Load parses CLI flags and environment variables.
func Load() *Config {
	cfg := &Config{
		Port:          8080,
		DBPath:        "./microhog.db",
		RetentionDays: 90,
	}

	if envPort := os.Getenv("PORT"); envPort != "" {
		if p, err := strconv.Atoi(envPort); err == nil {
			cfg.Port = p
		}
	}
	if envDB := os.Getenv("MICROHOG_DB"); envDB != "" {
		cfg.DBPath = envDB
	}
	if envRetention := os.Getenv("MICROHOG_RETENTION_DAYS"); envRetention != "" {
		if r, err := strconv.Atoi(envRetention); err == nil {
			cfg.RetentionDays = r
		}
	}
	cfg.SaltSecret = os.Getenv("MICROHOG_SALT_SECRET")

	flag.IntVar(&cfg.Port, "port", cfg.Port, "HTTP server listening port")
	flag.StringVar(&cfg.DBPath, "db", cfg.DBPath, "Path to SQLite database file")
	flag.StringVar(&cfg.SaltSecret, "salt-secret", cfg.SaltSecret, "Secret key used for daily salt anonymization")
	flag.IntVar(&cfg.RetentionDays, "retention-days", cfg.RetentionDays, "Retention period in days for events and replays")

	// Only parse flags if not already parsed (e.g. in tests)
	if !flag.Parsed() {
		flag.Parse()
	}

	if cfg.SaltSecret == "" {
		cfg.SaltSecret = generateRandomSecret(32)
	}

	return cfg
}

func generateRandomSecret(length int) string {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "microhog_fallback_secret_salt_2026"
	}
	return hex.EncodeToString(bytes)
}
