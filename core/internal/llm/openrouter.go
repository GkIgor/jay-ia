package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/GkIgor/jay-ia/core/internal/tools"
)

type OpenRouterClient struct {
	apiKey     string
	model      string
	baseURL    string
	httpClient *http.Client
}

func NewOpenRouterClient(apiKey, model string) (*OpenRouterClient, error) {
	if apiKey == "" {
		return nil, fmt.Errorf("api key is required for openrouter provider")
	}
	if model == "" {
		return nil, fmt.Errorf("model is required for openrouter provider")
	}
	return &OpenRouterClient{
		apiKey:  apiKey,
		model:   model,
		baseURL: "https://openrouter.ai/api/v1",
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}, nil
}

type openRouterRequest struct {
	Model    string              `json:"model"`
	Messages []openRouterMsg     `json:"messages"`
	Tools    []openRouterToolDef `json:"tools,omitempty"`
}

type openRouterMsg struct {
	Role       string               `json:"role"`
	Content    string               `json:"content,omitempty"`
	ToolCalls  []openRouterToolCall `json:"tool_calls,omitempty"`
	ToolCallID string               `json:"tool_call_id,omitempty"`
	Name       string               `json:"name,omitempty"`
}

type openRouterToolCall struct {
	ID       string             `json:"id"`
	Type     string             `json:"type"`
	Function openRouterFunction `json:"function"`
}

type openRouterFunction struct {
	Name      string `json:"name"`
	Arguments string `json:"arguments"`
}

type openRouterToolDef struct {
	Type     string                  `json:"type"`
	Function openRouterFunctionProto `json:"function"`
}

type openRouterFunctionProto struct {
	Name        string            `json:"name"`
	Description string            `json:"description,omitempty"`
	Parameters  *openRouterSchema `json:"parameters,omitempty"`
}

type openRouterSchema struct {
	Type       string                     `json:"type"`
	Properties map[string]*openRouterProp `json:"properties,omitempty"`
	Required   []string                   `json:"required,omitempty"`
}

type openRouterProp struct {
	Type        string `json:"type"`
	Description string `json:"description,omitempty"`
}

type openRouterResponse struct {
	Choices []struct {
		Message struct {
			Role      string               `json:"role"`
			Content   string               `json:"content"`
			ToolCalls []openRouterToolCall `json:"tool_calls"`
		} `json:"message"`
	} `json:"choices"`
	Error *struct {
		Message string `json:"message"`
		Code    any    `json:"code"`
	} `json:"error,omitempty"`
}

func (o *OpenRouterClient) GenerateContent(ctx context.Context, history []Message, availableTools []tools.Definition) (*Response, error) {
	reqMsgs, err := convertToOpenRouterMessages(history)
	if err != nil {
		return nil, err
	}

	reqTools := convertToOpenRouterTools(availableTools)

	reqBody := openRouterRequest{
		Model:    o.model,
		Messages: reqMsgs,
		Tools:    reqTools,
	}

	apiResp, err := o.sendRequest(ctx, reqBody)
	if err != nil {
		return nil, err
	}

	return mapOpenRouterResponse(apiResp)
}

