package game

import (
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"wallet-guesser/internal/twitter"
)

// WalletGuessResult represents the result of the wallet guessing process
type WalletGuessResult struct {
	TwitterHandle string   `json:"twitterHandle"`
	Addresses     []string `json:"addresses"`
	Sources       []string `json:"sources"`
	Confidence    int      `json:"confidence"` // 0-100
}

// WalletGuesser handles the process of guessing wallet addresses
type WalletGuesser struct {
	twitterClient *twitter.Client
	cacheMutex    sync.RWMutex
	resultCache   map[string]*WalletGuessResult
}

// NewWalletGuesser creates a new WalletGuesser
func NewWalletGuesser(twitterClient *twitter.Client) *WalletGuesser {
	return &WalletGuesser{
		twitterClient: twitterClient,
		resultCache:   make(map[string]*WalletGuessResult),
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

	// Step 2: Fetch users the target follows to find potential wallet connections
	if progressCallback != nil {
		progressCallback(fmt.Sprintf("The Jinn is now examining who @%s follows...", twitterHandle))
	}

	following, err := wg.twitterClient.FetchFollowing(twitterHandle, 250, progressCallback)
	if err != nil {
		log.Printf("Error fetching following for %s: %v", twitterHandle, err)
		// Continue anyway, we might have found addresses in the user's bio
	}

	if progressCallback != nil {
		progressCallback(fmt.Sprintf("The Jinn has found %d accounts...", len(following)))
	}

	// Cache the result
	wg.cacheMutex.Lock()
	wg.resultCache[twitterHandle] = result
	wg.cacheMutex.Unlock()

	return result, nil
}

// ClearCache clears the cache
func (wg *WalletGuesser) ClearCache() {
	wg.cacheMutex.Lock()
	defer wg.cacheMutex.Unlock()
	wg.resultCache = make(map[string]*WalletGuessResult)
}

// CacheStats returns statistics about the cache
func (wg *WalletGuesser) CacheStats() map[string]interface{} {
	wg.cacheMutex.RLock()
	defer wg.cacheMutex.RUnlock()

	return map[string]interface{}{
		"size":        len(wg.resultCache),
		"lastUpdated": time.Now().Format(time.RFC3339),
	}
}
