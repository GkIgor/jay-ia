package api

import (
	"context"
	"database/sql"
	"encoding/json"
	"testing"

	_ "modernc.org/sqlite"

	"github.com/GkIgor/jay-ia/core/internal/permission"
	"github.com/GkIgor/jay-ia/core/internal/service"
	"github.com/GkIgor/jay-ia/core/internal/storage"
	"github.com/GkIgor/jay-ia/sdk/ipc"
)

func setupTestMessageRPCEnvironment(t *testing.T) (*Router, *service.MessageService, *service.ChatService, *service.RegistrationService) {
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
	`
	if _, err := db.Exec(ddl); err != nil {
		t.Fatalf("falha ao executar DDL de teste: %v", err)
	}

	regRepo, _ := storage.NewRegistrationRepository(db)
	ruleRepo, _ := storage.NewSharedRuleRepository(db)
	chatRepo, _ := storage.NewChatRepository(db)
	msgRepo, _ := storage.NewMessageRepository(db)
	evaluator := permission.NewEvaluator()

	regSvc, _ := service.NewRegistrationService(regRepo, ruleRepo, evaluator)
	chatSvc, _ := service.NewChatService(chatRepo, regRepo, ruleRepo, evaluator)
	msgSvc, _ := service.NewMessageService(msgRepo, chatRepo, ruleRepo, evaluator)

	msgHandler, _ := NewMessageHandler(msgSvc)

	router := NewRouter()
	msgHandler.RegisterRoutes(router)

	return router, msgSvc, chatSvc, regSvc
}

func TestMessageHandler_CreateMessage_RPC(t *testing.T) {
	router, _, chatSvc, regSvc := setupTestMessageRPCEnvironment(t)
	ctx := context.Background()

	_, _ = regSvc.RegisterClient(ctx, "client_cpp", "")
	chat, _ := chatSvc.CreateChat(ctx, "client_cpp", "Chat 1", "")

	reqPayload := ipc.CreateMessageRequest{
		ChatID:  chat.ID,
		Content: "Olá Core",
	}
	reqEnv, _ := ipc.NewRequestEnvelope(ipc.MsgCreateMessage, "client_cpp", reqPayload)
	rawBytes, _ := json.Marshal(reqEnv)

	respBytes := router.Dispatch(ctx, rawBytes)
	var respEnv ipc.ResponseEnvelope
	if err := json.Unmarshal(respBytes, &respEnv); err != nil {
		t.Fatalf("falha ao desserializar resposta JSON: %v", err)
	}

	if respEnv.Status != ipc.ErrSuccess {
		t.Fatalf("esperava Status 0, obteve %d", respEnv.Status)
	}

	var createResp ipc.CreateMessageResponse
	if err := ipc.UnmarshalPayload(respEnv.Payload, &createResp); err != nil {
		t.Fatalf("falha ao desserializar payload de resposta: %v", err)
	}

	if createResp.CreatedMessage.Content != "Olá Core" || createResp.CreatedMessage.SequenceNo != 1 {
		t.Errorf("dados de mensagem incorretos: %+v", createResp.CreatedMessage)
	}
}

func TestMessageHandler_UpdateMessage_RPC(t *testing.T) {
	router, msgSvc, chatSvc, regSvc := setupTestMessageRPCEnvironment(t)
	ctx := context.Background()

	_, _ = regSvc.RegisterClient(ctx, "client_cpp", "")
	chat, _ := chatSvc.CreateChat(ctx, "client_cpp", "Chat 1", "")
	created, _ := msgSvc.CreateMessage(ctx, "client_cpp", ipc.CreateMessageRequest{ChatID: chat.ID, Content: "Texto Antigo"})

	reqPayload := ipc.UpdateMessageRequest{
		MessageID:  created.ID,
		NewContent: "Texto Atualizado",
	}
	reqEnv, _ := ipc.NewRequestEnvelope(ipc.MsgUpdateMessage, "client_cpp", reqPayload)
	rawBytes, _ := json.Marshal(reqEnv)

	respBytes := router.Dispatch(ctx, rawBytes)
	var respEnv ipc.ResponseEnvelope
	_ = json.Unmarshal(respBytes, &respEnv)

	if respEnv.Status != ipc.ErrSuccess {
		t.Fatalf("esperava Status 0 na edição, obteve %d", respEnv.Status)
	}

	var updateResp ipc.UpdateMessageResponse
	_ = ipc.UnmarshalPayload(respEnv.Payload, &updateResp)
	if updateResp.Message.Content != "Texto Atualizado" {
		t.Errorf("esperava novo conteúdo 'Texto Atualizado', obteve %s", updateResp.Message.Content)
	}
}

func TestMessageHandler_DeleteMessage_RPC(t *testing.T) {
	router, msgSvc, chatSvc, regSvc := setupTestMessageRPCEnvironment(t)
	ctx := context.Background()

	_, _ = regSvc.RegisterClient(ctx, "client_cpp", "")
	chat, _ := chatSvc.CreateChat(ctx, "client_cpp", "Chat 1", "")
	created, _ := msgSvc.CreateMessage(ctx, "client_cpp", ipc.CreateMessageRequest{ChatID: chat.ID, Content: "A Deletar"})

	reqPayload := ipc.DeleteMessageRequest{
		MessageID: created.ID,
	}
	reqEnv, _ := ipc.NewRequestEnvelope(ipc.MsgDeleteMessage, "client_cpp", reqPayload)
	rawBytes, _ := json.Marshal(reqEnv)

	respBytes := router.Dispatch(ctx, rawBytes)
	var respEnv ipc.ResponseEnvelope
	_ = json.Unmarshal(respBytes, &respEnv)

	if respEnv.Status != ipc.ErrSuccess {
		t.Fatalf("esperava Status 0 na exclusão, obteve %d", respEnv.Status)
	}

	var delResp ipc.DeleteMessageResponse
	_ = ipc.UnmarshalPayload(respEnv.Payload, &delResp)
	if !delResp.Success {
		t.Errorf("esperava success = true")
	}
}

func TestMessageHandler_GetMessages_RPC(t *testing.T) {
	router, msgSvc, chatSvc, regSvc := setupTestMessageRPCEnvironment(t)
	ctx := context.Background()

	_, _ = regSvc.RegisterClient(ctx, "client_cpp", "")
	chat, _ := chatSvc.CreateChat(ctx, "client_cpp", "Chat 1", "")
	_, _ = msgSvc.CreateMessage(ctx, "client_cpp", ipc.CreateMessageRequest{ChatID: chat.ID, Content: "M1"})
	_, _ = msgSvc.CreateMessage(ctx, "client_cpp", ipc.CreateMessageRequest{ChatID: chat.ID, Content: "M2"})

	reqPayload := ipc.GetMessagesRequest{
		ChatID:          chat.ID,
		SinceSequenceNo: 0,
		Limit:           10,
	}
	reqEnv, _ := ipc.NewRequestEnvelope(ipc.MsgGetMessages, "client_cpp", reqPayload)
	rawBytes, _ := json.Marshal(reqEnv)

	respBytes := router.Dispatch(ctx, rawBytes)
	var respEnv ipc.ResponseEnvelope
	_ = json.Unmarshal(respBytes, &respEnv)

	if respEnv.Status != ipc.ErrSuccess {
		t.Fatalf("esperava Status 0 na listagem, obteve %d", respEnv.Status)
	}

	var getResp ipc.GetMessagesResponse
	_ = ipc.UnmarshalPayload(respEnv.Payload, &getResp)
	if len(getResp.Messages) != 2 || getResp.ChatID != chat.ID {
		t.Fatalf("dados incorretos no GetMessagesResponse: %+v", getResp)
	}
}
