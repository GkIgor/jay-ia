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
	engine := newTestMigratedDB(t)
	regRepo, _ := NewRegistrationRepository(engine.DB())
	chatRepo, _ := NewChatRepository(engine.DB())
	msgRepo, _ := NewMessageRepository(engine.DB())

	_ = regRepo.Create(Registration{ID: "reg-user-1", Status: RegistrationActive})
	_ = chatRepo.Create(Chat{ID: "chat-m-1", OwnerRegistrationID: "reg-user-1", Title: "Chat 1"})

	msg := Message{
		ID:           "msg-1",
		ChatID:       "chat-m-1",
		AuthorType:   AuthorRegistration,
		AuthorID:     "reg-user-1",
		Role:         RoleUser,
		Content:      "Olá, Jay!",
		ContentType:  ContentTypeTextPlain,
		MetadataJSON: `{"client":"cli"}`,
	}

	if err := msgRepo.Create(msg); err != nil {
		t.Fatalf("falha ao criar mensagem: %v", err)
	}

	fetched, err := msgRepo.FindByID("msg-1")
	if err != nil {
		t.Fatalf("falha no FindByID: %v", err)
	}

	if fetched.ID != msg.ID || fetched.ChatID != msg.ChatID || fetched.Content != msg.Content {
		t.Errorf("dados incorretos na mensagem buscada: %+v", fetched)
	}
	if fetched.Status != MessageSent {
		t.Errorf("esperava status MessageSent (1), obteve: %d", fetched.Status)
	}
	if fetched.SequenceNo != 1 {
		t.Errorf("esperava sequence_no 1 auto-atribuído, obteve: %d", fetched.SequenceNo)
	}
}

func TestMessageRepo_Create_AutoSequenceNo(t *testing.T) {
	engine := newTestMigratedDB(t)
	regRepo, _ := NewRegistrationRepository(engine.DB())
	chatRepo, _ := NewChatRepository(engine.DB())
	msgRepo, _ := NewMessageRepository(engine.DB())

	_ = regRepo.Create(Registration{ID: "reg-user-1", Status: RegistrationActive})
	_ = chatRepo.Create(Chat{ID: "chat-m-auto", OwnerRegistrationID: "reg-user-1", Title: "Chat Auto"})

	m1 := Message{ID: "m-1", ChatID: "chat-m-auto", AuthorType: AuthorRegistration, AuthorID: "reg-user-1", Role: RoleUser, Content: "Msg 1"}
	m2 := Message{ID: "m-2", ChatID: "chat-m-auto", AuthorType: AuthorRegistration, AuthorID: "reg-user-1", Role: RoleUser, Content: "Msg 2"}
	m3 := Message{ID: "m-3", ChatID: "chat-m-auto", AuthorType: AuthorRegistration, AuthorID: "reg-user-1", Role: RoleUser, Content: "Msg 3"}

	_ = msgRepo.Create(m1)
	_ = msgRepo.Create(m2)
	_ = msgRepo.Create(m3)

	f1, _ := msgRepo.FindByID("m-1")
	f2, _ := msgRepo.FindByID("m-2")
	f3, _ := msgRepo.FindByID("m-3")

	if f1.SequenceNo != 1 || f2.SequenceNo != 2 || f3.SequenceNo != 3 {
		t.Fatalf("sequence_no auto-incrementais incorretos: [%d, %d, %d]", f1.SequenceNo, f2.SequenceNo, f3.SequenceNo)
	}
}

func TestMessageRepo_Create_ExplicitSequenceNo(t *testing.T) {
	engine := newTestMigratedDB(t)
	regRepo, _ := NewRegistrationRepository(engine.DB())
	chatRepo, _ := NewChatRepository(engine.DB())
	msgRepo, _ := NewMessageRepository(engine.DB())

	_ = regRepo.Create(Registration{ID: "reg-user-1", Status: RegistrationActive})
	_ = chatRepo.Create(Chat{ID: "chat-m-exp", OwnerRegistrationID: "reg-user-1", Title: "Chat Exp"})

	m := Message{ID: "m-exp-10", ChatID: "chat-m-exp", AuthorType: AuthorRegistration, AuthorID: "reg-user-1", Role: RoleUser, Content: "Msg Exp", SequenceNo: 10}

	if err := msgRepo.Create(m); err != nil {
		t.Fatalf("falha ao criar mensagem com sequence_no explícito: %v", err)
	}

	fetched, _ := msgRepo.FindByID("m-exp-10")
	if fetched.SequenceNo != 10 {
		t.Fatalf("esperava sequence_no explícito 10, obteve %d", fetched.SequenceNo)
	}
}

