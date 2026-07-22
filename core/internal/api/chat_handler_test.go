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

func setupTestChatRPCEnvironment(t *testing.T) (*Router, *service.ChatService, *service.RegistrationService) {
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
	`
	if _, err := db.Exec(ddl); err != nil {
		t.Fatalf("falha ao executar DDL de teste: %v", err)
	}

	regRepo, _ := storage.NewRegistrationRepository(db)
	ruleRepo, _ := storage.NewSharedRuleRepository(db)
	chatRepo, _ := storage.NewChatRepository(db)
	evaluator := permission.NewEvaluator()

	regSvc, _ := service.NewRegistrationService(regRepo, ruleRepo, evaluator)
	chatSvc, _ := service.NewChatService(chatRepo, regRepo, ruleRepo, evaluator)

	chatHandler, _ := NewChatHandler(chatSvc)

	router := NewRouter()
	chatHandler.RegisterRoutes(router)

	return router, chatSvc, regSvc
}

func TestChatHandler_CreateChat_RPC(t *testing.T) {
	router, _, regSvc := setupTestChatRPCEnvironment(t)
	ctx := context.Background()

	_, _ = regSvc.RegisterClient(ctx, "client_cpp", "")

	reqPayload := ipc.CreateChatRequest{
		Title:    "Chat de Engenharia",
		Metadata: `{"topic":"architecture"}`,
	}
	reqEnv, _ := ipc.NewRequestEnvelope(ipc.MsgCreateChat, "client_cpp", reqPayload)
	rawBytes, _ := json.Marshal(reqEnv)

	respBytes := router.Dispatch(ctx, rawBytes)
	var respEnv ipc.ResponseEnvelope
	if err := json.Unmarshal(respBytes, &respEnv); err != nil {
		t.Fatalf("falha ao desserializar resposta JSON: %v", err)
	}

	if respEnv.Status != ipc.ErrSuccess {
		t.Fatalf("esperava Status 0, obteve %d", respEnv.Status)
	}

	var createResp ipc.CreateChatResponse
	if err := ipc.UnmarshalPayload(respEnv.Payload, &createResp); err != nil {
		t.Fatalf("falha ao desserializar payload de resposta: %v", err)
	}

	if createResp.Chat.Title != "Chat de Engenharia" || !createResp.Chat.IsOwner {
		t.Errorf("dados de chat incorretos: %+v", createResp.Chat)
	}
}

func TestChatHandler_GetChat_RPC(t *testing.T) {
	router, chatSvc, regSvc := setupTestChatRPCEnvironment(t)
	ctx := context.Background()

	_, _ = regSvc.RegisterClient(ctx, "client_cpp", "")
	created, _ := chatSvc.CreateChat(ctx, "client_cpp", "Meu Chat", "")

	reqPayload := ipc.GetChatRequest{
		ChatID: created.ID,
	}
	reqEnv, _ := ipc.NewRequestEnvelope(ipc.MsgGetChat, "client_cpp", reqPayload)
	rawBytes, _ := json.Marshal(reqEnv)

	respBytes := router.Dispatch(ctx, rawBytes)
	var respEnv ipc.ResponseEnvelope
	_ = json.Unmarshal(respBytes, &respEnv)

	if respEnv.Status != ipc.ErrSuccess {
		t.Fatalf("esperava Status 0, obteve %d", respEnv.Status)
	}

	var getResp ipc.GetChatResponse
	_ = ipc.UnmarshalPayload(respEnv.Payload, &getResp)
	if getResp.Chat.ID != created.ID || getResp.Chat.Title != "Meu Chat" {
		t.Errorf("dados de chat incorretos no GetChat: %+v", getResp.Chat)
	}
}

func TestChatHandler_RenameChat_RPC(t *testing.T) {
	router, chatSvc, regSvc := setupTestChatRPCEnvironment(t)
	ctx := context.Background()

	_, _ = regSvc.RegisterClient(ctx, "client_cpp", "")
	created, _ := chatSvc.CreateChat(ctx, "client_cpp", "Chat Antigo", "")

	reqPayload := ipc.RenameChatRequest{
		ChatID:   created.ID,
		NewTitle: "Chat Novo Título",
	}
	reqEnv, _ := ipc.NewRequestEnvelope(ipc.MsgRenameChat, "client_cpp", reqPayload)
	rawBytes, _ := json.Marshal(reqEnv)

	respBytes := router.Dispatch(ctx, rawBytes)
	var respEnv ipc.ResponseEnvelope
	_ = json.Unmarshal(respBytes, &respEnv)

	if respEnv.Status != ipc.ErrSuccess {
		t.Fatalf("esperava Status 0, obteve %d", respEnv.Status)
	}

	var renameResp ipc.RenameChatResponse
	_ = ipc.UnmarshalPayload(respEnv.Payload, &renameResp)
	if renameResp.Chat.Title != "Chat Novo Título" {
		t.Errorf("esperava novo título 'Chat Novo Título', obteve %s", renameResp.Chat.Title)
	}
}

func TestChatHandler_DeleteChat_RPC(t *testing.T) {
	router, chatSvc, regSvc := setupTestChatRPCEnvironment(t)
	ctx := context.Background()

	_, _ = regSvc.RegisterClient(ctx, "client_cpp", "")
	created, _ := chatSvc.CreateChat(ctx, "client_cpp", "Chat a Excluir", "")

	reqPayload := ipc.DeleteChatRequest{
		ChatID: created.ID,
	}
	reqEnv, _ := ipc.NewRequestEnvelope(ipc.MsgDeleteChat, "client_cpp", reqPayload)
	rawBytes, _ := json.Marshal(reqEnv)

	respBytes := router.Dispatch(ctx, rawBytes)
	var respEnv ipc.ResponseEnvelope
	_ = json.Unmarshal(respBytes, &respEnv)

	if respEnv.Status != ipc.ErrSuccess {
		t.Fatalf("esperava Status 0 na exclusão, obteve %d", respEnv.Status)
	}

	var delResp ipc.DeleteChatResponse
	_ = ipc.UnmarshalPayload(respEnv.Payload, &delResp)
	if !delResp.Success {
		t.Errorf("esperava success = true no DeleteChatResponse")
	}
}

func TestChatHandler_ListChats_RPC(t *testing.T) {
	router, chatSvc, regSvc := setupTestChatRPCEnvironment(t)
	ctx := context.Background()

	_, _ = regSvc.RegisterClient(ctx, "client_cpp", "")
	_, _ = chatSvc.CreateChat(ctx, "client_cpp", "Chat 1", "")
	_, _ = chatSvc.CreateChat(ctx, "client_cpp", "Chat 2", "")

	reqPayload := ipc.ListChatsRequest{
		IncludeShared: false,
		Limit:         10,
	}
	reqEnv, _ := ipc.NewRequestEnvelope(ipc.MsgListChats, "client_cpp", reqPayload)
	rawBytes, _ := json.Marshal(reqEnv)

	respBytes := router.Dispatch(ctx, rawBytes)
	var respEnv ipc.ResponseEnvelope
	_ = json.Unmarshal(respBytes, &respEnv)

	if respEnv.Status != ipc.ErrSuccess {
		t.Fatalf("esperava Status 0 na listagem, obteve %d", respEnv.Status)
	}

	var listResp ipc.ListChatsResponse
	_ = ipc.UnmarshalPayload(respEnv.Payload, &listResp)
	if listResp.Total != 2 || len(listResp.Chats) != 2 {
		t.Fatalf("esperava 2 chats na lista, obteve %d", listResp.Total)
	}
}
