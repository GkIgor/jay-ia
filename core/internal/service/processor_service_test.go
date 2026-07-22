package service

import (
	"context"
	"database/sql"
	"errors"
	"sync/atomic"
	"testing"
	"time"

	_ "modernc.org/sqlite"

	"github.com/GkIgor/jay-ia/core/internal/llm"
	"github.com/GkIgor/jay-ia/core/internal/permission"
	"github.com/GkIgor/jay-ia/core/internal/storage"
	"github.com/GkIgor/jay-ia/core/internal/tools"
	"github.com/GkIgor/jay-ia/sdk/ipc"
)

type mockLLMClient struct {
	response  *llm.Response
	err       error
	callCount int32
}

func (m *mockLLMClient) GenerateContent(ctx context.Context, history []llm.Message, availableTools []tools.Definition) (*llm.Response, error) {
	atomic.AddInt32(&m.callCount, 1)
	if m.err != nil {
		return nil, m.err
	}
	return m.response, nil
}

type errMockMessageStore struct {
	MessageStore
	createErr error
}

func (m *errMockMessageStore) Create(msg storage.Message) error {
	if m.createErr != nil {
		return m.createErr
	}
	return m.MessageStore.Create(msg)
}

func setupTestProcessorService(t *testing.T, llmClient llm.Client) (*ProcessorService, *MessageService, *ChatService, *RegistrationService, MessageStore, *storage.ChatRepository, *storage.SharedRuleRepository, *storage.ToolRepository, *permission.Evaluator) {
	db, err := sql.Open("sqlite", ":memory:?cache=shared")
	if err != nil {
		t.Fatalf("falha ao abrir banco sqlite em memória: %v", err)
	}
	db.SetMaxOpenConns(1)
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

	regSvc, _ := NewRegistrationService(regRepo, ruleRepo, evaluator)
	chatSvc, _ := NewChatService(chatRepo, regRepo, ruleRepo, evaluator)
	msgSvc, _ := NewMessageService(msgRepo, chatRepo, ruleRepo, evaluator)
	procSvc, err := NewProcessorService(msgRepo, chatRepo, toolRepo, ruleRepo, evaluator, llmClient)
	if err != nil {
		t.Fatalf("falha ao instanciar ProcessorService: %v", err)
	}

	return procSvc, msgSvc, chatSvc, regSvc, msgRepo, chatRepo, ruleRepo, toolRepo, evaluator
}

func TestProcessorService_ProcessChat_Success(t *testing.T) {
	mockClient := &mockLLMClient{
		response: &llm.Response{Text: "Olá! Sou o assistente Jay."},
	}
	procSvc, msgSvc, chatSvc, regSvc, _, _, _, _, _ := setupTestProcessorService(t, mockClient)
	ctx := context.Background()

	_, _ = regSvc.RegisterClient(ctx, "owner_a", "")
	chat, _ := chatSvc.CreateChat(ctx, "owner_a", "Chat 1", "")
	_, _ = msgSvc.CreateMessage(ctx, "owner_a", ipc.CreateMessageRequest{ChatID: chat.ID, Content: "Olá Jay!"})

	agentMsg, err := procSvc.ProcessChat(ctx, "owner_a", chat.ID)
	if err != nil {
		t.Fatalf("falha ao processar chat: %v", err)
	}

	if agentMsg.Content != "Olá! Sou o assistente Jay." {
		t.Errorf("conteúdo gerado pela LLM incorreto: %s", agentMsg.Content)
	}
	if agentMsg.AuthorType != storage.AuthorAgent || agentMsg.Role != storage.RoleAssistant {
		t.Errorf("autoria/papel incorretos: authorType=%d, role=%d", agentMsg.AuthorType, agentMsg.Role)
	}
	if agentMsg.SequenceNo != 2 {
		t.Errorf("esperava sequence_no = 2 para a mensagem do assistente, obteve %d", agentMsg.SequenceNo)
	}
}

