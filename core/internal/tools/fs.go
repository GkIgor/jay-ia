package tools

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
)

// ReadFileTool lê o conteúdo de um arquivo específico.
type ReadFileTool struct{}

// Describe descreve a assinatura de entrada.
func (t ReadFileTool) Describe() Definition {
	return Definition{
		Name:        "fs.read_file",
		Description: "Lê o conteúdo completo de um arquivo local.",
		Parameters: []Parameter{
			{
				Name:        "path",
				Type:        "string",
				Description: "Caminho para o arquivo.",
				Required:    true,
			},
		},
	}
}

// Execute executa a leitura.
func (t ReadFileTool) Execute(ctx context.Context, req Request) (Result, error) {
	if req.Progress != nil {
		req.Progress <- ProgressUpdate{State: ProgressStarted, Message: "Lendo arquivo..."}
	}

	pathVal, ok := req.Args["path"]
	if !ok {
		return Result{Success: false, Error: "missing required argument: path"}, fmt.Errorf("missing path argument")
	}
	path, ok := pathVal.(string)
	if !ok {
		return Result{Success: false, Error: "argument 'path' must be a string"}, fmt.Errorf("invalid path type")
	}

	content, err := os.ReadFile(path)
	if err != nil {
		return Result{Success: false, Error: fmt.Sprintf("failed to read file: %v", err)}, err
	}

	if req.Progress != nil {
		req.Progress <- ProgressUpdate{State: ProgressFinished, Percent: 100, Message: "Leitura concluída"}
	}

	return Result{
		Success: true,
		Output:  string(content),
	}, nil
}

// WriteFileTool escreve ou sobrescreve conteúdo em um arquivo.
type WriteFileTool struct{}

// Describe descreve a assinatura de entrada.
func (t WriteFileTool) Describe() Definition {
	return Definition{
		Name:        "fs.write_file",
		Description: "Escreve ou sobrescreve o conteúdo de um arquivo local.",
		Parameters: []Parameter{
			{
				Name:        "path",
				Type:        "string",
				Description: "Caminho do arquivo de destino.",
				Required:    true,
			},
			{
				Name:        "content",
				Type:        "string",
				Description: "Conteúdo a ser escrito no arquivo.",
				Required:    true,
			},
		},
	}
}

// Execute executa a escrita garantindo a estrutura de diretórios pai.
func (t WriteFileTool) Execute(ctx context.Context, req Request) (Result, error) {
	if req.Progress != nil {
		req.Progress <- ProgressUpdate{State: ProgressStarted, Message: "Escrevendo arquivo..."}
	}

	pathVal, ok := req.Args["path"]
	if !ok {
		return Result{Success: false, Error: "missing required argument: path"}, fmt.Errorf("missing path argument")
	}
	path, ok := pathVal.(string)
	if !ok {
		return Result{Success: false, Error: "argument 'path' must be a string"}, fmt.Errorf("invalid path type")
	}

	contentVal, ok := req.Args["content"]
	if !ok {
		return Result{Success: false, Error: "missing required argument: content"}, fmt.Errorf("missing content argument")
	}
	content, ok := contentVal.(string)
	if !ok {
		return Result{Success: false, Error: "argument 'content' must be a string"}, fmt.Errorf("invalid content type")
	}

	// Garante que o diretório pai existe
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return Result{Success: false, Error: fmt.Sprintf("failed to create directories: %v", err)}, err
	}

	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		return Result{Success: false, Error: fmt.Sprintf("failed to write file: %v", err)}, err
	}

	if req.Progress != nil {
		req.Progress <- ProgressUpdate{State: ProgressFinished, Percent: 100, Message: "Escrita concluída"}
	}

	return Result{
		Success: true,
		Output:  "File written successfully",
	}, nil
}

// ListDirTool lista os arquivos e pastas em um caminho.
type ListDirTool struct{}

// Describe descreve a assinatura de entrada.
func (t ListDirTool) Describe() Definition {
	return Definition{
		Name:        "fs.list_dir",
		Description: "Lista o conteúdo de um diretório.",
		Parameters: []Parameter{
			{
				Name:        "path",
				Type:        "string",
				Description: "Caminho do diretório (opcional, padrão é o diretório atual).",
				Required:    false,
			},
		},
	}
}

// Execute executa a leitura do diretório.
func (t ListDirTool) Execute(ctx context.Context, req Request) (Result, error) {
	if req.Progress != nil {
		req.Progress <- ProgressUpdate{State: ProgressStarted, Message: "Listando diretório..."}
	}

	path := "."
	if pathVal, ok := req.Args["path"]; ok {
		if p, ok := pathVal.(string); ok {
			path = p
		}
	}

	entries, err := os.ReadDir(path)
	if err != nil {
		return Result{Success: false, Error: fmt.Sprintf("failed to list directory: %v", err)}, err
	}

	var list []string
	for _, entry := range entries {
		name := entry.Name()
		if entry.IsDir() {
			name += "/"
		}
		list = append(list, name)
	}

	if req.Progress != nil {
		req.Progress <- ProgressUpdate{State: ProgressFinished, Percent: 100, Message: "Listagem concluída"}
	}

	return Result{
		Success: true,
		Output:  list,
	}, nil
}
