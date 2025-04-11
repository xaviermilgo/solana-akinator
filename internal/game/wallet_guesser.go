package game

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"wallet-guesser/internal/domain"

	log "github.com/sirupsen/logrus"
)

// WalletGuesser implements domain.WalletGuesserService
type WalletGuesser struct {
	twitterClient    domain.TwitterService
	blockchainClient domain.BlockchainService
	avoidListService domain.AvoidListService
	cacheMutex       sync.RWMutex
	resultCache      map[string]*domain.WalletGuessResult
	tokenCache       map[string]domain.TokenInfo
}

// NewWalletGuesser creates a new WalletGuesser
func NewWalletGuesser(
	twitterClient domain.TwitterService,
	blockchainClient domain.BlockchainService,
	avoidListService domain.AvoidListService,
) *WalletGuesser {
	return &WalletGuesser{
		twitterClient:    twitterClient,
		blockchainClient: blockchainClient,
		avoidListService: avoidListService,
		resultCache:      make(map[string]*domain.WalletGuessResult),
		tokenCache:       make(map[string]domain.TokenInfo),
	}
}

// GuessWallet tries to guess the wallet address for a given Twitter handle
func (wg *WalletGuesser) GuessWallet(twitterHandle string, progressCallback domain.ProgressCallback) (*domain.WalletGuessResult, error) {
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
	result := &domain.WalletGuessResult{
		TwitterHandle: twitterHandle,
		Addresses:     []string{},
		Sources:       []string{},
		Confidence:    0,
	}

	// Fetch accounts the user follows
	following, err := wg.twitterClient.FetchFollowing(twitterHandle, 500, progressCallback)
	if err != nil {
		log.Errorf("Error fetching following for %s: %v", twitterHandle, err)
		return nil, fmt.Errorf("failed to fetch accounts followed by @%s: %w", twitterHandle, err)
	}

	// Extract token addresses and process them
	potentialTokens := wg.extractPotentialTokens(following, progressCallback)
	if len(potentialTokens) == 0 {
		// Cache the empty result to avoid repeated lookups
		wg.cacheMutex.Lock()
		wg.resultCache[twitterHandle] = result
		wg.cacheMutex.Unlock()
		return result, nil
	}

	// Find wallet addresses for each token
	rankedWallets := wg.findWalletsForTokens(potentialTokens, progressCallback)

	// Process the ranked wallets into the result
	result = wg.processRankedWallets(twitterHandle, rankedWallets, potentialTokens)

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

// ClearCache clears the cache
func (wg *WalletGuesser) ClearCache() {
	wg.cacheMutex.Lock()
	defer wg.cacheMutex.Unlock()
	wg.resultCache = make(map[string]*domain.WalletGuessResult)
	wg.tokenCache = make(map[string]domain.TokenInfo)
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
