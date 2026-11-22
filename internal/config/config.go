// Package config loads venue-server settings from the environment.
package config

import (
	"os"
	"time"
)

// Config holds runtime settings for the venue server.
type Config struct {
	Addr          string
	DBPath        string
	Version       string
	SeedDemo      bool
	Timeout       time.Duration
	OwnerEmail    string
	OwnerName     string
	OwnerPassword string
}

// Load reads configuration with sensible, non-technical defaults.
func Load() Config {
	return Config{
		Addr:          env("J3_ADDR", "127.0.0.1:8787"),
		DBPath:        env("J3_DB", "./data/j3.db"),
		Version:       env("J3_VERSION", "dev"),
		SeedDemo:      os.Getenv("J3_SEED_DEMO") == "1",
		Timeout:       15 * time.Second,
		OwnerEmail:    env("J3_OWNER_EMAIL", "j3@local"),
		OwnerName:     env("J3_OWNER_NAME", "J3 Owner"),
		OwnerPassword: os.Getenv("J3_OWNER_PASSWORD"),
	}
}

func env(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
