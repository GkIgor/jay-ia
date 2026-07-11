package ipc

// Message represents a generic IPC message wrapper
type Message struct {
	Type    string `json:"type"` // e.g. "command", "event", "response", "error"
	Payload any    `json:"payload"`
}

// Command represents an action requested by the client
type Command struct {
	ID     string `json:"id"`
	Action string `json:"action"`
	Data   any    `json:"data,omitempty"`
}

// Response represents a reply to a command
type Response struct {
	RefID  string `json:"ref_id"`
	Status string `json:"status"` // e.g. "ok", "error"
	Data   any    `json:"data,omitempty"`
}

// Event represents an asynchronous notification emitted by the Core
type Event struct {
	Topic string `json:"topic"`
	Data  any    `json:"data,omitempty"`
}

// Error represents an error response
type Error struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}
