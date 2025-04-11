package main

import (
	"fmt"
	"net/http"
	"os"

	"wallet-guesser/internal/api/websocket"
	"wallet-guesser/internal/avoidlist"
	"wallet-guesser/internal/blockchain"
	"wallet-guesser/internal/config"
	"wallet-guesser/internal/game"
	"wallet-guesser/internal/twitter"

	log "github.com/sirupsen/logrus"
)

func main() {
	// Configure logging
	log.SetOutput(os.Stdout)
	log.SetFormatter(&log.TextFormatter{
		FullTimestamp: true,
	})

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Set log level based on debug setting
	if cfg.Debug {
		log.SetLevel(log.DebugLevel)
	} else {
		log.SetLevel(log.InfoLevel)
	}
	log.Info("Starting Wallet Guesser server...")

	// Initialize avoid list service
	avoidListSvc := avoidlist.NewService(cfg.DuneApiKey, cfg.AvoidListPath)
	if err := avoidListSvc.LoadFromFile(); err != nil {
		log.Warnf("Could not load avoid list, will start with empty list: %v", err)
	} else {
		stats := avoidListSvc.GetAvoidListStats()
		log.Infof("Loaded avoid list with %d entries, last updated at %s",
			stats["totalEntries"], stats["lastUpdated"])
	}

	// Initialize Twitter client
	twitterClient := twitter.NewClient(
		twitter.WithApifyToken(cfg.ApifyToken),
	)

	// Initialize Blockchain client
	blockchainClient := blockchain.NewClient(cfg.SolanaRpcEndpoint, avoidListSvc)

	// Initialize the wallet guesser
	walletGuesser := game.NewWalletGuesser(twitterClient, blockchainClient, avoidListSvc)

	// Initialize API handlers
	wsHandler := websocket.NewHandler(walletGuesser)

	// Set up WebSocket endpoint
	http.HandleFunc("/ws", wsHandler.HandleWebSocket)

	// Set up CORS headers for development
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		// Serve static files or handle other routes as needed
		http.NotFound(w, r)
	})

	// Create status endpoint
	http.HandleFunc("/api/status", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status":"ok","version":"1.0.0"}`))
	})

	// Start the server
	serverAddr := fmt.Sprintf(":%d", cfg.Port)
	log.Infof("Server listening on %s", serverAddr)
	if err := http.ListenAndServe(serverAddr, nil); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
