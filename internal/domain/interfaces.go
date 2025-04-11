package domain

// TwitterService defines the interface for Twitter API interactions
type TwitterService interface {
	// FetchFollowing fetches the accounts a user is following
	FetchFollowing(username string, limit int, progressCallback ProgressCallback) ([]TwitterUser, error)
}

// BlockchainService defines the interface for blockchain interactions
type BlockchainService interface {
	// GetWalletsForToken returns all wallet addresses that have interacted with a specific token
	GetWalletsForToken(mintAddress string, progressCallback ProgressCallback) ([]string, error)
	// GetTokenInfo gets information about a token from its mint address
	GetTokenInfo(mintAddress string) (map[string]interface{}, error)
}

// WalletGuesserService defines the interface for wallet guessing functionality
type WalletGuesserService interface {
	// GuessWallet tries to guess the wallet address for a given Twitter handle
	GuessWallet(twitterHandle string, progressCallback ProgressCallback) (*WalletGuessResult, error)
	// ClearCache clears the cache
	ClearCache()
	// CacheStats returns statistics about the cache
	CacheStats() map[string]interface{}
}

// AvoidListService defines the interface for the avoid list functionality
type AvoidListService interface {
	// ShouldAvoid checks if an address should be avoided
	ShouldAvoid(address string) (bool, string)
	// UpdateAvoidList updates the avoid list from the remote API
	UpdateAvoidList() error
	// GetAvoidListStats returns statistics about the avoid list
	GetAvoidListStats() map[string]interface{}
}

// WebSocketHandler defines the interface for WebSocket message handling
type WebSocketHandler interface {
	// HandleMessage handles an incoming WebSocket message
	HandleMessage(message *WebSocketMessage, sendMessage func(message *WebSocketMessage) error) error
}

// ProgressCallback is a function type for reporting progress
type ProgressCallback func(message string)
