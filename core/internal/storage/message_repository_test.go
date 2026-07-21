package storage

import (
	"errors"
	"fmt"
	"testing"
	"time"
)

func TestMessageRepository_NilDatabase(t *testing.T) {
	_, err := NewMessageRepository(nil)
	if !errors.Is(err, ErrNilDatabase) {
		t.Fatalf("esperava ErrNilDatabase, obteve: %v", err)
	}
}

func TestMessageRepo_Create_Success(t *testing.T) {
	db := newTestMigratedDB(t)
	regRepo, _ := NewRegistrationRepository(db)
	chatRepo, _ := NewChatRepository(db)
	msgRepo, _ := NewMessageRepository(db)

	_ = regRepo.Create(Registration{ID: "reg-1", Status: RegistrationActive})
	_ = chatRepo.Create(Chat{ID: "chat-1", OwnerRegistrationID: "reg-1", Title: "Chat 1"})

	msg := Message{
		ID:           "msg-1",
		ChatID:       "chat-1",
		AuthorType:   AuthorRegistration,
		AuthorID:     "reg-1",
		Role:         RoleUser,
		Content:      "Olá, mundo!",
		MetadataJSON: `{"client":"test"}`,
	}

	if err := msgRepo.Create(msg); err != nil {
		t.Fatalf("falha ao criar mensagem: %v", err)
	}

	fetched, err := msgRepo.FindByID("msg-1")
	if err != nil {
		t.Fatalf("falha no FindByID: %v", err)
	}

	if fetched.ID != msg.ID {
		t.Errorf("esperava ID %s, obteve %s", msg.ID, fetched.ID)
	}
	if fetched.ChatID != msg.ChatID {
		t.Errorf("esperava ChatID %s, obteve %s", msg.ChatID, fetched.ChatID)
	}
	if fetched.AuthorType != AuthorRegistration {
		t.Errorf("esperava AuthorType Registration (1), obteve %d", fetched.AuthorType)
	}
	if fetched.AuthorID != "reg-1" {
		t.Errorf("esperava AuthorID reg-1, obteve %s", fetched.AuthorID)
	}
	if fetched.Role != RoleUser {
		t.Errorf("esperava Role User (1), obteve %d", fetched.Role)
	}
	if fetched.Content != "Olá, mundo!" {
		t.Errorf("esperava Content 'Olá, mundo!', obteve %s", fetched.Content)
	}
	if fetched.ContentType != ContentTypeTextPlain {
		t.Errorf("esperava ContentType default TextPlain (1), obteve %d", fetched.ContentType)
	}
	if fetched.Status != MessageSent {
		t.Errorf("esperava Status default MessageSent (1), obteve %d", fetched.Status)
	}
	if fetched.SequenceNo != 1 {
		t.Errorf("esperava SequenceNo automático 1, obteve %d", fetched.SequenceNo)
	}
}

func TestMessageRepo_Create_AutoSequenceNo(t *testing.T) {
	db := newTestMigratedDB(t)
	regRepo, _ := NewRegistrationRepository(db)
	chatRepo, _ := NewChatRepository(db)
	msgRepo, _ := NewMessageRepository(db)

	_ = regRepo.Create(Registration{ID: "reg-1", Status: RegistrationActive})
	_ = chatRepo.Create(Chat{ID: "chat-seq", OwnerRegistrationID: "reg-1", Title: "Chat Seq"})

	m1 := Message{ID: "msg-a", ChatID: "chat-seq", AuthorType: AuthorRegistration, AuthorID: "reg-1", Role: RoleUser, Content: "M1"}
	m2 := Message{ID: "msg-b", ChatID: "chat-seq", AuthorType: AuthorAgent, AuthorID: "agent-1", Role: RoleAssistant, Content: "M2"}
	m3 := Message{ID: "msg-c", ChatID: "chat-seq", AuthorType: AuthorRegistration, AuthorID: "reg-1", Role: RoleUser, Content: "M3"}

	if err := msgRepo.Create(m1); err != nil {
		t.Fatalf("falha ao criar m1: %v", err)
	}
	if err := msgRepo.Create(m2); err != nil {
		t.Fatalf("falha ao criar m2: %v", err)
	}
	if err := msgRepo.Create(m3); err != nil {
		t.Fatalf("falha ao criar m3: %v", err)
	}

	f1, _ := msgRepo.FindByID("msg-a")
	f2, _ := msgRepo.FindByID("msg-b")
	f3, _ := msgRepo.FindByID("msg-c")

	if f1.SequenceNo != 1 || f2.SequenceNo != 2 || f3.SequenceNo != 3 {
		t.Fatalf("sequence_no automático incorreto. Obteve: m1=%d, m2=%d, m3=%d", f1.SequenceNo, f2.SequenceNo, f3.SequenceNo)
	}
}

