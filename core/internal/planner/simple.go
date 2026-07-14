package planner

import (
	"context"
	"fmt"
	"strings"
)

// SimplePlanner is a rule-based/CLI dispatcher for Phase 1
type SimplePlanner struct{}

func NewSimplePlanner() *SimplePlanner {
	return &SimplePlanner{}
}

func (p *SimplePlanner) Plan(ctx context.Context, input string, planCtx PlanningContext) (*Plan, error) {
	input = strings.TrimSpace(input)

	// Command Mode (starts with /)
	if strings.HasPrefix(input, "/") {
		parts := strings.SplitN(input, " ", 2)
		cmd := parts[0]

		switch cmd {
		case "/ping":
			return &Plan{
				Steps: []Step{
					{
						Type: StepRespond,
						Params: map[string]any{
							"text": "pong",
						},
					},
				},
			}, nil

		case "/memo":
			if len(parts) < 2 {
				return p.respondWithError("Usage: /memo key=value")
			}
			kv := strings.SplitN(parts[1], "=", 2)
			if len(kv) < 2 {
				return p.respondWithError("Usage: /memo key=value")
			}
			key := strings.TrimSpace(kv[0])
			val := strings.TrimSpace(kv[1])
			if key == "" || val == "" {
				return p.respondWithError("Usage: /memo key=value (key and value cannot be empty)")
			}

			return &Plan{
				Steps: []Step{
					{
						Type: StepMemoryPut,
						Params: map[string]any{
							"key":   key,
							"value": val,
						},
					},
					{
						Type: StepRespond,
						Params: map[string]any{
							"text": fmt.Sprintf("Saved '%s' to memory.", key),
						},
					},
				},
			}, nil

		case "/recall":
			if len(parts) < 2 {
				return p.respondWithError("Usage: /recall key")
			}
			key := strings.TrimSpace(parts[1])
			if key == "" {
				return p.respondWithError("Usage: /recall key")
			}

			val, ok := planCtx.WorkingMemory[key]
			if !ok {
				return &Plan{
					Steps: []Step{
						{
							Type: StepRespond,
							Params: map[string]any{
								"text": fmt.Sprintf("I don't remember anything for '%s'.", key),
							},
						},
					},
				}, nil
			}

			return &Plan{
				Steps: []Step{
					{
						Type: StepRespond,
						Params: map[string]any{
							"text": fmt.Sprintf("I remember: %s = %s", key, val),
						},
					},
				},
			}, nil

		case "/write":
			if len(parts) < 2 {
				return p.respondWithError("Usage: /write <path> <content>")
			}
			argsParts := strings.SplitN(parts[1], " ", 2)
			if len(argsParts) < 2 {
				return p.respondWithError("Usage: /write <path> <content>")
			}
			path := strings.TrimSpace(argsParts[0])
			content := strings.TrimSpace(argsParts[1])
			return &Plan{
				Steps: []Step{
					{
						Type: StepToolExecute,
						Params: map[string]any{
							"tool": "fs.write_file",
							"args": map[string]any{
								"path":    path,
								"content": content,
							},
						},
					},
					{
						Type: StepRespond,
						Params: map[string]any{
							"text": fmt.Sprintf("File written to %s.", path),
						},
					},
				},
			}, nil

		case "/read":
			if len(parts) < 2 {
				return p.respondWithError("Usage: /read <path>")
			}
			path := strings.TrimSpace(parts[1])
			return &Plan{
				Steps: []Step{
					{
						Type: StepToolExecute,
						Params: map[string]any{
							"tool": "fs.read_file",
							"args": map[string]any{
								"path": path,
							},
						},
					},
				},
			}, nil

		default:
			return &Plan{
				Steps: []Step{
					{
						Type: StepRespond,
						Params: map[string]any{
							"text": fmt.Sprintf("Unknown command: %s", cmd),
						},
					},
				},
			}, nil
		}
	}

	// Normal conversation mode (fallback echo)
	return &Plan{
		Steps: []Step{
			{
				Type: StepRespond,
				Params: map[string]any{
					"text": fmt.Sprintf("I received your message: \"%s\"", input),
				},
			},
		},
	}, nil
}

func (p *SimplePlanner) respondWithError(msg string) (*Plan, error) {
	return &Plan{
		Steps: []Step{
			{
				Type: StepRespond,
				Params: map[string]any{
					"text": msg,
				},
			},
		},
	}, nil
}
