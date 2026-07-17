package llm

import (
	"context"
	"strings"

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

	// Regras para simular inteligentemente chamadas de ferramentas locais offline
	if len(history) > 0 {
		lastMsg := history[len(history)-1]
		if len(lastMsg.Parts) > 0 && lastMsg.Parts[0].Text != "" {
			text := strings.ToLower(lastMsg.Parts[0].Text)
			// Se o usuário pedir para criar/escrever o arquivo nota.txt
			if (strings.Contains(text, "escreva") || strings.Contains(text, "crie") || strings.Contains(text, "arquivo")) && strings.Contains(text, "nota.txt") {
				// Verifica se a ferramenta já foi executada e o resultado está no histórico
				hasResult := false
				for _, msg := range history {
					if msg.Role == RoleFunction {
						hasResult = true
						break
					}
				}

				if !hasResult {
					// Primeiro turno: IA solicita chamada de ferramenta
					return &Response{
						FunctionCalls: []FunctionCall{
							{
								Name: "fs.write_file",
								Args: map[string]any{
									"path":    "nota.txt",
									"content": "Aprovado",
								},
							},
						},
					}, nil
				} else {
					// Segundo turno (pós-execução): IA retorna o texto de sucesso final
					return &Response{
						Text: "Arquivo 'nota.txt' criado com sucesso com o texto 'Aprovado'.",
					}, nil
				}
			}
		}
	}

	return &Response{Text: "Mock Response"}, nil
}