func TestMessageRepo_Create_ExplicitSequenceNo(t *testing.T) {
	db := newTestMigratedDB(t)
	regRepo, _ := NewRegistrationRepository(db)
	chatRepo, _ := NewChatRepository(db)
	msgRepo, _ := NewMessageRepository(db)

	_ = regRepo.Create(Registration{ID: "reg-1", Status: RegistrationActive})
	_ = chatRepo.Create(Chat{ID: "chat-exp", OwnerRegistrationID: "reg-1", Title: "Chat Explicit"})

	msg := Message{
		ID:         "msg-exp",
		ChatID:     "chat-exp",
		AuthorType: AuthorRegistration,
		AuthorID:   "reg-1",
		Role:       RoleUser,
		Content:    "Importado",
		SequenceNo: 10,
	}

	if err := msgRepo.Create(msg); err != nil {
		t.Fatalf("falha ao criar mensagem com sequence_no explícito: %v", err)
	}

	fetched, _ := msgRepo.FindByID("msg-exp")
	if fetched.SequenceNo != 10 {
		t.Fatalf("esperava sequence_no 10, obteve %d", fetched.SequenceNo)
	}
}

func TestMessageRepo_Create_DuplicateSequenceNo(t *testing.T) {
	db := newTestMigratedDB(t)
	regRepo, _ := NewRegistrationRepository(db)
	chatRepo, _ := NewChatRepository(db)
	msgRepo, _ := NewMessageRepository(db)

	_ = regRepo.Create(Registration{ID: "reg-1", Status: RegistrationActive})
	_ = chatRepo.Create(Chat{ID: "chat-dup-seq", OwnerRegistrationID: "reg-1", Title: "Chat Dup Seq"})

	m1 := Message{ID: "m1", ChatID: "chat-dup-seq", AuthorType: AuthorRegistration, AuthorID: "reg-1", Role: RoleUser, Content: "C1", SequenceNo: 5}
	if err := msgRepo.Create(m1); err != nil {
		t.Fatalf("falha na primeira inserção: %v", err)
	}

	m2 := Message{ID: "m2", ChatID: "chat-dup-seq", AuthorType: AuthorRegistration, AuthorID: "reg-1", Role: RoleUser, Content: "C2", SequenceNo: 5}
	err := msgRepo.Create(m2)
	if !errors.Is(err, ErrAlreadyExists) {
		t.Fatalf("esperava ErrAlreadyExists ao duplicar sequence_no explícito, obteve: %v", err)
	}
}

func TestMessageRepo_Create_InvalidChat(t *testing.T) {
	db := newTestMigratedDB(t)
	msgRepo, _ := NewMessageRepository(db)

	msg := Message{
		ID:         "msg-invalid-chat",
		ChatID:     "chat-que-nao-existe",
		AuthorType: AuthorRegistration,
		AuthorID:   "reg-1",
		Role:       RoleUser,
		Content:    "Teste",
	}

	err := msgRepo.Create(msg)
	if !errors.Is(err, ErrInvalidChat) {
		t.Fatalf("esperava ErrInvalidChat ao referenciar chat inexistente, obteve: %v", err)
	}
}

