package blockchain

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"wallet-guesser/internal/domain"
)

// Client handles interactions with the Solana blockchain
type Client struct {
	rpcEndpoint string
	httpClient  *http.Client
	avoidList   domain.AvoidListService
}

// NewClient creates a new blockchain client
func NewClient(rpcEndpoint string, avoidList domain.AvoidListService) *Client {
	return &Client{
		rpcEndpoint: rpcEndpoint,
		httpClient:  &http.Client{Timeout: 30 * time.Second},
		avoidList:   avoidList,
	}
}

// GetProgramAccounts fetches all accounts owned by a program
func (c *Client) GetProgramAccounts(programID string, filters []map[string]interface{}, progressCallback domain.ProgressCallback) ([]map[string]interface{}, error) {
	if progressCallback != nil {
		progressCallback(fmt.Sprintf("Fetching accounts for program %s...", programID))
	}

	// Prepare the RPC request
	req := RpcRequest{
		Jsonrpc: "2.0",
		ID:      1,
		Method:  "getProgramAccounts",
		Params: []interface{}{
			programID,
			map[string]interface{}{
				"encoding": "jsonParsed",
				"filters":  filters,
			},
		},
	}

	// Send the request
	rpcResp, err := c.sendRpcRequest(req)
	if err != nil {
		return nil, err
	}

	// Parse the result
	var accounts []map[string]interface{}
	if err := json.Unmarshal(rpcResp.Result, &accounts); err != nil {
		return nil, fmt.Errorf("failed to unmarshal program accounts: %w", err)
	}

	if progressCallback != nil {
		progressCallback(fmt.Sprintf("Found %d accounts for program %s", len(accounts), programID))
	}

	return accounts, nil
}

// GetWalletsForToken returns all wallet addresses that have interacted with a specific token
func (c *Client) GetWalletsForToken(mintAddress string, progressCallback domain.ProgressCallback) ([]string, error) {
	// Check if the token should be avoided
	if c.avoidList != nil {
		if shouldAvoid, reason := c.avoidList.ShouldAvoid(mintAddress); shouldAvoid {
			if progressCallback != nil {
				progressCallback(fmt.Sprintf("Skipping token %s: %s", mintAddress, reason))
			}
			return nil, nil
		}
	}

	if progressCallback != nil {
		progressCallback(fmt.Sprintf("Searching for wallets that interacted with token %s...", mintAddress))
	}

	// Create a filter to only get accounts for this mint
	filters := []map[string]interface{}{
		TokenBalanceFilter(mintAddress),
	}

	// Get all token accounts for this mint
	accounts, err := c.GetProgramAccounts(TokenProgramID, filters, progressCallback)
	if err != nil {
		return nil, fmt.Errorf("failed to get program accounts: %w", err)
	}

	// Extract wallet addresses (owners) from the accounts
	wallets := make(map[string]struct{}) // Use map to deduplicate

	for _, account := range accounts {
		// Extract owner address from account data
		if accountData, ok := account["account"].(map[string]interface{}); ok {
			if data, ok := accountData["data"].(map[string]interface{}); ok {
				if parsed, ok := data["parsed"].(map[string]interface{}); ok {
					if info, ok := parsed["info"].(map[string]interface{}); ok {
						if owner, ok := info["owner"].(string); ok {
							// Check if the wallet should be avoided
							if c.avoidList != nil {
								if shouldAvoid, reason := c.avoidList.ShouldAvoid(owner); shouldAvoid {
									if progressCallback != nil {
										progressCallback(fmt.Sprintf("Skipping wallet %s: %s", owner, reason))
									}
									continue
								}
							}
							wallets[owner] = struct{}{}
						}
					}
				}
			}
		}
	}

	// Convert map keys to slice
	result := make([]string, 0, len(wallets))
	for wallet := range wallets {
		result = append(result, wallet)
	}

	if progressCallback != nil {
		progressCallback(fmt.Sprintf("Found %d wallets that interacted with token %s", len(result), mintAddress))
	}

	return result, nil
}

// GetTokenInfo gets information about a token from its mint address
func (c *Client) GetTokenInfo(mintAddress string) (map[string]interface{}, error) {
	// This would typically query token metadata from the blockchain or a token registry
	// For now, we'll return a placeholder
	return map[string]interface{}{
		"address": mintAddress,
		"symbol":  "Unknown",
		"name":    "Unknown Token",
	}, nil
}

// sendRpcRequest sends a JSON-RPC request to the Solana node
func (c *Client) sendRpcRequest(request RpcRequest) (*RpcResponse, error) {
	// Marshal the request
	reqBody, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create an HTTP request
	httpReq, err := http.NewRequest("POST", c.rpcEndpoint, bytes.NewReader(reqBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}

	// Set headers
	httpReq.Header.Set("Content-Type", "application/json")

	// Send the request
	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send HTTP request: %w", err)
	}
	defer resp.Body.Close()

	// Read the response body
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// Parse the response
	var rpcResp RpcResponse
	if err := json.Unmarshal(respBody, &rpcResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	// Check if there was an RPC error
	if rpcResp.Error != nil {
		return nil, fmt.Errorf("RPC error: %s (code %d)", rpcResp.Error.Message, rpcResp.Error.Code)
	}

	return &rpcResp, nil
}
