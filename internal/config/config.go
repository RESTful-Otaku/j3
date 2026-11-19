// Package config loads venue-server settings from the environment.
package config

import (
	"os"
	"time"
)

// Config holds runtime settings for the venue server.
type Config struct {
	Addr     string
	DBPath   string
	Version  string
	SeedDemo bool
	Timeout  time.Duration
}

// Load reads configuration with sensible, non-technical defaults.
func Load() Config {
	return Config{
		Addr:     env("J3_ADDR", ":8787"),
		DBPath:   env("J3_DB", "./data/j3.db"),
		Version:  env("J3_VERSION", "dev"),
		SeedDemo: os.Getenv("J3_SEED_DEMO") == "1",
		Timeout:  15 * time.Second,
	}
}

func env(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
