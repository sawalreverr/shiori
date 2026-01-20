package config

import (
	"os"
	"strconv"
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

// DefaultConfig returns default settings
func DefaultConfig() *Config {
	dbPath := getEnv("DATABASE_PATH", "./data/shiori.db")
	port := getEnvInt("PORT", 8080)
	scrapeInterval := getEnvDuration("SCRAPE_INTERVAL", 5*time.Minute)
	timeout := getEnvDuration("TIMEOUT", 30*time.Second)
	delay := getEnvDuration("DELAY", 2*time.Second)
	userAgent := getEnv("USER_AGENT", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 Chrome/120.0.0.0 Safari/537.36")

	return &Config{
		Port:            port,
		DatabasePath:    dbPath,
		ScraperInterval: scrapeInterval,
		Timeout:         timeout,
		Delay:           delay,
		UserAgent:       userAgent,
	}
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	if value, ok := os.LookupEnv(key); ok {
		if i, err := strconv.Atoi(value); err == nil {
			return i
		}
	}
	return fallback
}

func getEnvDuration(key string, fallback time.Duration) time.Duration {
	if value, ok := os.LookupEnv(key); ok {
		if d, err := time.ParseDuration(value); err == nil {
			return d
		}
	}
	return fallback
}
