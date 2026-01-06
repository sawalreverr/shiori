package config

import "time"

type Config struct {
	// Server settings
	Port int

	// Scraper settings
	ScraperInterval time.Duration
	Timeout         time.Duration
	Delay           time.Duration
	UserAgent       string
}

// default settings
func DefaultConfig() *Config {
	return &Config{
		Port:            8080,
		ScraperInterval: 5 * time.Minute,
		Timeout:         30 * time.Second,
		Delay:           2 * time.Second,
		UserAgent:       "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 Chrome/120.0.0.0 Safari/537.36",
	}
}
