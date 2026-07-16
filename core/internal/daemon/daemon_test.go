package daemon

import (
	"context"
	"encoding/json"
	"net"
	"path/filepath"
	"testing"
	"time"

	"github.com/GkIgor/jay-ia/sdk/ipc"
)

func TestDaemonPermissionFlow(t *testing.T) {
	// Isola a socket do IPC criando em diretório temporário
	tempDir := t.TempDir()
	t.Setenv("XDG_RUNTIME_DIR", tempDir)

	d, err := New()
	if err != nil {
		t.Fatalf("Failed to create daemon: %v", err)
	}

	if err := d.ipcServer.Start(); err != nil {
		t.Fatalf("Failed to start IPC Server: %v", err)
	}
	defer d.ipcServer.Stop()

	// Dá um pequeno tempo para o socket UNIX inicializar
	time.Sleep(50 * time.Millisecond)

	socketPath := filepath.Join(tempDir, "jay", "jay.sock")
	conn, err := net.Dial("unix", socketPath)
	if err != nil {
		t.Fatalf("Failed to connect mock client to socket: %v", err)
	}
	defer conn.Close()

	// Canal para colher resposta da rotina de RequestPermission
	allowedChan := make(chan bool)
	errChan := make(chan error)

	go func() {
		allowed, err := d.RequestPermission(context.Background(), "fs.write_file", "fs.write")
		if err != nil {
			errChan <- err
			return
		}
		allowedChan <- allowed
	}()

	// Mock Client: Lê o evento de broadcast disparado pelo Core
	decoder := json.NewDecoder(conn)
	var eventMsg ipc.Message
	if err := decoder.Decode(&eventMsg); err != nil {
		t.Fatalf("Failed to decode permission request event: %v", err)
	}

	if eventMsg.Type != "request.permission" {
		t.Errorf("Expected event message type 'request.permission', got %q", eventMsg.Type)
	}

	data, ok := eventMsg.Payload.(map[string]any)
	if !ok {
		t.Fatalf("Expected event payload to be map[string]any, got %T", eventMsg.Payload)
	}

	refID, _ := data["ref_id"].(string)
	permission, _ := data["permission"].(string)
	prompt, _ := data["prompt"].(string)

	if refID == "" {
		t.Error("ref_id should not be empty")
	}
	if permission != "fs.write" {
		t.Errorf("Expected permission 'fs.write', got %q", permission)
	}
	expectedPrompt := "Jay quer executar a ferramenta 'fs.write_file' que exige a permissão 'fs.write'. Permitir?"
	if prompt != expectedPrompt {
		t.Errorf("Expected prompt %q, got %q", expectedPrompt, prompt)
	}

	// Mock Client: Devolve a aprovação
	respMsg := ipc.Message{
		Type: "permission.response",
		Payload: map[string]any{
			"ref_id":   refID,
			"allowed":  true,
			"modality": "keyboard",
		},
	}

	encoder := json.NewEncoder(conn)
	if err := encoder.Encode(respMsg); err != nil {
		t.Fatalf("Failed to write permission response: %v", err)
	}

	// Verifica se a rotina bloqueante do Core foi liberada e retornou true
	select {
	case allowed := <-allowedChan:
		if !allowed {
			t.Error("Expected RequestPermission to return true (allowed)")
		}
	case err := <-errChan:
		t.Fatalf("RequestPermission failed: %v", err)
	case <-time.After(1 * time.Second):
		t.Fatal("Timeout waiting for RequestPermission to unblock")
	}
}