func convertToOpenRouterMessages(history []Message) ([]openRouterMsg, error) {
	reqMsgs := make([]openRouterMsg, 0, len(history))

	for _, msg := range history {
		switch msg.Role {
		case RoleUser:
			if len(msg.Parts) > 0 {
				reqMsgs = append(reqMsgs, openRouterMsg{
					Role:    "user",
					Content: msg.Parts[0].Text,
				})
			}

		case RoleModel:
			if len(msg.Parts) > 0 {
				part := msg.Parts[0]

				if part.FunctionCall != nil {
					argsBytes, err := json.Marshal(part.FunctionCall.Args)
					if err != nil {
						return nil, fmt.Errorf("failed to marshal function call args: %w", err)
					}
					callID := part.FunctionCall.ID
					if callID == "" {
						callID = "call_" + part.FunctionCall.Name
					}
					reqMsgs = append(reqMsgs, openRouterMsg{
						Role: "assistant",
						ToolCalls: []openRouterToolCall{
							{
								ID:   callID,
								Type: "function",
								Function: openRouterFunction{
									Name:      part.FunctionCall.Name,
									Arguments: string(argsBytes),
								},
							},
						},
					})
				} else {
					reqMsgs = append(reqMsgs, openRouterMsg{
						Role:    "assistant",
						Content: part.Text,
					})
				}
			}

		case RoleFunction:
			if len(msg.Parts) > 0 {
				part := msg.Parts[0]
				var respStr string
				if m, ok := part.FunctionResp.Response.(map[string]interface{}); ok {
					b, _ := json.Marshal(m)
					respStr = string(b)
				} else if s, ok := part.FunctionResp.Response.(string); ok {
					respStr = s
				} else {
					b, _ := json.Marshal(part.FunctionResp.Response)
					respStr = string(b)
				}

				reqMsgs = append(reqMsgs, openRouterMsg{
					Role:       "tool",
					Content:    respStr,
					ToolCallID: part.FunctionResp.ID,
					Name:       part.FunctionResp.Name,
				})
			}
		}
	}
	return reqMsgs, nil
}

func convertToOpenRouterTools(availableTools []tools.Definition) []openRouterToolDef {
	if len(availableTools) == 0 {
		return nil
	}

	reqTools := make([]openRouterToolDef, 0, len(availableTools))
	for _, t := range availableTools {
		properties := make(map[string]*openRouterProp)
		var required []string

		for _, p := range t.Parameters {
			properties[p.Name] = &openRouterProp{
				Type:        p.Type,
				Description: p.Description,
			}
			if p.Required {
				required = append(required, p.Name)
			}
		}

		var paramsSchema *openRouterSchema
		if len(properties) > 0 {
			paramsSchema = &openRouterSchema{
				Type:       "object",
				Properties: properties,
				Required:   required,
			}
		}

		reqTools = append(reqTools, openRouterToolDef{
			Type: "function",
			Function: openRouterFunctionProto{
				Name:        t.Name,
				Description: t.Description,
				Parameters:  paramsSchema,
			},
		})
	}
	return reqTools
}

func (o *OpenRouterClient) sendRequest(ctx context.Context, reqBody openRouterRequest) (*openRouterResponse, error) {
	reqBytes, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", o.baseURL+"/chat/completions", bytes.NewReader(reqBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to create http request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+o.apiKey)
	req.Header.Set("HTTP-Referer", "https://github.com/GkIgor/jay-ia")
	req.Header.Set("X-Title", "Jay IA")

	resp, err := o.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("http request failed: %w", err)
	}
	defer resp.Body.Close()

	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("openrouter api returned status %d: %s", resp.StatusCode, string(respBytes))
	}

	var apiResp openRouterResponse
	if err := json.Unmarshal(respBytes, &apiResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if apiResp.Error != nil {
		return nil, fmt.Errorf("openrouter error: %s", apiResp.Error.Message)
	}

	return &apiResp, nil
}

func mapOpenRouterResponse(apiResp *openRouterResponse) (*Response, error) {
	if len(apiResp.Choices) == 0 {
		return nil, fmt.Errorf("openrouter returned empty choices")
	}

	choice := apiResp.Choices[0]
	var result Response
	result.Text = choice.Message.Content

	for _, tc := range choice.Message.ToolCalls {
		if tc.Type == "function" {
			var args map[string]interface{}
			if err := json.Unmarshal([]byte(tc.Function.Arguments), &args); err != nil {
				return nil, fmt.Errorf("failed to unmarshal tool arguments: %w", err)
			}
			result.FunctionCalls = append(result.FunctionCalls, FunctionCall{
				ID:   tc.ID,
				Name: tc.Function.Name,
				Args: args,
			})
		}
	}

	return &result, nil
}
