package tools

import (
	"context"
	"sync"
)

// NativeProvider implementa Provider para gerenciar ferramentas estáticas locais compiladas com o Core.
type NativeProvider struct {
	mu    sync.RWMutex
	tools []Tool
}

// NewNativeProvider inicializa o provedor nativo.
func NewNativeProvider() *NativeProvider {
	return &NativeProvider{
		tools: make([]Tool, 0),
	}
}

// Name retorna a identificação do provedor.
func (np *NativeProvider) Name() string {
	return "native"
}

// RegisterTool adiciona uma ferramenta estática ao provedor.
func (np *NativeProvider) RegisterTool(t Tool) {
	np.mu.Lock()
	defer np.mu.Unlock()
	np.tools = append(np.tools, t)
}

// ListTools retorna a listagem de ferramentas registradas.
func (np *NativeProvider) ListTools(ctx context.Context) ([]Tool, error) {
	np.mu.RLock()
	defer np.mu.RUnlock()

	// Retorna uma cópia do slice para evitar problemas de concorrência externo
	list := make([]Tool, len(np.tools))
	copy(list, np.tools)
	return list, nil
}
