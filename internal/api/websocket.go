package api

import (
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
	"wallet-guesser/internal/game"
)

// Message represents a WebSocket message structure
type Message struct {
	Type    string      `json:"type"`
	Payload interface{} `json:"payload"`
}

// Handler manages the API endpoints
type Handler struct {
	upgrader websocket.Upgrader
	clients  map[*websocket.Conn]bool
	mutex    sync.Mutex
}

// NewHandler creates a new API handler
func NewHandler() *Handler {
	return &Handler{
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				// Allow all connections in development
				return true
			},
		},
		clients: make(map[*websocket.Conn]bool),
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
			// We'll implement game start logic later
			if err := sendJinnState(conn, "idle"); err != nil {
				log.Printf("Error sending jinn state: %v", err)
			}
		case "USER_INPUT":
			// We'll implement input handling later
			// For now, just acknowledge receipt
			if err := sendJinnState(conn, "thinking"); err != nil {
				log.Printf("Error sending jinn state: %v", err)
			}
		default:
			log.Printf("Unknown message type: %s", msg.Type)
		}
	}
}

// sendGameState sends the current game state to the client
func sendGameState(conn *websocket.Conn, state *game.State) error {
	return conn.WriteJSON(Message{
		Type:    "GAME_STATE",
		Payload: state,
	})
}

// sendJinnState sends a jinn state update to the client
func sendJinnState(conn *websocket.Conn, state string) error {
	return conn.WriteJSON(Message{
		Type: "JINN_STATE",
		Payload: map[string]string{
			"state": state,
		},
	})
}
