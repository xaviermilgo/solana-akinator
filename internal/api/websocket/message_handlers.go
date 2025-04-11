package websocket

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/gorilla/websocket"
	log "github.com/sirupsen/logrus"
	"wallet-guesser/internal/domain"
)

// handleUserInput processes a user's input (Twitter handle)
func (h *Handler) handleUserInput(conn *websocket.Conn, payload json.RawMessage) error {
	var inputPayload domain.UserInputPayload
	if err := json.Unmarshal(payload, &inputPayload); err != nil {
		return fmt.Errorf("error unmarshaling user input payload: %w", err)
	}

	twitterHandle := inputPayload.Twitter
	if twitterHandle == "" {
		return SendJinnState(conn, string(domain.JinnStateAsking), "The Jinn needs a Twitter handle to divine the wallet address.")
	}

	// Start the wallet guessing process in a goroutine
	go h.processWalletGuess(conn, twitterHandle)

	return nil
}

// processWalletGuess handles the wallet guessing process
func (h *Handler) processWalletGuess(conn *websocket.Conn, twitterHandle string) {
	// First, update the UI to show we're thinking
	if err := SendJinnState(conn, string(domain.JinnStateThinking), "Hmm... I'm consulting the mystical blockchain ledgers..."); err != nil {
		log.Errorf("Error sending thinking state: %v", err)
		return
	}

	// Define a progress callback to update the user
	progressCallback := func(message string) {
		log.Infof("[%s] update: %s", twitterHandle, message)
		// Send progress update to the client
		if err := SendProgressUpdate(conn, message); err != nil {
			log.Errorf("Error sending progress update: %v", err)
		}
	}

	// Call the wallet guesser
	result, err := h.walletGuesserSvc.GuessWallet(twitterHandle, progressCallback)
	if err != nil {
		log.Errorf("Error guessing wallet: %v", err)
		SendJinnState(conn, string(domain.JinnStateWrong), "The crypto spirits are not cooperating today. Please try again later.")
		return
	}

	// Process the result
	if len(result.Addresses) > 0 {
		// We found potential wallet addresses
		confidence := result.Confidence

		// Send the result back to the client
		if err := SendWalletGuesserResult(conn, result); err != nil {
			log.Errorf("Error sending wallet result: %v", err)
			return
		}

		// Update the Jinn state based on confidence
		if confidence >= 70 {
			SendJinnState(conn, string(domain.JinnStateConfident), "Aha! I sense strong wallet energy from this Twitter handle!")
			time.Sleep(1500 * time.Millisecond)
			SendJinnState(conn, string(domain.JinnStateCorrect), "Behold! I have divined your wallet correctly!")
		} else if confidence >= 40 {
			SendJinnState(conn, string(domain.JinnStateAsking), "I sense some wallet energy, but I'm not entirely sure...")
		} else {
			SendJinnState(conn, string(domain.JinnStateWrong), "The blockchain spirits have whispered some addresses, but I'm uncertain...")
		}
	} else {
		// No wallet addresses found
		SendJinnState(conn, string(domain.JinnStateWrong), "I could not divine any wallet addresses for this Twitter handle.")
	}
}
