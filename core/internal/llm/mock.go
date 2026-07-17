package llm

import (
	"context"

	"github.com/GkIgor/jay-ia/core/internal/tools"
)

// MockClient simula chamadas a modelos para testes locais livres de rede
type MockClient struct {
	ResponseFn func(history []Message) (*Response, error)
}

func NewMockClient() *MockClient {
	return &MockClient{}
}

func (m *MockClient) GenerateContent(ctx context.Context, history []Message, availableTools []tools.Definition) (*Response, error) {
	if m.ResponseFn != nil {
		return m.ResponseFn(history)
	}
	return &Response{Text: "Mock Response"}, nil
}