func TestMessageRepo_Create_DuplicateSequenceNo(t *testing.T) {
	engine := newTestMigratedDB(t)
	regRepo, _ := NewRegistrationRepository(engine.DB())
	chatRepo, _ := NewChatRepository(engine.DB())
	msgRepo, _ := NewMessageRepository(engine.DB())

	_ = regRepo.Create(Registration{ID: "reg-user-1", Status: RegistrationActive})
	_ = chatRepo.Create(Chat{ID: "chat-m-dup", OwnerRegistrationID: "reg-user-1", Title: "Chat Dup"})

	m1 := Message{ID: "m-dup-1", ChatID: "chat-m-dup", AuthorType: AuthorRegistration, AuthorID: "reg-user-1", Role: RoleUser, Content: "M1", SequenceNo: 5}
	_ = msgRepo.Create(m1)

	m2 := Message{ID: "m-dup-2", ChatID: "chat-m-dup", AuthorType: AuthorRegistration, AuthorID: "reg-user-1", Role: RoleUser, Content: "M2", SequenceNo: 5}
	err := msgRepo.Create(m2)
	if !errors.Is(err, ErrAlreadyExists) {
		t.Fatalf("esperava ErrAlreadyExists para sequence_no explícito duplicado, obteve: %v", err)
	}
}

func TestMessageRepo_Create_InvalidChat(t *testing.T) {
	engine := newTestMigratedDB(t)
	msgRepo, _ := NewMessageRepository(engine.DB())

	m := Message{ID: "m-inv", ChatID: "chat-inexistente", AuthorType: AuthorRegistration, AuthorID: "reg-1", Role: RoleUser, Content: "Test"}
	err := msgRepo.Create(m)
	if !errors.Is(err, ErrInvalidChat) {
		t.Fatalf("esperava ErrInvalidChat para chat inexistente, obteve: %v", err)
	}
}

func TestMessageRepo_Create_StatusDeletedProhibited(t *testing.T) {
	engine := newTestMigratedDB(t)
	regRepo, _ := NewRegistrationRepository(engine.DB())
	chatRepo, _ := NewChatRepository(engine.DB())
	msgRepo, _ := NewMessageRepository(engine.DB())

	_ = regRepo.Create(Registration{ID: "reg-user-1", Status: RegistrationActive})
	_ = chatRepo.Create(Chat{ID: "chat-m-1", OwnerRegistrationID: "reg-user-1", Title: "C1"})

	m := Message{ID: "m-del-init", ChatID: "chat-m-1", AuthorType: AuthorRegistration, AuthorID: "reg-user-1", Role: RoleUser, Content: "T", Status: MessageDeleted}
	err := msgRepo.Create(m)
	if !errors.Is(err, ErrInvalidArgument) {
		t.Fatalf("esperava ErrInvalidArgument ao criar mensagem com status MessageDeleted, obteve: %v", err)
	}
}

func TestMessageRepo_FindByID_NotFound(t *testing.T) {
	engine := newTestMigratedDB(t)
	msgRepo, _ := NewMessageRepository(engine.DB())

	_, err := msgRepo.FindByID("m-inexistente")
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("esperava ErrNotFound, obteve: %v", err)
	}
}

func TestMessageRepo_Update_Success(t *testing.T) {
	engine := newTestMigratedDB(t)
	regRepo, _ := NewRegistrationRepository(engine.DB())
	chatRepo, _ := NewChatRepository(engine.DB())
	msgRepo, _ := NewMessageRepository(engine.DB())

	_ = regRepo.Create(Registration{ID: "reg-user-1", Status: RegistrationActive})
	_ = chatRepo.Create(Chat{ID: "chat-m-1", OwnerRegistrationID: "reg-user-1", Title: "C1"})

	m := Message{ID: "m-up", ChatID: "chat-m-1", AuthorType: AuthorRegistration, AuthorID: "reg-user-1", Role: RoleUser, Content: "Conteúdo Original"}
	_ = msgRepo.Create(m)

	initial, _ := msgRepo.FindByID("m-up")

	time.Sleep(1005 * time.Millisecond)

	updated := Message{
		ID:           "m-up",
		ChatID:       "chat-modificado-deve-ser-ignorado",
		Content:      "Conteúdo Modificado",
		ContentType:  ContentTypeMarkdown,
		MetadataJSON: `{"edited":true}`,
	}

	if err := msgRepo.Update(updated); err != nil {
		t.Fatalf("falha no Update: %v", err)
	}

	fetched, _ := msgRepo.FindByID("m-up")
	if fetched.Content != "Conteúdo Modificado" {
		t.Errorf("esperava novo conteúdo, obteve: %s", fetched.Content)
	}
	if fetched.ContentType != ContentTypeMarkdown {
		t.Errorf("esperava ContentTypeMarkdown (2), obteve: %d", fetched.ContentType)
	}
	if fetched.Status != MessageEdited {
		t.Errorf("esperava status MessageEdited (2), obteve: %d", fetched.Status)
	}
	if fetched.ChatID != "chat-m-1" {
		t.Errorf("ChatID não deveria ter mudado, obteve: %s", fetched.ChatID)
	}
	if fetched.UpdatedAt == initial.UpdatedAt {
		t.Errorf("updated_at deveria ter mudado")
	}
}

