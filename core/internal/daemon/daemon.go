package daemon

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/GkIgor/jay-ia/core/internal/ipc"
	"github.com/GkIgor/jay-ia/core/internal/memory"
	"github.com/GkIgor/jay-ia/core/internal/planner"
	"github.com/GkIgor/jay-ia/core/internal/state"
	sdkipc "github.com/GkIgor/jay-ia/sdk/ipc"
)

// Daemon coordinates all core components
type Daemon struct {
	currentState state.State
	memoryStore  memory.MemoryStore
	ipcServer    *ipc.Server
	planner      planner.Planner
}

// New creates a new Jay Daemon
func New() (*Daemon, error) {
	d := &Daemon{
		currentState: state.Idle,
		memoryStore:  memory.NewInMemoryStore(),
		planner:      planner.NewSimplePlanner(),
	}

	// Create IPC server with the daemon's message handler callback
	ipcServer, err := ipc.NewServer(d.handleIPCMessage)
	if err != nil {
		return nil, err
	}
	d.ipcServer = ipcServer

	return d, nil
}

// Start starts the daemon and blocks until stopped
func (d *Daemon) Start() error {
	log.Printf("Starting Jay Daemon... Current State: %s", d.currentState)

	if err := d.ipcServer.Start(); err != nil {
		return err
	}

	// Wait for termination signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	<-sigChan
	log.Println("Termination signal received. Shutting down...")

	d.Stop()
	return nil
}

// Stop safely shuts down the daemon
func (d *Daemon) Stop() {
	if d.ipcServer != nil {
		d.ipcServer.Stop()
	}
	log.Println("Jay Daemon stopped.")
}

// handleIPCMessage acts as the main orchestrator for commands received from IPC.
func (d *Daemon) handleIPCMessage(msg sdkipc.Message) sdkipc.Message {
	if msg.Type != "command" {
		return sdkipc.Message{
			Type: "error",
			Payload: sdkipc.Error{
				Code:    400,
				Message: fmt.Sprintf("invalid message type: %s", msg.Type),
			},
		}
	}

	// Unmarshal command payload
	payloadBytes, err := json.Marshal(msg.Payload)
	if err != nil {
		return d.errorResponse("internal_error", "failed to process command payload")
	}

	var cmd sdkipc.Command
	if err := json.Unmarshal(payloadBytes, &cmd); err != nil {
		return d.errorResponse("invalid_command", "invalid command structure")
	}

	// 1. Resolve/Prepare PlanningContext (Perception phase)
	d.currentState = state.Thinking
	defer func() { d.currentState = state.Idle }()

	planCtx := planner.PlanningContext{
		WorkingMemory: make(map[string]string),
	}

	input := ""
	if cmd.Data != nil {
		if s, ok := cmd.Data.(string); ok {
			input = s
		}
	}

	// If no data string but Action exists, construct command-style input
	if input == "" && cmd.Action != "" {
		input = "/" + cmd.Action
	}

	// If the command is recall, pre-fetch memory key so it's ready in the context
	if cmd.Action == "recall" || strings.HasPrefix(input, "/recall") {
		var key string
		if strings.HasPrefix(input, "/recall") {
			parts := strings.SplitN(input, " ", 2)
			if len(parts) >= 2 {
				key = strings.TrimSpace(parts[1])
			}
		} else {
			key = input
		}

		if key != "" {
			val, err := d.memoryStore.Get(key)
			if err == nil {
				if valStr, ok := val.(string); ok {
					planCtx.WorkingMemory[key] = valStr
				}
			}
		}
	}

	// 2. Call Planner (pure function)
	plan, err := d.planner.Plan(context.Background(), input, planCtx)
	if err != nil {
		return d.errorResponse(cmd.ID, fmt.Sprintf("planning error: %v", err))
	}

	// 3. Execute Steps (Side-effects execution phase)
	d.currentState = state.Executing
	var responseText string

	for _, step := range plan.Steps {
		switch step.Type {
		case planner.StepRespond:
			if text, ok := step.Params["text"].(string); ok {
				responseText = text
			}
		case planner.StepMemoryPut:
			key, kOk := step.Params["key"].(string)
			val, vOk := step.Params["value"].(string)
			if kOk && vOk {
				if err := d.memoryStore.Put(key, val); err != nil {
					log.Printf("Failed to write to memory: %v", err)
				}
			}
		case planner.StepHumanEscalate:
			responseText = "Escalated to human."
		}
	}

	return sdkipc.Message{
		Type: "response",
		Payload: sdkipc.Response{
			RefID:  cmd.ID,
			Status: "ok",
			Data:   responseText,
		},
	}
}

func (d *Daemon) errorResponse(refID string, message string) sdkipc.Message {
	return sdkipc.Message{
		Type: "response",
		Payload: sdkipc.Response{
			RefID:  refID,
			Status: "error",
			Data:   message,
		},
	}
}