func TestMessageRepo_Create_StatusDeletedProhibited(t *testing.T) {
	db := newTestMigratedDB(t)
	regRepo, _ := NewRegistrationRepository(db)
	chatRepo, _ := NewChatRepository(db)
	msgRepo, _ := NewMessageRepository(db)

	_ = regRepo.Create(Registration{ID: "reg-1", Status: RegistrationActive})
	_ = chatRepo.Create(Chat{ID: "chat-1", OwnerRegistrationID: "reg-1", Title: "Chat 1"})

	msg := Message{
		ID:         "msg-del-init",
		ChatID:     "chat-1",
		AuthorType: AuthorRegistration,
		AuthorID:   "reg-1",
		Role:       RoleUser,
		Content:    "Teste",
		Status:     MessageDeleted,
	}

	err := msgRepo.Create(msg)
	if !errors.Is(err, ErrInvalidArgument) {
		t.Fatalf("esperava ErrInvalidArgument ao criar mensagem com status MessageDeleted, obteve: %v", err)
	}
}

func TestMessageRepo_FindByID_NotFound(t *testing.T) {
	db := newTestMigratedDB(t)
	msgRepo, _ := NewMessageRepository(db)

	_, err := msgRepo.FindByID("msg-inexistente")
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("esperava ErrNotFound, obteve: %v", err)
	}
}

func TestMessageRepo_Update_Success(t *testing.T) {
	db := newTestMigratedDB(t)
	regRepo, _ := NewRegistrationRepository(db)
	chatRepo, _ := NewChatRepository(db)
	msgRepo, _ := NewMessageRepository(db)

	_ = regRepo.Create(Registration{ID: "reg-1", Status: RegistrationActive})
	_ = chatRepo.Create(Chat{ID: "chat-1", OwnerRegistrationID: "reg-1", Title: "Chat 1"})

	msg := Message{
		ID:          "msg-to-update",
		ChatID:      "chat-1",
		AuthorType:  AuthorRegistration,
		AuthorID:    "reg-1",
		Role:        RoleUser,
		Content:     "Conteúdo Original",
		ContentType: ContentTypeTextPlain,
		Status:      MessageSent,
	}
	_ = msgRepo.Create(msg)

	time.Sleep(1005 * time.Millisecond)

	updated := Message{
		ID:          "msg-to-update",
		ChatID:      "chat-tentativa-alterar", // Deve ser ignorado pelo UPDATE
		AuthorID:    "novo-author",            // Deve ser ignorado pelo UPDATE
		Content:     "Conteúdo Editado",
		ContentType: ContentTypeMarkdown,
	}

	if err := msgRepo.Update(updated); err != nil {
		t.Fatalf("falha no Update: %v", err)
	}

	fetched, err := msgRepo.FindByID("msg-to-update")
	if err != nil {
		t.Fatalf("falha ao buscar mensagem atualizada: %v", err)
	}

	if fetched.Content != "Conteúdo Editado" {
		t.Errorf("esperava novo conteúdo, obteve: %s", fetched.Content)
	}
	if fetched.ContentType != ContentTypeMarkdown {
		t.Errorf("esperava novo ContentType Markdown (2), obteve: %d", fetched.ContentType)
	}
	if fetched.Status != MessageEdited {
		t.Errorf("esperava status alterado para MessageEdited (2), obteve: %d", fetched.Status)
	}
	if fetched.ChatID != "chat-1" || fetched.AuthorID != "reg-1" {
		t.Errorf("campos imutáveis mudaram no Update: ChatID=%s, AuthorID=%s", fetched.ChatID, fetched.AuthorID)
	}
}

func TestMessageRepo_Update_DeletedMessage(t *testing.T) {
	db := newTestMigratedDB(t)
	regRepo, _ := NewRegistrationRepository(db)
	chatRepo, _ := NewChatRepository(db)
	msgRepo, _ := NewMessageRepository(db)

	_ = regRepo.Create(Registration{ID: "reg-1", Status: RegistrationActive})
	_ = chatRepo.Create(Chat{ID: "chat-1", OwnerRegistrationID: "reg-1", Title: "Chat 1"})

	_ = msgRepo.Create(Message{ID: "msg-to-del", ChatID: "chat-1", AuthorType: AuthorRegistration, AuthorID: "reg-1", Role: RoleUser, Content: "Original"})
	_ = msgRepo.Delete("msg-to-del")

	err := msgRepo.Update(Message{ID: "msg-to-del", Content: "Tentativa de Edição"})
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("esperava ErrNotFound ao tentar Update em mensagem soft-deleted, obteve: %v", err)
	}
}

