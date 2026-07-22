package daemon

import (
	"encoding/json"
	"net"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/GkIgor/jay-ia/sdk/ipc"
)

func TestDaemon_Bootstrap_Success(t *testing.T) {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test_jay.db")

	t.Setenv("LLM_PROVIDER", "mock")

	d, err := NewDaemon(dbPath)
	if err != nil {
		t.Fatalf("falha ao inicializar Daemon: %v", err)
	}
	defer d.Stop()

	if d.Router() == nil {
		t.Fatalf("esperava Router RPC instanciado no Daemon")
	}
}

func TestDaemon_UnknownLLMProvider(t *testing.T) {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test_jay.db")

	t.Setenv("LLM_PROVIDER", "provedor_desconhecido_invalido")

	_, err := NewDaemon(dbPath)
	if err == nil {
		t.Fatalf("esperava erro ao informar provedor LLM desconhecido, obteve nil")
	}
}

func TestDaemon_IPC_EndToEnd(t *testing.T) {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "e2e_jay.db")
	socketDir := filepath.Join(tempDir, "jay_sock")

	_ = os.MkdirAll(socketDir, 0700)
	socketPath := filepath.Join(socketDir, "jay", "jay.sock")

	t.Setenv("XDG_RUNTIME_DIR", socketDir)
	t.Setenv("LLM_PROVIDER", "mock")

	d, err := NewDaemon(dbPath)
	if err != nil {
		t.Fatalf("falha ao instanciar Daemon: %v", err)
	}

	go func() {
		_ = d.Start()
	}()
	defer d.Stop()

	// Aguarda o listener do servidor de socket subir
	time.Sleep(50 * time.Millisecond)

	// Conecta cliente de teste ao socket Unix
	conn, err := net.Dial("unix", socketPath)
	if err != nil {
		t.Fatalf("falha ao conectar ao socket Unix %s: %v", socketPath, err)
	}
	defer conn.Close()

	encoder := json.NewEncoder(conn)
	decoder := json.NewDecoder(conn)

	// 1. Envia MsgRegisterClient (100)
	regReqPayload := ipc.RegisterClientRequest{ClientID: "client_cpp_e2e"}
	regReqEnv, _ := ipc.NewRequestEnvelope(ipc.MsgRegisterClient, "client_cpp_e2e", regReqPayload)

	if err := encoder.Encode(regReqEnv); err != nil {
		t.Fatalf("falha ao enviar envelope de registro: %v", err)
	}

	var regRespEnv ipc.ResponseEnvelope
	if err := decoder.Decode(&regRespEnv); err != nil {
		t.Fatalf("falha ao receber resposta do registro: %v", err)
	}

	if regRespEnv.Status != ipc.ErrSuccess {
		t.Fatalf("esperava status 0 no registro, obteve %d", regRespEnv.Status)
	}

	// 2. Envia MsgCreateChat (200)
	chatReqPayload := ipc.CreateChatRequest{Title: "Chat E2E Integração"}
	chatReqEnv, _ := ipc.NewRequestEnvelope(ipc.MsgCreateChat, "client_cpp_e2e", chatReqPayload)

	if err := encoder.Encode(chatReqEnv); err != nil {
		t.Fatalf("falha ao enviar envelope de criação de chat: %v", err)
	}

	var chatRespEnv ipc.ResponseEnvelope
	if err := decoder.Decode(&chatRespEnv); err != nil {
		t.Fatalf("falha ao receber resposta de criação de chat: %v", err)
	}

	if chatRespEnv.Status != ipc.ErrSuccess {
		t.Fatalf("esperava status 0 na criação do chat, obteve %d", chatRespEnv.Status)
	}

	var createChatResp ipc.CreateChatResponse
	if err := ipc.UnmarshalPayload(chatRespEnv.Payload, &createChatResp); err != nil {
		t.Fatalf("falha ao desserializar payload do chat: %v", err)
	}

	if createChatResp.Chat.Title != "Chat E2E Integração" || createChatResp.Chat.OwnerRegistrationID != "client_cpp_e2e" {
		t.Errorf("dados de chat E2E incorretos: %+v", createChatResp.Chat)
	}
}
