package conversation

import (
	"sync"

	"github.com/GkIgor/jay-ia/core/internal/llm"
)

// Manager gerencia o histórico da sessão de forma estruturada e concorrente
type Manager struct {
	mu      sync.RWMutex
	history []llm.Message
}

func NewManager() *Manager {
	return &Manager{
		history: make([]llm.Message, 0),
	}
}

// GetHistory retorna uma cópia profunda/estática do histórico atual de mensagens
func (m *Manager) GetHistory() []llm.Message {
	m.mu.RLock()
	defer m.mu.RUnlock()

	historyCopy := make([]llm.Message, len(m.history))
	copy(historyCopy, m.history)
	return historyCopy
}

// AddUserMessage anexa uma nova mensagem vinda do usuário
func (m *Manager) AddUserMessage(text string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.history = append(m.history, llm.Message{
		Role: llm.RoleUser,
		Parts: []llm.Part{
			{Text: text},
		},
	})
}

// AddModelMessage anexa uma nova resposta em texto gerada pela LLM
func (m *Manager) AddModelMessage(text string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.history = append(m.history, llm.Message{
		Role: llm.RoleModel,
		Parts: []llm.Part{
			{Text: text},
		},
	})
}

// AddFunctionCall anexa a solicitação de execução de ferramenta emitida pelo modelo.
func (m *Manager) AddFunctionCall(id string, name string, args map[string]interface{}, metadata map[string]string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.history = append(m.history, llm.Message{
		Role: llm.RoleModel,
		Parts: []llm.Part{
			{
				FunctionCall: &llm.FunctionCall{
					ID:       id,
					Name:     name,
					Args:     args,
					Metadata: metadata,
				},
				Metadata: metadata,
			},
		},
	})
}

// AddFunctionResponse anexa o resultado da execução da ferramenta local de volta ao histórico.
func (m *Manager) AddFunctionResponse(id string, name string, response interface{}) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.history = append(m.history, llm.Message{
		Role: llm.RoleFunction,
		Parts: []llm.Part{
			{
				FunctionResp: &llm.FunctionResponse{
					ID:       id,
					Name:     name,
					Response: response,
				},
			},
		},
	})
}

// Clear esvazia o histórico de conversa
func (m *Manager) Clear() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.history = make([]llm.Message, 0)
}
