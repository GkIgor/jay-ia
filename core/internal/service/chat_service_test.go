package service

import (
	"context"
	"database/sql"
	"testing"

	_ "modernc.org/sqlite"

	"github.com/GkIgor/jay-ia/core/internal/permission"
	"github.com/GkIgor/jay-ia/core/internal/storage"
	"github.com/GkIgor/jay-ia/sdk/ipc"
)

func setupTestChatService(t *testing.T) (*ChatService, *RegistrationService) {
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

	regSvc, _ := NewRegistrationService(regRepo, ruleRepo, evaluator)
	chatSvc, err := NewChatService(chatRepo, regRepo, ruleRepo, evaluator)
	if err != nil {
		t.Fatalf("falha ao instanciar ChatService: %v", err)
	}

	return chatSvc, regSvc
}

func TestChatService_CreateChat(t *testing.T) {
	chatSvc, regSvc := setupTestChatService(t)
	ctx := context.Background()

	_, _ = regSvc.RegisterClient(ctx, "owner_a", "")

	chat, err := chatSvc.CreateChat(ctx, "owner_a", "Chat de Teste", `{"category":"general"}`)
	if err != nil {
		t.Fatalf("falha ao criar chat: %v", err)
	}

	if chat.ID == "" {
		t.Errorf("UUID do chat não pode ser vazio")
	}
	if chat.OwnerRegistrationID != "owner_a" || chat.Title != "Chat de Teste" || chat.Status != storage.ChatActive {
		t.Errorf("dados do chat incorretos: %+v", chat)
	}
}

func TestChatService_DeleteChat_SoftDelete(t *testing.T) {
	chatSvc, regSvc := setupTestChatService(t)
	ctx := context.Background()

	_, _ = regSvc.RegisterClient(ctx, "owner_a", "")
	chat, _ := chatSvc.CreateChat(ctx, "owner_a", "Chat a Deletar", "")

	// Exclui o chat
	if err := chatSvc.DeleteChat(ctx, "owner_a", chat.ID); err != nil {
		t.Fatalf("falha ao deletar chat: %v", err)
	}

	// Consulta subsequente deve retornar ErrNotFound
	_, err := chatSvc.GetChat(ctx, "owner_a", chat.ID)
	if err != storage.ErrNotFound {
		t.Fatalf("esperava ErrNotFound para chat deletado, obteve: %v", err)
	}

	// Exclusão repetida deve retornar ErrNotFound
	errRepeat := chatSvc.DeleteChat(ctx, "owner_a", chat.ID)
	if errRepeat != storage.ErrNotFound {
		t.Fatalf("esperava ErrNotFound em exclusão repetida, obteve: %v", errRepeat)
	}
}

func TestChatService_RenameChat(t *testing.T) {
	chatSvc, regSvc := setupTestChatService(t)
	ctx := context.Background()

	_, _ = regSvc.RegisterClient(ctx, "owner_a", "")
	chat, _ := chatSvc.CreateChat(ctx, "owner_a", "Título Antigo", "")

	renamed, err := chatSvc.RenameChat(ctx, "owner_a", chat.ID, "Título Novo")
	if err != nil {
		t.Fatalf("falha ao renomear chat: %v", err)
	}
	if renamed.Title != "Título Novo" {
		t.Errorf("esperava 'Título Novo', obteve %s", renamed.Title)
	}
}

func TestChatService_GetChat_SecurityHiding(t *testing.T) {
	chatSvc, regSvc := setupTestChatService(t)
	ctx := context.Background()

	_, _ = regSvc.RegisterClient(ctx, "owner_a", "")
	_, _ = regSvc.RegisterClient(ctx, "guest_b", "")
	chat, _ := chatSvc.CreateChat(ctx, "owner_a", "Chat Privado", "")

	// Leitura pelo proprietário -> Sucesso
	c, err := chatSvc.GetChat(ctx, "owner_a", chat.ID)
	if err != nil || c.ID != chat.ID {
		t.Fatalf("proprietário deveria conseguir ler o chat")
	}

	// Leitura por guest não autorizado -> ErrNotFound (Ocultação de Segurança)
	_, err = chatSvc.GetChat(ctx, "guest_b", chat.ID)
	if err != storage.ErrNotFound {
		t.Fatalf("esperava ErrNotFound para ocultação de segurança, obteve: %v", err)
	}
}

func TestChatService_ListChats_IncludeShared(t *testing.T) {
	chatSvc, regSvc := setupTestChatService(t)
	ctx := context.Background()

	_, _ = regSvc.RegisterClient(ctx, "owner_1", "")
	_, _ = regSvc.RegisterClient(ctx, "client_2", "")

	chat1, _ := chatSvc.CreateChat(ctx, "owner_1", "Chat do Owner 1", "")
	chat2, _ := chatSvc.CreateChat(ctx, "client_2", "Chat do Client 2", "")

	// Owner 1 compartilha seus chats com client_2
	rulesPayload := []ipc.SharedRulePayload{
		{TargetScope: 1, Pattern: chat1.ID, MatchType: 1, AllowedActions: 15}, // ScopeChats (1), MatchExact
	}
	_, _ = regSvc.UpdateSharedRules(ctx, "owner_1", rulesPayload)

	// client_2 lista com includeShared = true
	list, err := chatSvc.ListChats(ctx, "client_2", true, 50)
	if err != nil {
		t.Fatalf("falha ao listar chats: %v", err)
	}

	if len(list) != 2 {
		t.Fatalf("esperava 2 chats (1 próprio + 1 compartilhado), obteve %d", len(list))
	}

	foundShared := false
	for _, c := range list {
		if c.ID == chat1.ID {
			foundShared = true
		}
	}
	if !foundShared {
		t.Errorf("chat compartilhado não foi encontrado no resultado da listagem")
	}
	_ = chat2
}
