package tools

import (
	"context"
	"fmt"
	"sync"
)

// Provider é responsável por gerenciar e listar um grupo de ferramentas.
type Provider interface {
	Name() string
	ListTools(ctx context.Context) ([]Tool, error)
}

// ToolBus gerencia todos os providers e indexa as ferramentas para despacho rápido O(1).
type ToolBus struct {
	mu        sync.RWMutex
	providers []Provider
	tools     map[string]Tool
}

// NewToolBus cria um novo barramento de ferramentas.
func NewToolBus() *ToolBus {
	return &ToolBus{
		providers: make([]Provider, 0),
		tools:     make(map[string]Tool),
	}
}

// RegisterProvider registra um novo provedor de ferramentas e reconstrói o cache interno.
func (tb *ToolBus) RegisterProvider(p Provider) {
	tb.mu.Lock()
	defer tb.mu.Unlock()
	tb.providers = append(tb.providers, p)
	tb.rebuildCacheLocked()
}

func (tb *ToolBus) rebuildCacheLocked() {
	newTools := make(map[string]Tool)
	ctx := context.Background()
	for _, p := range tb.providers {
		toolsList, err := p.ListTools(ctx)
		if err != nil {
			// Logar erro e continuar para não quebrar outros provedores válidos
			continue
		}
		for _, t := range toolsList {
			newTools[t.Describe().Name] = t
		}
	}
	tb.tools = newTools
}

// Execute despacha a execução de uma ferramenta pelo seu nome de forma rápida.
func (tb *ToolBus) Execute(ctx context.Context, name string, req Request) (Result, error) {
	tb.mu.RLock()
	tool, exists := tb.tools[name]
	tb.mu.RUnlock()

	if !exists {
		return Result{Success: false, Error: "tool not found"}, fmt.Errorf("tool %s not found", name)
	}

	return tool.Execute(ctx, req)
}

// GetTool recupera uma ferramenta pelo nome do cache de ferramentas.
func (tb *ToolBus) GetTool(name string) (Tool, bool) {
	tb.mu.RLock()
	defer tb.mu.RUnlock()
	t, ok := tb.tools[name]
	return t, ok
}

// ListAvailableTools expõe as definições de todas as ferramentas catalogadas.
func (tb *ToolBus) ListAvailableTools() []Definition {
	tb.mu.RLock()
	defer tb.mu.RUnlock()

	defs := make([]Definition, 0, len(tb.tools))
	for _, t := range tb.tools {
		defs = append(defs, t.Describe())
	}
	return defs
}
