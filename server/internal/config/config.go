package config

import (
	"os"
	"time"
)

type Config struct {
	// Server settings
	Port int

	// Database settings
	DatabasePath string

	// Scraper settings
	ScraperInterval time.Duration
	Timeout         time.Duration
	Delay           time.Duration
	UserAgent       string
}

// DefaultConfig returns default settings, with optional env overrides
func DefaultConfig() *Config {
	dbPath := os.Getenv("DATABASE_PATH")
	if dbPath == "" {
		dbPath = "./data/shiori.db"
	}

	return &Config{
		Port:            8080,
		DatabasePath:    dbPath,
		ScraperInterval: 5 * time.Minute,
		Timeout:         30 * time.Second,
		Delay:           2 * time.Second,
		UserAgent:       "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 Chrome/120.0.0.0 Safari/537.36",
	}
}
