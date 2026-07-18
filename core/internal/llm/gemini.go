package llm

import (
	"context"
	"fmt"
	"os"

	"github.com/GkIgor/jay-ia/core/internal/tools"
	"google.golang.org/genai"
)

// GeminiClient encapsula o cliente oficial do SDK do Google GenAI
type GeminiClient struct {
	client *genai.Client
}

func NewGeminiClient(apiKey string) (*GeminiClient, error) {
	ctx := context.Background()
	c, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey: apiKey,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create genai client: %w", err)
	}
	return &GeminiClient{client: c}, nil
}

func (g *GeminiClient) GenerateContent(ctx context.Context, history []Message, availableTools []tools.Definition) (*Response, error) {
	// 1. Converte o histórico de mensagens puro do Core para o formato esperado pelo SDK
	contents := make([]*genai.Content, 0, len(history))
	for _, msg := range history {
		role := string(msg.Role)
		// Normalização de papéis (Gemini aceita 'user', 'model', 'function')
		if role == "model" {
			role = "model"
		} else if role == "user" {
			role = "user"
		} else if role == "function" {
			role = "function"
		}

		parts := make([]*genai.Part, 0, len(msg.Parts))
		for _, part := range msg.Parts {
			sdkPart := &genai.Part{}
			if part.Text != "" {
				sdkPart.Text = part.Text
			} else if part.FunctionCall != nil {
				argsMap := make(map[string]interface{})
				for k, v := range part.FunctionCall.Args {
					argsMap[k] = v
				}
				sdkPart.FunctionCall = &genai.FunctionCall{
					Name: part.FunctionCall.Name,
					Args: argsMap,
				}
			} else if part.FunctionResp != nil {
				respMap := make(map[string]interface{})
				if m, ok := part.FunctionResp.Response.(map[string]interface{}); ok {
					respMap = m
				} else {
					respMap["result"] = part.FunctionResp.Response
				}
				sdkPart.FunctionResponse = &genai.FunctionResponse{
					Name:     part.FunctionResp.Name,
					Response: respMap,
				}
			}
			parts = append(parts, sdkPart)
		}

		contents = append(contents, &genai.Content{
			Role:  role,
			Parts: parts,
		})
	}

	// 2. Converte as ferramentas do ToolBus do Core para o formato de ferramentas (Tools) do SDK
	var sdkTools []*genai.Tool
	if len(availableTools) > 0 {
		decls := make([]*genai.FunctionDeclaration, 0, len(availableTools))
		for _, t := range availableTools {
			properties := make(map[string]*genai.Schema)
			var required []string

			for _, p := range t.Parameters {
				var pType genai.Type
				switch p.Type {
				case "string":
					pType = genai.TypeString
				case "number":
					pType = genai.TypeNumber
				case "boolean":
					pType = genai.TypeBoolean
				default:
					pType = genai.TypeString
				}

				properties[p.Name] = &genai.Schema{
					Type:        pType,
					Description: p.Description,
				}

				if p.Required {
					required = append(required, p.Name)
				}
			}

			var paramsSchema *genai.Schema
			if len(properties) > 0 {
				paramsSchema = &genai.Schema{
					Type:       genai.TypeObject,
					Properties: properties,
					Required:   required,
				}
			}

			decls = append(decls, &genai.FunctionDeclaration{
				Name:        t.Name,
				Description: t.Description,
				Parameters:  paramsSchema,
			})
		}

		sdkTools = append(sdkTools, &genai.Tool{
			FunctionDeclarations: decls,
		})
	}

	// 3. Executa a chamada do modelo de linguagem (com base no GEMINI_MODEL ou fallback para gemini-3.5-flash)
	model := os.Getenv("GEMINI_MODEL")
	if model == "" {
		model = "gemini-3.5-flash"
	}
	config := &genai.GenerateContentConfig{
		Tools: sdkTools,
	}

	resp, err := g.client.Models.GenerateContent(ctx, model, contents, config)
	if err != nil {
		return nil, fmt.Errorf("gemini api call failed: %w", err)
	}

	// 4. Mapeia a resposta de volta ao tipo puro agnóstico Response
	var result Response
	if len(resp.Candidates) > 0 && resp.Candidates[0].Content != nil {
		content := resp.Candidates[0].Content
		for _, part := range content.Parts {
			if part.Text != "" {
				result.Text = part.Text
			}
			if part.FunctionCall != nil {
				args := make(map[string]interface{})
				for k, v := range part.FunctionCall.Args {
					args[k] = v
				}
				result.FunctionCalls = append(result.FunctionCalls, FunctionCall{
					Name: part.FunctionCall.Name,
					Args: args,
				})
			}
		}
	}

	return &result, nil
}
