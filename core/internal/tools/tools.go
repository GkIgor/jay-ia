package tools

import (
	"context"
)

// Parameter descreve uma entrada esperada pela ferramenta.
type Parameter struct {
	Name        string `json:"name"`
	Type        string `json:"type"` // "string", "number", "boolean"
	Description string `json:"description"`
	Required    bool   `json:"required"`
}

// Definition define os metadados descritivos da ferramenta.
type Definition struct {
	Name        string      `json:"name"`
	Description string      `json:"description"`
	Parameters  []Parameter `json:"parameters,omitempty"`
	Permissions []string    `json:"permissions,omitempty"`
}

// Result define o output final da ferramenta.
type Result struct {
	Success bool   `json:"success"`
	Output  any    `json:"output,omitempty"`
	Error   string `json:"error,omitempty"`
}

// ProgressState representa o estágio atual da execução.
type ProgressState string

const (
	ProgressStarted  ProgressState = "started"
	ProgressRunning  ProgressState = "running"
	ProgressFinished ProgressState = "finished"
)

// ProgressUpdate transmite informações sobre o progresso parcial de uma ferramenta.
type ProgressUpdate struct {
	State   ProgressState `json:"state"`
	Percent float64       `json:"percent,omitempty"`
	Message string        `json:"message,omitempty"`
}

// Request envelopa todos os parâmetros necessários para execução da ferramenta.
type Request struct {
	Args     map[string]any
	Progress chan<- ProgressUpdate
}

// Tool é a interface comum que toda capacidade executável deve implementar.
type Tool interface {
	Describe() Definition
	Execute(ctx context.Context, req Request) (Result, error)
}
