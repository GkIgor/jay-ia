package llm

import (
	"context"

	"github.com/GkIgor/jay-ia/core/internal/tools"
)

// Role representa o papel na conversa (user, model ou function)
type Role string

const (
	RoleUser     Role = "user"
	RoleModel    Role = "model"
	RoleFunction Role = "function"
)

// FunctionCall representa o pedido de chamada de ferramenta pela LLM
type FunctionCall struct {
	Name string                 `json:"name"`
	Args map[string]interface{} `json:"args"`
}

// FunctionResponse representa a resposta da ferramenta executada localmente
type FunctionResponse struct {
	Name     string      `json:"name"`
	Response interface{} `json:"response"`
}

// Part representa os elementos que compõem uma mensagem (texto ou chamada/resposta de ferramenta)
type Part struct {
	Text         string            `json:"text,omitempty"`
	FunctionCall *FunctionCall     `json:"function_call,omitempty"`
	FunctionResp *FunctionResponse `json:"function_resp,omitempty"`
}

// Message representa o elemento básico do histórico da conversa
type Message struct {
	Role  Role   `json:"role"`
	Parts []Part `json:"parts"`
}

// Response encapsula o retorno simplificado e agnóstico de qualquer LLM
type Response struct {
	Text          string         `json:"text,omitempty"`
	FunctionCalls []FunctionCall `json:"function_calls,omitempty"`
}

// Client define o contrato para que a Jay converse com modelos de linguagem de forma independente de SDK
type Client interface {
	GenerateContent(ctx context.Context, history []Message, availableTools []tools.Definition) (*Response, error)
}
