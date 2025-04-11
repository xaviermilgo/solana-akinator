// cmd/server/main.go
package main

import (
	"fmt"
	"log"
	"net/http"

	"wallet-guesser/internal/api"
	"wallet-guesser/internal/blockchain"
	"wallet-guesser/internal/config"
	"wallet-guesser/internal/game"
	"wallet-guesser/internal/twitter"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize Twitter client
	twitterClient := twitter.NewClient(
		twitter.WithApifyToken(cfg.ApifyToken),
	)

	// Initialize Blockchain client
	blockchainClient := blockchain.NewClient(cfg.SolanaRpcEndpoint)

	// Initialize the wallet guesser
	walletGuesser := game.NewWalletGuesser(twitterClient, blockchainClient)

	// Initialize API handlers
	apiHandler := api.NewHandler(walletGuesser)

	// Set up WebSocket endpoint
	http.HandleFunc("/ws", apiHandler.HandleWebSocket)

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

	// Start the server
	serverAddr := fmt.Sprintf(":%d", cfg.Port)
	log.Printf("Starting Wallet Guesser server on %s", serverAddr)
	if err := http.ListenAndServe(serverAddr, nil); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
