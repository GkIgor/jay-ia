package planner_test

import (
	"context"
	"testing"

	"github.com/GkIgor/jay-ia/core/internal/planner"
)

func TestSimplePlanner_Plan(t *testing.T) {
	p := planner.NewSimplePlanner()
	ctx := context.Background()

	tests := []struct {
		name       string
		input      string
		planCtx    planner.PlanningContext
		wantSteps  []planner.StepType
		wantText   string
		wantMemory map[string]string // key: value expected in StepMemoryPut params
	}{
		{
			name:      "Ping Command",
			input:     "/ping",
			wantSteps: []planner.StepType{planner.StepRespond},
			wantText:  "pong",
		},
		{
			name:      "Memo Command Valid",
			input:     "/memo language=Go",
			wantSteps: []planner.StepType{planner.StepMemoryPut, planner.StepRespond},
			wantText:  "Saved 'language' to memory.",
			wantMemory: map[string]string{
				"key":   "language",
				"value": "Go",
			},
		},
		{
			name:      "Memo Command Invalid",
			input:     "/memo language",
			wantSteps: []planner.StepType{planner.StepRespond},
			wantText:  "Usage: /memo key=value",
		},
		{
			name:  "Recall Command Found",
			input: "/recall language",
			planCtx: planner.PlanningContext{
				WorkingMemory: map[string]string{
					"language": "Go",
				},
			},
			wantSteps: []planner.StepType{planner.StepRespond},
			wantText:  "I remember: language = Go",
		},
		{
			name:      "Recall Command Not Found",
			input:     "/recall nonexistent",
			wantSteps: []planner.StepType{planner.StepRespond},
			wantText:  "I don't remember anything for 'nonexistent'.",
		},
		{
			name:      "Normal Text Mode",
			input:     "Hello Jay",
			wantSteps: []planner.StepType{planner.StepRespond},
			wantText:  "I received your message: \"Hello Jay\"",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			plan, err := p.Plan(ctx, tt.input, tt.planCtx)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if len(plan.Steps) != len(tt.wantSteps) {
				t.Fatalf("expected %d steps, got %d", len(tt.wantSteps), len(plan.Steps))
			}

			for i, step := range plan.Steps {
				if step.Type != tt.wantSteps[i] {
					t.Errorf("step %d: expected type %s, got %s", i, tt.wantSteps[i], step.Type)
				}

				if step.Type == planner.StepRespond {
					text := step.Params["text"].(string)
					if text != tt.wantText {
						t.Errorf("expected text %q, got %q", tt.wantText, text)
					}
				}

				if step.Type == planner.StepMemoryPut && tt.wantMemory != nil {
					key := step.Params["key"].(string)
					val := step.Params["value"].(string)
					if key != tt.wantMemory["key"] || val != tt.wantMemory["value"] {
						t.Errorf("expected memory put key=%q val=%q, got key=%q val=%q", tt.wantMemory["key"], tt.wantMemory["value"], key, val)
					}
				}
			}
		})
	}
}
