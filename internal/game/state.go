package game

// State represents the current game state
type State struct {
	JinnState JinnState `json:"jinnState"`
	Twitter   string    `json:"twitter,omitempty"`
}

type JinnState string

// Valid jinn states
const (
	JinnStateIdle      JinnState = "idle"
	JinnStateThinking            = "thinking"
	JinnStateAsking              = "asking"
	JinnStateConfident           = "confident"
	JinnStateCorrect             = "correct"
	JinnStateWrong               = "wrong"
	JinnStateGlitched            = "glitched"
)

// NewGameState creates a new game state with default values
func NewGameState() *State {
	return &State{
		JinnState: JinnStateIdle,
	}
}
