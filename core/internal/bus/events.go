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

// ToolProgressEvent is emitted when a running tool reports incremental progress
type ToolProgressEvent struct {
	ToolName string
	State    string
	Percent  float64
	Message  string
}

func (e ToolProgressEvent) EventName() string { return "tool.progress" }

// ToolCompletedEvent is emitted when a tool finishes execution
type ToolCompletedEvent struct {
	ToolName string
	Success  bool
	Output   any
	Error    string
}

func (e ToolCompletedEvent) EventName() string { return "tool.completed" }

// PermissionRequestedEvent is emitted when a tool requires user consent
type PermissionRequestedEvent struct {
	RequestID  string
	Permission string
	Prompt     string
}

func (e PermissionRequestedEvent) EventName() string { return "permission.request" }
