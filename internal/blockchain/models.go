package blockchain

import (
	"encoding/json"
)

// RpcRequest represents a JSON-RPC request
type RpcRequest struct {
	Jsonrpc string        `json:"jsonrpc"`
	ID      int           `json:"id"`
	Method  string        `json:"method"`
	Params  []interface{} `json:"params"`
}

// RpcResponse represents a JSON-RPC response
type RpcResponse struct {
	Jsonrpc string          `json:"jsonrpc"`
	ID      int             `json:"id"`
	Result  json.RawMessage `json:"result"`
	Error   *RpcError       `json:"error,omitempty"`
}

// RpcError represents a JSON-RPC error
type RpcError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// AccountInfo represents the data of a Solana account
type AccountInfo struct {
	Lamports   uint64   `json:"lamports"`
	Owner      string   `json:"owner"`
	Data       []string `json:"data"`
	Executable bool     `json:"executable"`
	RentEpoch  uint64   `json:"rentEpoch"`
}

// TokenBalanceFilter creates a filter for SPL token balances
func TokenBalanceFilter(mintAddress string) map[string]interface{} {
	return map[string]interface{}{
		"memcmp": map[string]interface{}{
			"offset": 0,
			"bytes":  mintAddress,
		},
	}
}

// Default program IDs used in Solana
const (
	TokenProgramID = "TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA"
)