func TestMessageRepo_Update_DeletedMessage(t *testing.T) {
	engine := newTestMigratedDB(t)
	regRepo, _ := NewRegistrationRepository(engine.DB())
	chatRepo, _ := NewChatRepository(engine.DB())
	msgRepo, _ := NewMessageRepository(engine.DB())

	_ = regRepo.Create(Registration{ID: "reg-user-1", Status: RegistrationActive})
	_ = chatRepo.Create(Chat{ID: "chat-m-1", OwnerRegistrationID: "reg-user-1", Title: "C1"})
	_ = msgRepo.Create(Message{ID: "m-to-del", ChatID: "chat-m-1", AuthorType: AuthorRegistration, AuthorID: "reg-user-1", Role: RoleUser, Content: "T"})
	_ = msgRepo.Delete("m-to-del")

	err := msgRepo.Update(Message{ID: "m-to-del", Content: "Editando"})
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("esperava ErrNotFound ao tentar Update em mensagem soft-deleted, obteve: %v", err)
	}
}

func TestMessageRepo_Delete_SoftDelete_Idempotent(t *testing.T) {
	engine := newTestMigratedDB(t)
	regRepo, _ := NewRegistrationRepository(engine.DB())
	chatRepo, _ := NewChatRepository(engine.DB())
	msgRepo, _ := NewMessageRepository(engine.DB())

	_ = regRepo.Create(Registration{ID: "reg-user-1", Status: RegistrationActive})
	_ = chatRepo.Create(Chat{ID: "chat-m-1", OwnerRegistrationID: "reg-user-1", Title: "C1"})
	_ = msgRepo.Create(Message{ID: "m-del", ChatID: "chat-m-1", AuthorType: AuthorRegistration, AuthorID: "reg-user-1", Role: RoleUser, Content: "T"})

	if err := msgRepo.Delete("m-del"); err != nil {
		t.Fatalf("falha no Delete: %v", err)
	}

	_, err := msgRepo.FindByID("m-del")
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("esperava ErrNotFound no FindByID após Delete, obteve: %v", err)
	}

	// Idempotência: segunda chamada retorna nil
	if err := msgRepo.Delete("m-del"); err != nil {
		t.Fatalf("esperava nil na segunda chamada de Delete (idempotente), obteve: %v", err)
	}

	// Deletar mensagem inexistente também é idempotente
	if err := msgRepo.Delete("m-inexistente-que-nunca-existiu"); err != nil {
		t.Fatalf("esperava nil no Delete de ID inexistente (idempotente), obteve: %v", err)
	}
}

func TestMessageRepo_ListByChat_SinceSequence(t *testing.T) {
	engine := newTestMigratedDB(t)
	regRepo, _ := NewRegistrationRepository(engine.DB())
	chatRepo, _ := NewChatRepository(engine.DB())
	msgRepo, _ := NewMessageRepository(engine.DB())

	_ = regRepo.Create(Registration{ID: "reg-user-1", Status: RegistrationActive})
	_ = chatRepo.Create(Chat{ID: "chat-m-list", OwnerRegistrationID: "reg-user-1", Title: "C List"})

	for i := 1; i <= 5; i++ {
		_ = msgRepo.Create(Message{ID: fmt.Sprintf("msg-seq-%d", i), ChatID: "chat-m-list", AuthorType: AuthorRegistration, AuthorID: "reg-user-1", Role: RoleUser, Content: "M"})
	}

	msgs, err := msgRepo.ListByChat("chat-m-list", 2, 10)
	if err != nil {
		t.Fatalf("falha no ListByChat: %v", err)
	}

	if len(msgs) != 3 {
		t.Fatalf("esperava 3 mensagens (seq 3, 4, 5), obteve %d", len(msgs))
	}
	if msgs[0].SequenceNo != 3 || msgs[1].SequenceNo != 4 || msgs[2].SequenceNo != 5 {
		t.Errorf("sequence_nos incorretos: [%d, %d, %d]", msgs[0].SequenceNo, msgs[1].SequenceNo, msgs[2].SequenceNo)
	}
}

func TestMessageRepo_ListByChat_LimitCap(t *testing.T) {
	engine := newTestMigratedDB(t)
	regRepo, _ := NewRegistrationRepository(engine.DB())
	chatRepo, _ := NewChatRepository(engine.DB())
	msgRepo, _ := NewMessageRepository(engine.DB())

	_ = regRepo.Create(Registration{ID: "reg-user-1", Status: RegistrationActive})
	_ = chatRepo.Create(Chat{ID: "chat-m-cap", OwnerRegistrationID: "reg-user-1", Title: "C Cap"})

	for i := 1; i <= 10; i++ {
		_ = msgRepo.Create(Message{ID: fmt.Sprintf("msg-cap-%d", i), ChatID: "chat-m-cap", AuthorType: AuthorRegistration, AuthorID: "reg-user-1", Role: RoleUser, Content: "M"})
	}

	// Solicitando limit 1000 deve ser limitado em 500 (retornando todas as 10 disponíveis)
	msgs, err := msgRepo.ListByChat("chat-m-cap", 0, 1000)
	if err != nil {
		t.Fatalf("falha no ListByChat com limit alto: %v", err)
	}
	if len(msgs) != 10 {
		t.Fatalf("esperava 10 mensagens, obteve %d", len(msgs))
	}
}
