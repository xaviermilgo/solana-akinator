package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/joho/godotenv"
	log "github.com/sirupsen/logrus"
	"wallet-guesser/internal/avoidlist"
)

func main() {
	// Parse command line arguments
	var outputFile string
	var verbose bool

	flag.StringVar(&outputFile, "output", "", "Path to output file (default: data/avoidlist.json)")
	flag.BoolVar(&verbose, "verbose", false, "Enable verbose logging")
	flag.Parse()

	// Configure logging
	log.SetOutput(os.Stdout)
	if verbose {
		log.SetLevel(log.DebugLevel)
	} else {
		log.SetLevel(log.InfoLevel)
	}

	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.WithError(err).Warnf("Could not load environment file")
	}

	// Get API key from environment
	apiKey := os.Getenv("DUNE_API_KEY")
	if apiKey == "" {
		log.Fatal("DUNE_API_KEY environment variable not set")
	}

	// Create avoid list service
	service := avoidlist.NewService(apiKey, outputFile)

	// Try to load existing data first
	if err := service.LoadFromFile(); err != nil {
		log.Warnf("Could not load existing avoid list: %v", err)
	}

	// Update the avoid list
	log.Info("Updating avoid list...")
	if err := service.UpdateAvoidList(); err != nil {
		log.Fatalf("Failed to update avoid list: %v", err)
	}

	// Print stats
	stats := service.GetAvoidListStats()
	fmt.Printf("Avoid list updated successfully!\n")
	fmt.Printf("  Total entries: %d\n", stats["totalEntries"])
	fmt.Printf("  Token entries: %d\n", stats["tokenCount"])
	fmt.Printf("  Wallet entries: %d\n", stats["walletCount"])
	fmt.Printf("  Last updated: %s\n", stats["lastUpdated"])
}