func TestProcessorService_ProcessChat_PersistFailure(t *testing.T) {
	mockClient := &mockLLMClient{
		response: &llm.Response{Text: "Resposta gerada pela IA"},
	}
	_, _, chatSvc, regSvc, realMsgRepo, chatRepo, ruleRepo, toolRepo, evaluator := setupTestProcessorService(t, mockClient)
	ctx := context.Background()

	_, _ = regSvc.RegisterClient(ctx, "owner_a", "")
	chat, err := chatSvc.CreateChat(ctx, "owner_a", "Chat Teste", "")
	if err != nil || chat == nil {
		t.Fatalf("falha ao criar chat de teste: %v", err)
	}

	errStore := &errMockMessageStore{
		MessageStore: realMsgRepo,
		createErr:    errors.New("falha simulada de gravação no banco SQLite"),
	}

	procSvcWithErr, _ := NewProcessorService(errStore, chatRepo, toolRepo, ruleRepo, evaluator, mockClient)

	// Executa ProcessChat -> LLM responde com sucesso, mas a gravação no SQLite falha
	_, err = procSvcWithErr.ProcessChat(ctx, "owner_a", chat.ID)
	if err == nil {
		t.Fatalf("esperava erro de persistência, obteve nil")
	}

	// Valida que a LLM foi chamada exatamente 1 vez e NÃO foi feito retry
	if atomic.LoadInt32(&mockClient.callCount) != 1 {
		t.Errorf("esperava exatamente 1 chamada à LLM sem retry, obteve %d", mockClient.callCount)
	}
}

func TestProcessorService_ProcessChat_ConcurrencyLock(t *testing.T) {
	mockClient := &mockLLMClient{
		response: &llm.Response{Text: "Resposta concorrente"},
	}
	procSvc, msgSvc, chatSvc, regSvc, _, _, _, _, _ := setupTestProcessorService(t, mockClient)
	ctx := context.Background()

	_, _ = regSvc.RegisterClient(ctx, "owner_a", "")
	chat, _ := chatSvc.CreateChat(ctx, "owner_a", "Chat Concorrente", "")
	_, _ = msgSvc.CreateMessage(ctx, "owner_a", ipc.CreateMessageRequest{ChatID: chat.ID, Content: "Msg 1"})

	// Dispara 2 goroutines simultâneas de ProcessChat no mesmo chatID
	var err1, err2 error
	done := make(chan struct{})

	go func() {
		_, err1 = procSvc.ProcessChat(ctx, "owner_a", chat.ID)
		done <- struct{}{}
	}()
	go func() {
		_, err2 = procSvc.ProcessChat(ctx, "owner_a", chat.ID)
		done <- struct{}{}
	}()

	<-done
	<-done

	if err1 != nil || err2 != nil {
		t.Fatalf("falha ao processar requisições concorrentes: err1=%v, err2=%v", err1, err2)
	}
	if atomic.LoadInt32(&mockClient.callCount) != 2 {
		t.Errorf("esperava 2 chamadas à LLM serializadas, obteve %d", mockClient.callCount)
	}
}

func TestProcessorService_ProcessChat_ContextTimeout(t *testing.T) {
	slowMockClient := &mockLLMClient{
		err: context.DeadlineExceeded,
	}
	procSvc, _, chatSvc, regSvc, _, _, _, _, _ := setupTestProcessorService(t, slowMockClient)
	ctx := context.Background()

	_, _ = regSvc.RegisterClient(ctx, "owner_a", "")
	chat, _ := chatSvc.CreateChat(ctx, "owner_a", "Chat 1", "")

	cancCtx, cancel := context.WithTimeout(ctx, 1*time.Millisecond)
	defer cancel()

	time.Sleep(2 * time.Millisecond)

	_, err := procSvc.ProcessChat(cancCtx, "owner_a", chat.ID)
	if err == nil {
		t.Fatalf("esperava erro de contexto expirado, obteve nil")
	}
}
