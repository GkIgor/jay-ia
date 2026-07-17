package planner

import (
	"context"
	"testing"

	"github.com/GkIgor/jay-ia/core/internal/llm"
	"github.com/GkIgor/jay-ia/core/internal/tools"
)

func TestLLMPlannerSimpleRouting(t *testing.T) {
	tb := tools.NewToolBus()
	mockClient := llm.NewMockClient()
	p := NewLLMPlanner(mockClient, tb)

	ctx := context.Background()
	planCtx := PlanningContext{}

	// Se começar com /, deve chamar o SimplePlanner (ex: /ping)
	plan, err := p.Plan(ctx, "/ping", planCtx)
	if err != nil {
		t.Fatalf("failed to plan: %v", err)
	}
	if len(plan.Steps) != 1 || plan.Steps[0].Type != StepRespond || plan.Steps[0].Params["text"] != "pong" {
		t.Errorf("expected pong response plan, got: %+v", plan)
	}
}

func TestLLMPlannerBuildPlan(t *testing.T) {
	// Test respond plan
	resp := &llm.Response{
		Text: "Hello there!",
	}
	plan := BuildPlan(resp)
	if len(plan.Steps) != 1 || plan.Steps[0].Type != StepRespond || plan.Steps[0].Params["text"] != "Hello there!" {
		t.Errorf("expected respond step, got: %+v", plan)
	}

	// Test function call plan
	resp = &llm.Response{
		FunctionCalls: []llm.FunctionCall{
			{Name: "fs.read_file", Args: map[string]any{"path": "test.txt"}},
		},
	}
	plan = BuildPlan(resp)
	if len(plan.Steps) != 1 || plan.Steps[0].Type != StepToolExecute || plan.Steps[0].Params["tool"] != "fs.read_file" {
		t.Errorf("expected tool execute step, got: %+v", plan)
	}
}
