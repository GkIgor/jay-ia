package commands

import (
	"encoding/json"
	"fmt"
	"net"
	"os"
	"path/filepath"

	"github.com/GkIgor/jay-ia/sdk/ipc"
)

// getSocketPath resolve o caminho do Unix Domain Socket do Jay daemon.
func getSocketPath() string {
	xdgRuntimeDir := os.Getenv("XDG_RUNTIME_DIR")
	if xdgRuntimeDir != "" {
		return filepath.Join(xdgRuntimeDir, "jay", "jay.sock")
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return "/tmp/jay/jay.sock"
	}
	return filepath.Join(home, ".jay", "jay.sock")
}

// sendRPC conecta ao socket Unix do daemon, envia o RequestEnvelope e aguarda o ResponseEnvelope.
func sendRPC(msgType ipc.MessageType, clientID string, payload any) (*ipc.ResponseEnvelope, error) {
	socketPath := getSocketPath()

	conn, err := net.Dial("unix", socketPath)
	if err != nil {
		return nil, fmt.Errorf("falha ao conectar ao daemon em %s: %w (certifique-se de que o 'jayd' está rodando)", socketPath, err)
	}
	defer conn.Close()

	if clientID == "" {
		clientID = "cli_user"
	}

	reqEnv, err := ipc.NewRequestEnvelope(msgType, clientID, payload)
	if err != nil {
		return nil, fmt.Errorf("falha ao montar envelope de requisição: %w", err)
	}

	encoder := json.NewEncoder(conn)
	decoder := json.NewDecoder(conn)

	if err := encoder.Encode(reqEnv); err != nil {
		return nil, fmt.Errorf("falha ao enviar requisição no socket: %w", err)
	}

	var respEnv ipc.ResponseEnvelope
	if err := decoder.Decode(&respEnv); err != nil {
		return nil, fmt.Errorf("falha ao receber resposta do socket: %w", err)
	}

	return &respEnv, nil
}

// getErrorMessage extrai a mensagem de erro formatada do ResponseEnvelope.
func getErrorMessage(resp *ipc.ResponseEnvelope) string {
	if resp.Error != nil && resp.Error.Message != "" {
		return resp.Error.Message
	}
	return fmt.Sprintf("código de erro %d", resp.Status)
}
