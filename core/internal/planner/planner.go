package planner

import (
	"context"
)

type StepType string

const (
	StepRespond       StepType = "respond"
	StepMemoryPut     StepType = "memory_put"
	StepHumanEscalate StepType = "human_escalate"
)

type Step struct {
	Type   StepType       `json:"type"`
	Params map[string]any `json:"params"`
}

type Plan struct {
	Steps []Step `json:"steps"`
}

// PlanningContext contains only pure data queried by the Core
type PlanningContext struct {
	WorkingMemory map[string]string
}

type Planner interface {
	Plan(ctx context.Context, input string, planCtx PlanningContext) (*Plan, error)
}
