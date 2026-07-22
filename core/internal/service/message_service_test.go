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

func setupTestMessageService(t *testing.T) (*MessageService, *ChatService, *RegistrationService) {
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

	regSvc, _ := NewRegistrationService(regRepo, ruleRepo, evaluator)
	chatSvc, _ := NewChatService(chatRepo, regRepo, ruleRepo, evaluator)
	msgSvc, err := NewMessageService(msgRepo, chatRepo, ruleRepo, evaluator)
	if err != nil {
		t.Fatalf("falha ao instanciar MessageService: %v", err)
	}

	return msgSvc, chatSvc, regSvc
}

func TestMessageService_CreateMessage_SequenceIncrement(t *testing.T) {
	msgSvc, chatSvc, regSvc := setupTestMessageService(t)
	ctx := context.Background()

	_, _ = regSvc.RegisterClient(ctx, "owner_a", "")
	chat, _ := chatSvc.CreateChat(ctx, "owner_a", "Chat de Teste", "")

	msg1, err := msgSvc.CreateMessage(ctx, "owner_a", ipc.CreateMessageRequest{
		ChatID:  chat.ID,
		Content: "Primeira mensagem",
	})
	if err != nil {
		t.Fatalf("falha ao criar mensagem 1: %v", err)
	}
	if msg1.SequenceNo != 1 {
		t.Errorf("esperava sequence_no = 1, obteve %d", msg1.SequenceNo)
	}

	msg2, err := msgSvc.CreateMessage(ctx, "owner_a", ipc.CreateMessageRequest{
		ChatID:  chat.ID,
		Content: "Segunda mensagem",
	})
	if err != nil {
		t.Fatalf("falha ao criar mensagem 2: %v", err)
	}
	if msg2.SequenceNo != 2 {
		t.Errorf("esperava sequence_no = 2, obteve %d", msg2.SequenceNo)
	}

	msg3, err := msgSvc.CreateMessage(ctx, "owner_a", ipc.CreateMessageRequest{
		ChatID:  chat.ID,
		Content: "Terceira mensagem",
	})
	if err != nil {
		t.Fatalf("falha ao criar mensagem 3: %v", err)
	}
	if msg3.SequenceNo != 3 {
		t.Errorf("esperava sequence_no = 3, obteve %d", msg3.SequenceNo)
	}
}

func TestMessageService_UpdateMessage(t *testing.T) {
	msgSvc, chatSvc, regSvc := setupTestMessageService(t)
	ctx := context.Background()

	_, _ = regSvc.RegisterClient(ctx, "owner_a", "")
	chat, _ := chatSvc.CreateChat(ctx, "owner_a", "Chat 1", "")
	msg, _ := msgSvc.CreateMessage(ctx, "owner_a", ipc.CreateMessageRequest{
		ChatID:  chat.ID,
		Content: "Conteúdo Original",
	})

	updated, err := msgSvc.UpdateMessage(ctx, "owner_a", msg.ID, "Conteúdo Editado", `{"edited":true}`)
	if err != nil {
		t.Fatalf("falha ao editar mensagem: %v", err)
	}

	if updated.Content != "Conteúdo Editado" || updated.Status != storage.MessageEdited {
		t.Errorf("dados incorretos na mensagem editada: %+v", updated)
	}
}

func TestMessageService_DeleteMessage_SoftDelete(t *testing.T) {
	msgSvc, chatSvc, regSvc := setupTestMessageService(t)
	ctx := context.Background()

	_, _ = regSvc.RegisterClient(ctx, "owner_a", "")
	chat, _ := chatSvc.CreateChat(ctx, "owner_a", "Chat 1", "")
	msg, _ := msgSvc.CreateMessage(ctx, "owner_a", ipc.CreateMessageRequest{
		ChatID:  chat.ID,
		Content: "Mensagem a Excluir",
	})

	if err := msgSvc.DeleteMessage(ctx, "owner_a", msg.ID); err != nil {
		t.Fatalf("falha no soft delete da mensagem: %v", err)
	}

	// Tentativa de edição em mensagem deletada deve falhar com ErrNotFound
	_, err := msgSvc.UpdateMessage(ctx, "owner_a", msg.ID, "Tentativa", "")
	if err != storage.ErrNotFound {
		t.Fatalf("esperava ErrNotFound ao tentar editar mensagem deletada, obteve: %v", err)
	}
}

func TestMessageService_GetMessages_PullModel(t *testing.T) {
	msgSvc, chatSvc, regSvc := setupTestMessageService(t)
	ctx := context.Background()

	_, _ = regSvc.RegisterClient(ctx, "owner_a", "")
	chat, _ := chatSvc.CreateChat(ctx, "owner_a", "Chat 1", "")

	for i := 1; i <= 5; i++ {
		_, _ = msgSvc.CreateMessage(ctx, "owner_a", ipc.CreateMessageRequest{
			ChatID:  chat.ID,
			Content: "Msg",
		})
	}

	// Consulta desde since_sequence_no = 2 com limit = 2 -> deve retornar msgs 3 e 4, com hasMore = true
	msgs, hasMore, err := msgSvc.GetMessages(ctx, "owner_a", chat.ID, 2, 2)
	if err != nil {
		t.Fatalf("falha no GetMessages: %v", err)
	}

	if len(msgs) != 2 {
		t.Fatalf("esperava 2 mensagens no resultado, obteve %d", len(msgs))
	}
	if msgs[0].SequenceNo != 3 || msgs[1].SequenceNo != 4 {
		t.Errorf("sequenciais incorretos: %d e %d", msgs[0].SequenceNo, msgs[1].SequenceNo)
	}
	if !hasMore {
		t.Errorf("esperava hasMore = true")
	}
}

func TestMessageService_GetMessages_SecurityHiding(t *testing.T) {
	msgSvc, chatSvc, regSvc := setupTestMessageService(t)
	ctx := context.Background()

	_, _ = regSvc.RegisterClient(ctx, "owner_a", "")
	_, _ = regSvc.RegisterClient(ctx, "guest_b", "")
	chat, _ := chatSvc.CreateChat(ctx, "owner_a", "Chat Privado", "")
	_, _ = msgSvc.CreateMessage(ctx, "owner_a", ipc.CreateMessageRequest{ChatID: chat.ID, Content: "Segredo"})

	// Guest sem acesso -> ErrNotFound
	_, _, err := msgSvc.GetMessages(ctx, "guest_b", chat.ID, 0, 10)
	if err != storage.ErrNotFound {
		t.Fatalf("esperava ErrNotFound para consulta não autorizada, obteve: %v", err)
	}
}