func TestMessageRepo_Delete_SoftDelete_Idempotent(t *testing.T) {
	db := newTestMigratedDB(t)
	regRepo, _ := NewRegistrationRepository(db)
	chatRepo, _ := NewChatRepository(db)
	msgRepo, _ := NewMessageRepository(db)

	_ = regRepo.Create(Registration{ID: "reg-1", Status: RegistrationActive})
	_ = chatRepo.Create(Chat{ID: "chat-1", OwnerRegistrationID: "reg-1", Title: "Chat 1"})

	_ = msgRepo.Create(Message{ID: "msg-del-idem", ChatID: "chat-1", AuthorType: AuthorRegistration, AuthorID: "reg-1", Role: RoleUser, Content: "Apagar"})

	// Primeira exclusão
	if err := msgRepo.Delete("msg-del-idem"); err != nil {
		t.Fatalf("falha na primeira exclusão: %v", err)
	}

	_, err := msgRepo.FindByID("msg-del-idem")
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("esperava ErrNotFound no FindByID após Delete, obteve: %v", err)
	}

	// Segunda exclusão (idempotente)
	if err := msgRepo.Delete("msg-del-idem"); err != nil {
		t.Fatalf("esperava nil na segunda exclusão (idempotente), obteve: %v", err)
	}
}

func TestMessageRepo_ListByChat_SinceSequence(t *testing.T) {
	db := newTestMigratedDB(t)
	regRepo, _ := NewRegistrationRepository(db)
	chatRepo, _ := NewChatRepository(db)
	msgRepo, _ := NewMessageRepository(db)

	_ = regRepo.Create(Registration{ID: "reg-1", Status: RegistrationActive})
	_ = chatRepo.Create(Chat{ID: "chat-list", OwnerRegistrationID: "reg-1", Title: "Chat List"})

	for i := 1; i <= 5; i++ {
		_ = msgRepo.Create(Message{
			ID:         fmt.Sprintf("m-%d", i),
			ChatID:     "chat-list",
			AuthorType: AuthorRegistration,
			AuthorID:   "reg-1",
			Role:       RoleUser,
			Content:    fmt.Sprintf("Msg %d", i),
		})
	}

	// Consulta mensagens após sequence_no 2
	msgs, err := msgRepo.ListByChat("chat-list", 2, 100)
	if err != nil {
		t.Fatalf("falha no ListByChat: %v", err)
	}

	if len(msgs) != 3 {
		t.Fatalf("esperava 3 mensagens (sequence_no 3, 4, 5), obteve %d", len(msgs))
	}
	if msgs[0].SequenceNo != 3 || msgs[1].SequenceNo != 4 || msgs[2].SequenceNo != 5 {
		t.Fatalf("ordem de sequence_no incorreta: [%d, %d, %d]", msgs[0].SequenceNo, msgs[1].SequenceNo, msgs[2].SequenceNo)
	}
}

func TestMessageRepo_ListByChat_LimitCap(t *testing.T) {
	db := newTestMigratedDB(t)
	regRepo, _ := NewRegistrationRepository(db)
	chatRepo, _ := NewChatRepository(db)
	msgRepo, _ := NewMessageRepository(db)

	_ = regRepo.Create(Registration{ID: "reg-1", Status: RegistrationActive})
	_ = chatRepo.Create(Chat{ID: "chat-cap", OwnerRegistrationID: "reg-1", Title: "Chat Cap"})

	for i := 1; i <= 10; i++ {
		_ = msgRepo.Create(Message{
			ID:         fmt.Sprintf("m-cap-%d", i),
			ChatID:     "chat-cap",
			AuthorType: AuthorRegistration,
			AuthorID:   "reg-1",
			Role:       RoleUser,
			Content:    fmt.Sprintf("Msg %d", i),
		})
	}

	// Consulta com limit de 3
	msgs, err := msgRepo.ListByChat("chat-cap", 0, 3)
	if err != nil {
		t.Fatalf("falha no ListByChat com limit=3: %v", err)
	}

	if len(msgs) != 3 {
		t.Fatalf("esperava exatamente 3 mensagens devido ao limit, obteve %d", len(msgs))
	}
}
