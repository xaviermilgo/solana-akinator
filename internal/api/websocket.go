package api

import (
	"encoding/json"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	log "github.com/sirupsen/logrus"
	"wallet-guesser/internal/game"
)

// Message represents a WebSocket message structure
type Message struct {
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

// WalletGuesserResult represents the result of a wallet guessing process
type WalletGuesserResult struct {
	TwitterHandle string   `json:"twitterHandle"`
	Addresses     []string `json:"addresses"`
	Sources       []string `json:"sources"`
	Confidence    int      `json:"confidence"`
}

// Handler manages the API endpoints
type Handler struct {
	upgrader      websocket.Upgrader
	clients       map[*websocket.Conn]bool
	mutex         sync.Mutex
	walletGuesser *game.WalletGuesser
}

// NewHandler creates a new API handler
func NewHandler(walletGuesser *game.WalletGuesser) *Handler {
	return &Handler{
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				// Allow all connections in development
				return true
			},
		},
		clients:       make(map[*websocket.Conn]bool),
		walletGuesser: walletGuesser,
	}
}

// HandleWebSocket handles WebSocket connections
func (h *Handler) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	// Upgrade HTTP connection to WebSocket
	conn, err := h.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Failed to upgrade to WebSocket: %v", err)
		return
	}
	defer conn.Close()

	// Register new client
	h.mutex.Lock()
	h.clients[conn] = true
	h.mutex.Unlock()

	// Remove client when the function returns
	defer func() {
		h.mutex.Lock()
		delete(h.clients, conn)
		h.mutex.Unlock()
	}()

	// Send initial state
	initialState := game.NewGameState()
	if err := sendGameState(conn, initialState); err != nil {
		log.Printf("Error sending initial state: %v", err)
		return
	}

	// Message handling loop
	for {
		var msg Message
		err := conn.ReadJSON(&msg)
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket error: %v", err)
			}
			break
		}

		// Process the message based on its type
		switch msg.Type {
		case "START_GAME":
			// Set the game state to idle
			sendJinnState(conn, "idle", "I am the Crypto Jinn! I can divine your wallet address from your Twitter handle!")
		case "USER_INPUT":
			// Handle user input (Twitter handle)
			h.handleUserInput(conn, msg.Payload)
		default:
			log.Printf("Unknown message type: %s", msg.Type)
		}
	}
}

// handleUserInput processes a user's input (Twitter handle)
func (h *Handler) handleUserInput(conn *websocket.Conn, payload interface{}) {
	// Parse the payload
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		log.Printf("Error marshaling payload: %v", err)
		sendJinnState(conn, "glitched", "The Jinn has encountered an error interpreting your request.")
		return
	}

	var inputPayload UserInputPayload
	if err := json.Unmarshal(payloadBytes, &inputPayload); err != nil {
		log.Printf("Error unmarshaling payload: %v", err)
		sendJinnState(conn, "glitched", "The Jinn has encountered an error interpreting your request.")
		return
	}

	twitterHandle := inputPayload.Twitter
	if twitterHandle == "" {
		sendJinnState(conn, "asking", "The Jinn needs a Twitter handle to divine the wallet address.")
		return
	}

	// Start the wallet guessing process in a goroutine
	go func() {
		// First, update the UI to show we're thinking
		sendJinnState(conn, "thinking", "Hmm... I'm consulting the mystical blockchain ledgers...")

		// Define a progress callback to update the user
		progressCallback := func(message string) {
			log.Infof("[%s] update : %s", twitterHandle, message)
			// Send progress update to the client
			sendProgressUpdate(conn, message)
			// Small delay to make the updates more readable
			time.Sleep(300 * time.Millisecond)
		}

		// Call the wallet guesser
		result, err := h.walletGuesser.GuessWallet(twitterHandle, progressCallback)
		if err != nil {
			log.Printf("Error guessing wallet: %v", err)
			sendJinnState(conn, "wrong", "The crypto spirits are not cooperating today. Please try again later.")
			return
		}

		// Process the result
		if len(result.Addresses) > 0 {
			// We found potential wallet addresses
			confidence := result.Confidence

			// Send the result back to the client
			sendWalletGuesserResult(conn, result)

			// Update the Jinn state based on confidence
			if confidence >= 70 {
				sendJinnState(conn, "confident", "Aha! I sense strong wallet energy from this Twitter handle!")
				time.Sleep(1500 * time.Millisecond)
				sendJinnState(conn, "correct", "Behold! I have divined your wallet correctly!")
			} else if confidence >= 40 {
				sendJinnState(conn, "asking", "I sense some wallet energy, but I'm not entirely sure...")
			} else {
				sendJinnState(conn, "wrong", "The blockchain spirits have whispered some addresses, but I'm uncertain...")
			}
		} else {
			// No wallet addresses found
			sendJinnState(conn, "wrong", "I could not divine any wallet addresses for this Twitter handle.")
		}
	}()
}

// sendGameState sends the current game state to the client
func sendGameState(conn *websocket.Conn, state *game.State) error {
	return conn.WriteJSON(Message{
		Type:    "GAME_STATE",
		Payload: state,
	})
}

// sendJinnState sends a jinn state update to the client
func sendJinnState(conn *websocket.Conn, state game.JinnState, message string) {
	err := conn.WriteJSON(Message{
		Type: "JINN_STATE",
		Payload: map[string]string{
			"state":   string(state),
			"message": message,
		},
	})
	if err != nil {
		log.WithError(err).Error("Error sending jinn state")
	}
}

// sendProgressUpdate sends a progress update to the client
func sendProgressUpdate(conn *websocket.Conn, message string) {
	err := conn.WriteJSON(Message{
		Type: "PROGRESS_UPDATE",
		Payload: ProgressMessage{
			Message: message,
		},
	})
	if err != nil {
		log.WithError(err).Error("Error sending progress update")
	}
}

// sendWalletGuesserResult sends the wallet guesser result to the client
func sendWalletGuesserResult(conn *websocket.Conn, result *game.WalletGuessResult) {
	err := conn.WriteJSON(Message{
		Type: "WALLET_RESULT",
		Payload: WalletGuesserResult{
			TwitterHandle: result.TwitterHandle,
			Addresses:     result.Addresses,
			Sources:       result.Sources,
			Confidence:    result.Confidence,
		},
	})
	if err != nil {
		log.WithError(err).Error("Error sending wallet guess result")
	}
}
