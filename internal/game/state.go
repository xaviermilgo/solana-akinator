package game

import (
	"wallet-guesser/internal/domain"
)

// NewGameState creates a new game state with default values
func NewGameState() *domain.GameState {
	return &domain.GameState{
		JinnState: domain.JinnStateIdle,
	}
}
