package game

import (
	"fmt"
	"log"
	"sort"
	"strings"
	"sync"
	"time"

	"wallet-guesser/internal/blockchain"
	"wallet-guesser/internal/twitter"
)

// TokenInfo represents information about a token project
type TokenInfo struct {
	Symbol      string
	Name        string
	MintAddress string
}

// WalletGuessResult represents the result of the wallet guessing process
type WalletGuessResult struct {
	TwitterHandle string   `json:"twitterHandle"`
	Addresses     []string `json:"addresses"`
	Sources       []string `json:"sources"`
	Confidence    int      `json:"confidence"` // 0-100
}

// WalletGuesser handles the process of guessing wallet addresses
type WalletGuesser struct {
	twitterClient    *twitter.Client
	blockchainClient *blockchain.Client
	cacheMutex       sync.RWMutex
	resultCache      map[string]*WalletGuessResult
	tokenCache       map[string]TokenInfo // Cache of token info by Twitter handle
}

// NewWalletGuesser creates a new WalletGuesser
func NewWalletGuesser(twitterClient *twitter.Client, blockchainClient *blockchain.Client) *WalletGuesser {
	return &WalletGuesser{
		twitterClient:    twitterClient,
		blockchainClient: blockchainClient,
		resultCache:      make(map[string]*WalletGuessResult),
		tokenCache:       make(map[string]TokenInfo),
	}
}

// GuessWallet tries to guess the wallet address for a given Twitter handle
func (wg *WalletGuesser) GuessWallet(twitterHandle string, progressCallback func(message string)) (*WalletGuessResult, error) {
	// Clean the Twitter handle (remove @ if present)
	twitterHandle = strings.TrimPrefix(twitterHandle, "@")

	// Check cache first
	wg.cacheMutex.RLock()
	if result, found := wg.resultCache[twitterHandle]; found {
		wg.cacheMutex.RUnlock()
		if progressCallback != nil {
			progressCallback(fmt.Sprintf("Using cached results for @%s", twitterHandle))
		}
		return result, nil
	}
	wg.cacheMutex.RUnlock()

	if progressCallback != nil {
		progressCallback(fmt.Sprintf("The Jinn is analyzing @%s's Twitter profile...", twitterHandle))
	}

	// Initialize the result
	result := &WalletGuessResult{
		TwitterHandle: twitterHandle,
		Addresses:     []string{},
		Sources:       []string{},
		Confidence:    0,
	}

	// Step 1: Fetch accounts the user follows to identify token projects
	if progressCallback != nil {
		progressCallback(fmt.Sprintf("Examining Twitter accounts followed by @%s...", twitterHandle))
	}

	following, err := wg.twitterClient.FetchFollowing(twitterHandle, 500, progressCallback)
	if err != nil {
		log.Printf("Error fetching following for %s: %v", twitterHandle, err)
		return nil, fmt.Errorf("failed to fetch accounts followed by @%s: %w", twitterHandle, err)
	}

	// Step 2: Extract potential token mint addresses from the followed accounts
	var tokenMints []string
	tokenToSourceMap := make(map[string]string) // Maps token mint to Twitter handle

	if progressCallback != nil {
		progressCallback(fmt.Sprintf("Analyzing %d accounts to identify token projects...", len(following)))
	}

	// Loop through followed accounts to find token projects
	for _, user := range following {
		for _, mint := range user.PossibleMintAddresses {
			tokenMints = append(tokenMints, mint)
			tokenToSourceMap[mint] = fmt.Sprintf("@%s", user.Username)

			if progressCallback != nil {
				progressCallback(fmt.Sprintf("Found potential token mint address from @%s", user.Username))
			}
		}
	}

	if len(tokenMints) == 0 {
		if progressCallback != nil {
			progressCallback("No token projects identified from followed accounts")
		}

		// Cache the empty result to avoid repeated lookups
		wg.cacheMutex.Lock()
		wg.resultCache[twitterHandle] = result
		wg.cacheMutex.Unlock()

		return result, nil
	}

	if progressCallback != nil {
		progressCallback(fmt.Sprintf("Found %d potential token projects to analyze", len(tokenMints)))
	}

	// Step 3: For each token mint, get wallets that have interacted with it
	walletScores := make(map[string]int)
	walletToTokens := make(map[string][]string)

	for _, mintAddress := range tokenMints {
		if progressCallback != nil {
			progressCallback(fmt.Sprintf("Checking blockchain for wallets that interacted with token %s...", mintAddress))
		}

		// Get all wallets that have interacted with this token
		wallets, err := wg.blockchainClient.GetWalletsForToken(mintAddress, progressCallback)
		if err != nil {
			log.Printf("Error getting wallets for token %s: %v", mintAddress, err)
			continue
		}

		for _, wallet := range wallets {
			walletScores[wallet]++

			// Track which tokens this wallet has interacted with
			if _, exists := walletToTokens[wallet]; !exists {
				walletToTokens[wallet] = []string{mintAddress}
			} else {
				walletToTokens[wallet] = append(walletToTokens[wallet], mintAddress)
			}
		}
	}

	// Step 4: Rank wallets by score and compile results
	type WalletScore struct {
		Address string
		Score   int
	}

	var rankedWallets []WalletScore
	for addr, score := range walletScores {
		rankedWallets = append(rankedWallets, WalletScore{addr, score})
	}

	// Sort wallets by score (highest first)
	sort.Slice(rankedWallets, func(i, j int) bool {
		return rankedWallets[i].Score > rankedWallets[j].Score
	})

	// Take top results (limit to 5)
	maxResults := 5
	if len(rankedWallets) < maxResults {
		maxResults = len(rankedWallets)
	}

	for i := 0; i < maxResults; i++ {
		wallet := rankedWallets[i]
		result.Addresses = append(result.Addresses, wallet.Address)

		// Create source description
		tokens := walletToTokens[wallet.Address]
		sourceTokens := make([]string, 0)

		for _, token := range tokens {
			if source, exists := tokenToSourceMap[token]; exists {
				sourceTokens = append(sourceTokens, source)
			}
		}

		// Deduplicate sources
		uniqueSources := make(map[string]bool)
		for _, src := range sourceTokens {
			uniqueSources[src] = true
		}

		var sourcesList []string
		for src := range uniqueSources {
			sourcesList = append(sourcesList, src)
		}

		sourceText := fmt.Sprintf("Match score: %d/100. Matched tokens from: %s",
			calculateConfidence(wallet.Score, len(tokenMints)),
			strings.Join(sourcesList, ", "))

		result.Sources = append(result.Sources, sourceText)
	}

	// Set overall confidence based on highest match
	if len(rankedWallets) > 0 {
		result.Confidence = calculateConfidence(rankedWallets[0].Score, len(tokenMints))
	}

	// Cache the result
	wg.cacheMutex.Lock()
	wg.resultCache[twitterHandle] = result
	wg.cacheMutex.Unlock()

	if progressCallback != nil {
		if len(result.Addresses) > 0 {
			progressCallback(fmt.Sprintf("Analysis complete. Found %d potential wallet addresses.", len(result.Addresses)))
		} else {
			progressCallback("Analysis complete. No strong wallet matches found.")
		}
	}

	return result, nil
}

