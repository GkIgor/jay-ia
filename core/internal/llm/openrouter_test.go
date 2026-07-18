package llm

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/GkIgor/jay-ia/core/internal/tools"
)

func TestOpenRouterClient(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer test-api-key" {
			t.Errorf("expected Authorization header Bearer test-api-key, got %s", r.Header.Get("Authorization"))
		}

		var reqBody openRouterRequest
		if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
			t.Fatalf("failed to decode request body: %v", err)
		}

		if reqBody.Model != "google/gemini-2.5-flash" {
			t.Errorf("expected model google/gemini-2.5-flash, got %s", reqBody.Model)
		}

		resp := openRouterResponse{}
		resp.Choices = []struct {
			Message struct {
				Role      string               `json:"role"`
				Content   string               `json:"content"`
				ToolCalls []openRouterToolCall `json:"tool_calls"`
			} `json:"message"`
		}{
			{
				Message: struct {
					Role      string               `json:"role"`
					Content   string               `json:"content"`
					ToolCalls []openRouterToolCall `json:"tool_calls"`
				}{
					Role:    "assistant",
					Content: "Hello world!",
				},
			},
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client, err := NewOpenRouterClient("test-api-key", "google/gemini-2.5-flash")
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}

	client.baseURL = server.URL
	client.httpClient = server.Client()

	history := []Message{
		{
			Role:  RoleUser,
			Parts: []Part{{Text: "Hi"}},
		},
	}

	resp, err := client.GenerateContent(context.Background(), history, nil)
	if err != nil {
		t.Fatalf("failed to generate content: %v", err)
	}

	if resp.Text != "Hello world!" {
		t.Errorf("expected response text 'Hello world!', got %s", resp.Text)
	}
}

func TestOpenRouterClientToolCall(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var reqBody openRouterRequest
		if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
			t.Fatalf("failed to decode request: %v", err)
		}

		resp := openRouterResponse{}
		resp.Choices = []struct {
			Message struct {
				Role      string               `json:"role"`
				Content   string               `json:"content"`
				ToolCalls []openRouterToolCall `json:"tool_calls"`
			} `json:"message"`
		}{
			{
				Message: struct {
					Role      string               `json:"role"`
					Content   string               `json:"content"`
					ToolCalls []openRouterToolCall `json:"tool_calls"`
				}{
					Role: "assistant",
					ToolCalls: []openRouterToolCall{
						{
							ID:   "call_xyz123",
							Type: "function",
							Function: openRouterFunction{
								Name:      "fs.list_dir",
								Arguments: `{"path":"/"}`,
							},
						},
					},
				},
			},
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client, err := NewOpenRouterClient("test-api-key", "google/gemini-2.5-flash")
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}
	client.baseURL = server.URL
	client.httpClient = server.Client()

	availableTools := []tools.Definition{
		{
			Name:        "fs.list_dir",
			Description: "list directory",
			Parameters: []tools.Parameter{
				{
					Name:        "path",
					Type:        "string",
					Description: "path to list",
					Required:    true,
				},
			},
		},
	}

	resp, err := client.GenerateContent(context.Background(), nil, availableTools)
	if err != nil {
		t.Fatalf("failed to generate content: %v", err)
	}

	if len(resp.FunctionCalls) != 1 {
		t.Fatalf("expected 1 function call, got %d", len(resp.FunctionCalls))
	}

	fc := resp.FunctionCalls[0]
	if fc.Name != "fs.list_dir" {
		t.Errorf("expected tool name 'fs.list_dir', got %s", fc.Name)
	}

	pathVal, ok := fc.Args["path"].(string)
	if !ok || pathVal != "/" {
		t.Errorf("expected arg path to be '/', got %v", fc.Args["path"])
	}
}

func TestOpenRouterClientHistoryConversion(t *testing.T) {
	history := []Message{
		{
			Role:  RoleUser,
			Parts: []Part{{Text: "List files"}},
		},
		{
			Role: RoleModel,
			Parts: []Part{
				{
					FunctionCall: &FunctionCall{
						ID:   "call_unique_456",
						Name: "fs.list_dir",
						Args: map[string]interface{}{"path": "/usr"},
					},
				},
			},
		},
		{
			Role: RoleFunction,
			Parts: []Part{
				{
					FunctionResp: &FunctionResponse{
						ID:       "call_unique_456",
						Name:     "fs.list_dir",
						Response: "file1, file2",
					},
				},
			},
		},
	}

	msgs, err := convertToOpenRouterMessages(history)
	if err != nil {
		t.Fatalf("failed to convert messages: %v", err)
	}

	if len(msgs) != 3 {
		t.Fatalf("expected 3 messages, got %d", len(msgs))
	}

	// Verifica a primeira mensagem (User)
	if msgs[0].Role != "user" || msgs[0].Content != "List files" {
		t.Errorf("invalid user message: %+v", msgs[0])
	}

	// Verifica a segunda mensagem (Assistant Call)
	if msgs[1].Role != "assistant" || len(msgs[1].ToolCalls) != 1 {
		t.Fatalf("invalid assistant message structure: %+v", msgs[1])
	}
	tc := msgs[1].ToolCalls[0]
	if tc.ID != "call_unique_456" || tc.Function.Name != "fs.list_dir" {
		t.Errorf("invalid tool call content: %+v", tc)
	}

	// Verifica a terceira mensagem (Tool Response)
	if msgs[2].Role != "tool" || msgs[2].ToolCallID != "call_unique_456" || msgs[2].Content != "file1, file2" {
		t.Errorf("invalid tool response message: %+v", msgs[2])
	}
}
