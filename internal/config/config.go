package config

import (
	"os"
	"strconv"
)

// Config holds all configuration for the application
type Config struct {
	Port             int
	TwitterAPIKey    string
	TwitterAPISecret string
}

// Load loads configuration from environment variables
func Load() (*Config, error) {
	port := 8080 // Default port
	if portStr := os.Getenv("PORT"); portStr != "" {
		if p, err := strconv.Atoi(portStr); err == nil {
			port = p
		}
	}

	return &Config{
		Port:             port,
		TwitterAPIKey:    os.Getenv("TWITTER_API_KEY"),
		TwitterAPISecret: os.Getenv("TWITTER_API_SECRET"),
	}, nil
}
