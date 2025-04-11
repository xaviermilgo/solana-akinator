package domain

// JinnState represents the state of the Jinn character
type JinnState string

// Valid Jinn states
const (
	JinnStateIdle      JinnState = "idle"
	JinnStateThinking  JinnState = "thinking"
	JinnStateAsking    JinnState = "asking"
	JinnStateConfident JinnState = "confident"
	JinnStateCorrect   JinnState = "correct"
	JinnStateWrong     JinnState = "wrong"
	JinnStateGlitched  JinnState = "glitched"
)

// GameState represents the current state of the game
type GameState struct {
	JinnState JinnState `json:"jinnState"`
	Twitter   string    `json:"twitter,omitempty"`
}

// TwitterUser represents data for a Twitter user
type TwitterUser struct {
	Username              string   `json:"username"`
	DisplayName           string   `json:"displayName,omitempty"`
	Bio                   string   `json:"bio,omitempty"`
	Urls                  []string `json:"urls,omitempty"`
	PossibleMintAddresses []string `json:"possibleAddresses,omitempty"`
}

// WalletGuessResult represents the result of the wallet guessing process
type WalletGuessResult struct {
	TwitterHandle string   `json:"twitterHandle"`
	Addresses     []string `json:"addresses"`
	Sources       []string `json:"sources"`
	Confidence    int      `json:"confidence"` // 0-100
}

// WebSocketMessage represents a WebSocket message structure
type WebSocketMessage struct {
	Type    string      `json:"type"`
	Payload interface{} `json:"payload"`
}

// UserInputPayload represents the payload for USER_INPUT messages
type UserInputPayload struct {
	Twitter string `json:"twitter"`
}

// ProgressMessage represents a progress update message
type ProgressMessage struct {
	Message string `json:"message"`
}

// JinnStatePayload represents a Jinn state update message payload
type JinnStatePayload struct {
	State   string `json:"state"`
	Message string `json:"message"`
}

// AvoidListEntry represents an entry in the avoid list
type AvoidListEntry struct {
	Prefix string `json:"prefix"`
	Type   string `json:"type"` // "t" for token, "w" for wallet
}

type AvoidListZipped struct {
	ZippedPrefix []string `json:"zip_prefix"`
	ZippedType   []string `json:"zip_types"` // "t" for token, "w" for wallet
}

// AvoidListResponse represents the response from the avoid list API
type AvoidListResponse struct {
	Result struct {
		Rows []AvoidListZipped `json:"rows"`
	} `json:"result"`
}

// TokenInfo represents information about a token project
type TokenInfo struct {
	Symbol      string
	Name        string
	MintAddress string
}
