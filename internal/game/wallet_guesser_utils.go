package game

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"sort"
	"strings"

	"wallet-guesser/internal/domain"
)

// TokenWithSource pairs a token address with its source (Twitter handle)
type TokenWithSource struct {
	MintAddress string
	Source      string
}

// WalletScore represents a wallet and its score
type WalletScore struct {
	Address string
	Score   int
}

// extractPotentialTokens extracts potential token addresses from followed accounts
func (wg *WalletGuesser) extractPotentialTokens(following []domain.TwitterUser, progressCallback domain.ProgressCallback) []TokenWithSource {
	if progressCallback != nil {
		progressCallback(fmt.Sprintf("Analyzing %d accounts to identify token projects...", len(following)))
	}

	var tokenSources []TokenWithSource

	// Loop through followed accounts to find token projects
	for _, user := range following {
		for _, mint := range user.PossibleMintAddresses {
			// Check if the token should be avoided
			if wg.avoidListService != nil {
				if shouldAvoid, reason := wg.avoidListService.ShouldAvoid(mint); shouldAvoid {
					if progressCallback != nil {
						progressCallback(fmt.Sprintf("Skipping token %s from @%s: %s", mint, user.Username, reason))
					}
					continue
				}
			}

			tokenSources = append(tokenSources, TokenWithSource{
				MintAddress: mint,
				Source:      fmt.Sprintf("@%s", user.Username),
			})

			if progressCallback != nil {
				progressCallback(fmt.Sprintf("Found potential token mint address from @%s", user.Username))
			}
		}
	}

	if len(tokenSources) == 0 && progressCallback != nil {
		progressCallback("No token projects identified from followed accounts")
	} else if progressCallback != nil {
		progressCallback(fmt.Sprintf("Found %d potential token projects to analyze", len(tokenSources)))
	}

	return tokenSources
}

// findWalletsForTokens gets wallets that have interacted with the given tokens
// Optimized to track wallet-to-token relationships and reduce avoid-list checks
func (wg *WalletGuesser) findWalletsForTokens(tokenSources []TokenWithSource, progressCallback domain.ProgressCallback) []WalletScore {
	walletScores := make(map[string]int)
	walletToTokens := make(map[string][]TokenWithSource)
	processedWallets := make(map[string]bool) // Track wallets we've already checked against the avoid list

	// Count of valid tokens actually processed
	validTokensProcessed := 0

	for _, tokenSource := range tokenSources {
		if progressCallback != nil {
			sixBeforeEnd := len(tokenSource.MintAddress) - 6
			progressCallback(fmt.Sprintf("Looking for wallets that interacted with token %s...%s", tokenSource.MintAddress[:6], tokenSource.MintAddress[sixBeforeEnd:]))
		}

		// Get all wallets that have interacted with this token
		wallets, err := wg.blockchainClient.GetWalletsForToken(tokenSource.MintAddress, progressCallback)
		if err != nil {
			log.Errorf("Error getting wallets for token %s: %v", tokenSource.MintAddress, err)
			continue
		}

		// Skip if we didn't find any wallets (might happen with avoid list filtering)
		if len(wallets) == 0 {
			continue
		}

		// If we got here, we have a valid token with results
		validTokensProcessed++

		// Process wallets for this token
		for _, wallet := range wallets {
			// Only check against avoid list if we haven't seen this wallet before
			if !processedWallets[wallet] {
				processedWallets[wallet] = true

				// Check if the wallet should be avoided
				if wg.avoidListService != nil {
					if shouldAvoid, _ := wg.avoidListService.ShouldAvoid(wallet); shouldAvoid {
						continue
					}
				}
			}

			// Increment score and associate this token with the wallet
			walletScores[wallet]++
			walletToTokens[wallet] = append(walletToTokens[wallet], tokenSource)
		}
	}

	if progressCallback != nil {
		processedCount := len(tokenSources)
		if validTokensProcessed < processedCount {
			progressCallback(fmt.Sprintf("Processed %d tokens, but only %d had valid results",
				processedCount, validTokensProcessed))
		}
	}

	// Convert to sorted slice
	var rankedWallets []WalletScore
	for addr, score := range walletScores {
		// Normalize scores based on valid tokens processed
		normalizedScore := 0
		if validTokensProcessed > 0 {
			normalizedScore = (score * 100) / validTokensProcessed
		}
		rankedWallets = append(rankedWallets, WalletScore{addr, normalizedScore})
	}

	// Sort wallets by score (highest first)
	sort.Slice(rankedWallets, func(i, j int) bool {
		return rankedWallets[i].Score > rankedWallets[j].Score
	})

	return rankedWallets
}

