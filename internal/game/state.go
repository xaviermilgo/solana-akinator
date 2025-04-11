package game

// State represents the current game state
type State struct {
	JinnState string `json:"jinnState"`
	Twitter   string `json:"twitter,omitempty"`
}

// Valid jinn states
const (
	JinnStateIdle      = "idle"
	JinnStateThinking  = "thinking"
	JinnStateAsking    = "asking"
	JinnStateConfident = "confident"
	JinnStateCorrect   = "correct"
	JinnStateWrong     = "wrong"
	JinnStateGlitched  = "glitched"
)

// NewGameState creates a new game state with default values
func NewGameState() *State {
	return &State{
		JinnState: JinnStateIdle,
	}
}

// SetJinnState updates the jinn's state
func (s *State) SetJinnState(state string) {
	s.JinnState = state
}

// SetTwitterHandle sets the Twitter handle to be analyzed
func (s *State) SetTwitterHandle(handle string) {
	s.Twitter = handle
}
