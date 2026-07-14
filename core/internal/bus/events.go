package bus

// Event interface allows different types of events to be passed in the bus
type Event interface {
	EventName() string
}

// StateChangedEvent is emitted when the Avatar's logical state changes
type StateChangedEvent struct {
	NewState string
}

func (e StateChangedEvent) EventName() string { return "state.changed" }

// AnimationPlayEvent is emitted when an animation should be played
type AnimationPlayEvent struct {
	Animation string
}

func (e AnimationPlayEvent) EventName() string { return "animation.play" }
