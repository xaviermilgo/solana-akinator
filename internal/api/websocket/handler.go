package websocket

import (
	"encoding/json"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
	log "github.com/sirupsen/logrus"
	"wallet-guesser/internal/domain"
)

// Handler manages WebSocket connections
type Handler struct {
	upgrader            websocket.Upgrader
	clients             map[*websocket.Conn]bool
	mutex               sync.Mutex
	walletGuesserSvc    domain.WalletGuesserService
	messageHandlerFuncs map[string]MessageHandlerFunc
}

// MessageHandlerFunc is a function that handles a specific message type
type MessageHandlerFunc func(conn *websocket.Conn, payload json.RawMessage) error

// NewHandler creates a new WebSocket handler
func NewHandler(walletGuesserSvc domain.WalletGuesserService) *Handler {
	h := &Handler{
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				// Allow all connections in development
				return true
			},
		},
		clients:          make(map[*websocket.Conn]bool),
		walletGuesserSvc: walletGuesserSvc,
	}

	// Register message handlers
	h.messageHandlerFuncs = map[string]MessageHandlerFunc{
		"START_GAME": h.handleStartGame,
		"USER_INPUT": h.handleUserInput,
	}

	return h
}

// HandleWebSocket handles WebSocket connections
func (h *Handler) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	// Upgrade HTTP connection to WebSocket
	conn, err := h.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Errorf("Failed to upgrade to WebSocket: %v", err)
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
	initialState := &domain.GameState{JinnState: domain.JinnStateIdle}
	if err := SendGameState(conn, initialState); err != nil {
		log.Errorf("Error sending initial state: %v", err)
		return
	}

	// Message handling loop
	for {
		var msg domain.WebSocketMessage
		err := conn.ReadJSON(&msg)
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Errorf("WebSocket error: %v", err)
			}
			break
		}

		// Process the message based on its type
		if handlerFunc, ok := h.messageHandlerFuncs[msg.Type]; ok {
			// Marshal the payload for type-specific handling
			payloadBytes, err := json.Marshal(msg.Payload)
			if err != nil {
				log.Errorf("Error marshaling payload: %v", err)
				SendJinnState(conn, string(domain.JinnStateGlitched), "The Jinn has encountered an error interpreting your request.")
				continue
			}

			err = handlerFunc(conn, payloadBytes)
			if err != nil {
				log.Errorf("Error handling message '%s': %v", msg.Type, err)
				SendJinnState(conn, string(domain.JinnStateGlitched), "The Jinn has encountered an error processing your request.")
			}
		} else {
			log.Warnf("Unknown message type: %s", msg.Type)
		}
	}
}

// handleStartGame handles the START_GAME message
func (h *Handler) handleStartGame(conn *websocket.Conn, _ json.RawMessage) error {
	// Set the game state to idle
	return SendJinnState(conn, string(domain.JinnStateIdle), "I am the Crypto Jinn! I can divine your wallet address from your Twitter handle!")
}

// BroadcastMessage sends a message to all connected clients
func (h *Handler) BroadcastMessage(message domain.WebSocketMessage) {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	for conn := range h.clients {
		if err := conn.WriteJSON(message); err != nil {
			log.Errorf("Error broadcasting message: %v", err)
			conn.Close()
			delete(h.clients, conn)
		}
	}
}