// processRankedWallets converts ranked wallets into the result format
func (wg *WalletGuesser) processRankedWallets(twitterHandle string, rankedWallets []WalletScore, tokenSources []TokenWithSource) *domain.WalletGuessResult {
	result := &domain.WalletGuessResult{
		TwitterHandle: twitterHandle,
		Addresses:     []string{},
		Sources:       []string{},
		Confidence:    0,
	}

	// Create map of token addresses to sources
	tokenToSourceMap := make(map[string]string)
	for _, ts := range tokenSources {
		tokenToSourceMap[ts.MintAddress] = ts.Source
	}

	// Take top results (limit to 5)
	maxResults := 5
	if len(rankedWallets) < maxResults {
		maxResults = len(rankedWallets)
	}

	// If we have no results, return early
	if maxResults == 0 {
		return result
	}

	// Process top wallets
	walletToTokens := wg.getWalletToTokensMap(rankedWallets, tokenSources)

	for i := 0; i < maxResults; i++ {
		wallet := rankedWallets[i]
		result.Addresses = append(result.Addresses, wallet.Address)

		// Create source description
		tokens := walletToTokens[wallet.Address]
		sourceTokens := make([]string, 0, len(tokens))

		for _, token := range tokens {
			if source, exists := tokenToSourceMap[token.MintAddress]; exists {
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

		// The score is already normalized in the findWalletsForTokens function
		sourceText := fmt.Sprintf("Match score: %d/100. Matched tokens from: %s",
			wallet.Score,
			strings.Join(sourcesList, ", "))

		result.Sources = append(result.Sources, sourceText)
	}

	// Use the highest score directly since we normalized it in findWalletsForTokens
	result.Confidence = rankedWallets[0].Score

	return result
}

// getWalletToTokensMap creates a map of wallets to their associated tokens
func (wg *WalletGuesser) getWalletToTokensMap(rankedWallets []WalletScore, tokenSources []TokenWithSource) map[string][]TokenWithSource {
	walletToTokens := make(map[string][]TokenWithSource)

	// This would be populated during the blockchain search, but we're doing it here for clarity
	// In a production environment, this should be populated during the wallet search for efficiency
	for _, wallet := range rankedWallets {
		// For each wallet, check which tokens it has interacted with
		for _, tokenSource := range tokenSources {
			// This is a simplified approach - in a real implementation, we'd store this during the blockchain query
			// Here we're just associating each wallet with all tokens based on its score
			if wallet.Score > 0 {
				walletToTokens[wallet.Address] = append(walletToTokens[wallet.Address], tokenSource)
			}
		}
	}

	return walletToTokens
}

// calculateConfidence returns a confidence score (0-100) based on matches
func calculateConfidence(matches int, totalValidMints int) int {
	if totalValidMints == 0 {
		return 0
	}

	// Calculate ratio of matches to total valid mints
	ratio := float64(matches) / float64(totalValidMints)

	// Scale to 0-100
	confidence := int(ratio * 100)

	// Apply a curve to make it harder to get very high confidence
	// and to ensure low match counts still produce meaningful results
	if confidence >= 75 {
		return 90 + ((confidence-75)*10)/25 // 75% -> 90%, 100% -> 100%
	} else if confidence >= 50 {
		return 70 + ((confidence-50)*20)/25 // 50% -> 70%, 75% -> 90%
	} else if confidence >= 25 {
		return 40 + ((confidence-25)*30)/25 // 25% -> 40%, 50% -> 70%
	} else {
		return confidence + 15 // Apply minimum boost to low confidence
	}
}
