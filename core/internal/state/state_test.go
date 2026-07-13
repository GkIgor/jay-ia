package state_test

import (
	"testing"

	"github.com/GkIgor/jay-ia/core/internal/state"
)

func TestStateString(t *testing.T) {
	tests := []struct {
		name     string
		input    state.State
		expected string
	}{
		{"Idle", state.Idle, "Idle"},
		{"Thinking", state.Thinking, "Thinking"},
		{"Executing", state.Executing, "Executing"},
		{"Waiting", state.Waiting, "Waiting"},
		{"Learning", state.Learning, "Learning"},
		{"Sleeping", state.Sleeping, "Sleeping"},
		{"Unknown negative", state.State(-1), "Unknown"},
		{"Unknown out of bounds", state.State(99), "Unknown"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			actual := tc.input.String()
			if actual != tc.expected {
				t.Errorf("expected %s, got %s", tc.expected, actual)
			}
		})
	}
}
