package websocket

import (
	"github.com/gorilla/websocket"
	log "github.com/sirupsen/logrus"
	"wallet-guesser/internal/domain"
)

// SendGameState sends the current game state to the client
func SendGameState(conn *websocket.Conn, state *domain.GameState) error {
	return conn.WriteJSON(domain.WebSocketMessage{
		Type:    "GAME_STATE",
		Payload: state,
	})
}

// SendJinnState sends a jinn state update to the client
func SendJinnState(conn *websocket.Conn, state string, message string) error {
	err := conn.WriteJSON(domain.WebSocketMessage{
		Type: "JINN_STATE",
		Payload: domain.JinnStatePayload{
			State:   state,
			Message: message,
		},
	})
	if err != nil {
		log.WithError(err).Error("Error sending jinn state")
	}
	return err
}

// SendProgressUpdate sends a progress update to the client
func SendProgressUpdate(conn *websocket.Conn, message string) error {
	err := conn.WriteJSON(domain.WebSocketMessage{
		Type: "PROGRESS_UPDATE",
		Payload: domain.ProgressMessage{
			Message: message,
		},
	})
	if err != nil {
		log.WithError(err).Error("Error sending progress update")
	}
	return err
}

// SendWalletGuesserResult sends the wallet guesser result to the client
func SendWalletGuesserResult(conn *websocket.Conn, result *domain.WalletGuessResult) error {
	err := conn.WriteJSON(domain.WebSocketMessage{
		Type:    "WALLET_RESULT",
		Payload: result,
	})
	if err != nil {
		log.WithError(err).Error("Error sending wallet guess result")
	}
	return err
}
