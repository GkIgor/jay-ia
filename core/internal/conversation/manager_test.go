package conversation

import (
	"testing"

	"github.com/GkIgor/jay-ia/core/internal/llm"
)

func TestConversationManager(t *testing.T) {
	m := NewManager()

	if len(m.GetHistory()) != 0 {
		t.Errorf("expected empty history initially")
	}

	m.AddUserMessage("Hello")
	m.AddModelMessage("Hi there")
	m.AddFunctionCall("call-1", "fs.read_file", map[string]any{"path": "file.txt"}, nil)
	m.AddFunctionResponse("call-1", "fs.read_file", "file contents")

	h := m.GetHistory()
	if len(h) != 4 {
		t.Fatalf("expected 4 messages in history, got %d", len(h))
	}

	if h[0].Role != llm.RoleUser || h[0].Parts[0].Text != "Hello" {
		t.Errorf("unexpected first message: %+v", h[0])
	}

	if h[1].Role != llm.RoleModel || h[1].Parts[0].Text != "Hi there" {
		t.Errorf("unexpected second message: %+v", h[1])
	}

	if h[2].Role != llm.RoleModel || h[2].Parts[0].FunctionCall == nil || h[2].Parts[0].FunctionCall.Name != "fs.read_file" {
		t.Errorf("unexpected third message: %+v", h[2])
	}

	if h[3].Role != llm.RoleFunction || h[3].Parts[0].FunctionResp == nil || h[3].Parts[0].FunctionResp.Name != "fs.read_file" {
		t.Errorf("unexpected fourth message: %+v", h[3])
	}

	m.Clear()
	if len(m.GetHistory()) != 0 {
		t.Errorf("expected history to be empty after clear")
	}
}
