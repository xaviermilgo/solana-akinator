package config

import (
	"os"
	"strconv"

	"github.com/joho/godotenv"
	log "github.com/sirupsen/logrus"
)

// Config holds all configuration for the application
type Config struct {
	Port              int
	ApifyToken        string
	SolanaRpcEndpoint string
	DuneApiKey        string
	AvoidListPath     string
	Debug             bool
}

// Load loads configuration from environment variables
func Load() (*Config, error) {
	// Try to load .env file, but continue if it doesn't exist
	if err := godotenv.Load(); err != nil {
		log.Warnf("Error loading .env file: %v", err)
	}

	port := 8080 // Default port
	if portStr := os.Getenv("PORT"); portStr != "" {
		if p, err := strconv.Atoi(portStr); err == nil {
			port = p
		}
	}

	solanaRpcEndpoint := os.Getenv("SOLANA_RPC_ENDPOINT")
	if solanaRpcEndpoint == "" {
		log.Fatalf("SOLANA_RPC_ENDPOINT environment variable not set")
	}

	// Debug mode
	debug := false
	if debugStr := os.Getenv("DEBUG"); debugStr == "true" {
		debug = true
		log.SetLevel(log.DebugLevel)
	} else {
		log.SetLevel(log.InfoLevel)
	}

	// Avoid list path
	avoidListPath := os.Getenv("AVOID_LIST_PATH")
	if avoidListPath == "" {
		avoidListPath = "data/avoidlist.json"
	}

	return &Config{
		Port:              port,
		ApifyToken:        os.Getenv("APIFY_TOKEN"),
		SolanaRpcEndpoint: solanaRpcEndpoint,
		DuneApiKey:        os.Getenv("DUNE_API_KEY"),
		AvoidListPath:     avoidListPath,
		Debug:             debug,
	}, nil
}
