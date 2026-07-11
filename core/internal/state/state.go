package state

// State represents the transient operating condition of the Core
type State int

const (
	Idle State = iota
	Thinking
	Executing
	Waiting
	Learning
	Sleeping
)

func (s State) String() string {
	switch s {
	case Idle:
		return "Idle"
	case Thinking:
		return "Thinking"
	case Executing:
		return "Executing"
	case Waiting:
		return "Waiting"
	case Learning:
		return "Learning"
	case Sleeping:
		return "Sleeping"
	default:
		return "Unknown"
	}
}
