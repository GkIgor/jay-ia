package api

import (
	"context"
	"database/sql"
	"encoding/json"
	"testing"

	_ "modernc.org/sqlite"

	"github.com/GkIgor/jay-ia/core/internal/llm"
	"github.com/GkIgor/jay-ia/core/internal/permission"
	"github.com/GkIgor/jay-ia/core/internal/service"
	"github.com/GkIgor/jay-ia/core/internal/storage"
	"github.com/GkIgor/jay-ia/core/internal/tools"
	"github.com/GkIgor/jay-ia/sdk/ipc"
)

type mockLLMClientHandler struct {
	response *llm.Response
}

func (m *mockLLMClientHandler) GenerateContent(ctx context.Context, history []llm.Message, availableTools []tools.Definition) (*llm.Response, error) {
	return m.response, nil
}

func setupTestProcessorRPCEnvironment(t *testing.T) (*Router, *service.ProcessorService, *service.MessageService, *service.ChatService, *service.RegistrationService) {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("falha ao abrir banco sqlite em memória: %v", err)
	}
	t.Cleanup(func() { db.Close() })

	ddl := `
	CREATE TABLE IF NOT EXISTS registrations (
		id TEXT PRIMARY KEY NOT NULL,
		metadata_json TEXT NOT NULL DEFAULT '{}',
		status INTEGER NOT NULL DEFAULT 1,
		created_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ', 'now')),
		updated_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ', 'now'))
	);
	CREATE TABLE IF NOT EXISTS shared_rules (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		registration_id TEXT NOT NULL,
		target_scope INTEGER NOT NULL DEFAULT 0,
		pattern TEXT NOT NULL,
		match_type INTEGER NOT NULL DEFAULT 1,
		allowed_actions INTEGER NOT NULL DEFAULT 1,
		created_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ', 'now')),
		FOREIGN KEY (registration_id) REFERENCES registrations(id) ON DELETE CASCADE
	);
	CREATE TABLE IF NOT EXISTS chats (
		id TEXT PRIMARY KEY NOT NULL,
		owner_registration_id TEXT NOT NULL,
		title TEXT NOT NULL,
		status INTEGER NOT NULL DEFAULT 1,
		metadata_json TEXT NOT NULL DEFAULT '{}',
		created_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ', 'now')),
		updated_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ', 'now')),
		FOREIGN KEY (owner_registration_id) REFERENCES registrations(id) ON DELETE RESTRICT
	);
	CREATE TABLE IF NOT EXISTS messages (
		id TEXT PRIMARY KEY NOT NULL,
		chat_id TEXT NOT NULL,
		author_type INTEGER NOT NULL DEFAULT 1,
		author_id TEXT NOT NULL,
		role INTEGER NOT NULL,
		content TEXT NOT NULL,
		content_type INTEGER NOT NULL DEFAULT 1,
		status INTEGER NOT NULL DEFAULT 1,
		sequence_no INTEGER NOT NULL,
		metadata_json TEXT NOT NULL DEFAULT '{}',
		created_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ', 'now')),
		updated_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ', 'now')),
		FOREIGN KEY (chat_id) REFERENCES chats(id) ON DELETE CASCADE
	);
	CREATE TABLE IF NOT EXISTS tools (
		id TEXT PRIMARY KEY NOT NULL,
		registration_id TEXT NOT NULL,
		name TEXT NOT NULL,
		description TEXT NOT NULL DEFAULT '',
		version TEXT NOT NULL DEFAULT '1.0.0',
		schema_json TEXT NOT NULL DEFAULT '{}',
		status INTEGER NOT NULL DEFAULT 1,
		created_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ', 'now')),
		updated_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ', 'now')),
		FOREIGN KEY (registration_id) REFERENCES registrations(id) ON DELETE CASCADE
	);
	`
	if _, err := db.Exec(ddl); err != nil {
		t.Fatalf("falha ao executar DDL de teste: %v", err)
	}

	regRepo, _ := storage.NewRegistrationRepository(db)
	ruleRepo, _ := storage.NewSharedRuleRepository(db)
	chatRepo, _ := storage.NewChatRepository(db)
	msgRepo, _ := storage.NewMessageRepository(db)
	toolRepo, _ := storage.NewToolRepository(db)
	evaluator := permission.NewEvaluator()

	mockLLM := &mockLLMClientHandler{
		response: &llm.Response{Text: "Resposta do Assistente Jay via RPC"},
	}

	regSvc, _ := service.NewRegistrationService(regRepo, ruleRepo, evaluator)
	chatSvc, _ := service.NewChatService(chatRepo, regRepo, ruleRepo, evaluator)
	msgSvc, _ := service.NewMessageService(msgRepo, chatRepo, ruleRepo, evaluator)
	procSvc, _ := service.NewProcessorService(msgRepo, chatRepo, toolRepo, ruleRepo, evaluator, mockLLM)

	procHandler, _ := NewProcessorHandler(procSvc)

	router := NewRouter()
	procHandler.RegisterRoutes(router)

	return router, procSvc, msgSvc, chatSvc, regSvc
}

func TestProcessorHandler_ProcessChat_RPC(t *testing.T) {
	router, _, msgSvc, chatSvc, regSvc := setupTestProcessorRPCEnvironment(t)
	ctx := context.Background()

	_, _ = regSvc.RegisterClient(ctx, "client_cpp", "")
	chat, _ := chatSvc.CreateChat(ctx, "client_cpp", "Chat 1", "")
	_, _ = msgSvc.CreateMessage(ctx, "client_cpp", ipc.CreateMessageRequest{ChatID: chat.ID, Content: "Mensagem do Usuário"})

	reqPayload := ipc.ProcessChatRequest{
		ChatID: chat.ID,
	}
	reqEnv, _ := ipc.NewRequestEnvelope(ipc.MsgProcessChat, "client_cpp", reqPayload)
	rawBytes, _ := json.Marshal(reqEnv)

	respBytes := router.Dispatch(ctx, rawBytes)
	var respEnv ipc.ResponseEnvelope
	if err := json.Unmarshal(respBytes, &respEnv); err != nil {
		t.Fatalf("falha ao desserializar resposta JSON: %v", err)
	}

	if respEnv.Status != ipc.ErrSuccess {
		t.Fatalf("esperava Status 0, obteve %d", respEnv.Status)
	}

	var procResp ipc.ProcessChatResponse
	if err := ipc.UnmarshalPayload(respEnv.Payload, &procResp); err != nil {
		t.Fatalf("falha ao desserializar payload de resposta: %v", err)
	}

	if procResp.ProcessedMessage.Content != "Resposta do Assistente Jay via RPC" || procResp.ProcessedMessage.Role != ipc.RoleAssistant {
		t.Errorf("dados da mensagem processada incorretos: %+v", procResp.ProcessedMessage)
	}
}