// calculateConfidence returns a confidence score (0-100) based on matches
func calculateConfidence(matches int, totalTokens int) int {
	if totalTokens == 0 {
		return 0
	}

	// Base confidence on percentage of tokens matched, with a curve
	rawPercent := (float64(matches) / float64(totalTokens)) * 100

	// Apply a curve to make it harder to get very high confidence
	if rawPercent >= 75 {
		return 90 + int((rawPercent-75)/25)*10 // 75% -> 90%, 100% -> 100%
	} else if rawPercent >= 50 {
		return 70 + int((rawPercent-50)/25)*20 // 50% -> 70%, 75% -> 90%
	} else if rawPercent >= 25 {
		return 40 + int((rawPercent-25)/25)*30 // 25% -> 40%, 50% -> 70%
	} else {
		return int(rawPercent) + 15 // 0% -> 15%, 25% -> 40%
	}
}

// contains checks if a string is in a slice
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// ClearCache clears the cache
func (wg *WalletGuesser) ClearCache() {
	wg.cacheMutex.Lock()
	defer wg.cacheMutex.Unlock()
	wg.resultCache = make(map[string]*WalletGuessResult)
	wg.tokenCache = make(map[string]TokenInfo)
}

// CacheStats returns statistics about the cache
func (wg *WalletGuesser) CacheStats() map[string]interface{} {
	wg.cacheMutex.RLock()
	defer wg.cacheMutex.RUnlock()

	return map[string]interface{}{
		"resultsSize": len(wg.resultCache),
		"tokensSize":  len(wg.tokenCache),
		"lastUpdated": time.Now().Format(time.RFC3339),
	}
}
