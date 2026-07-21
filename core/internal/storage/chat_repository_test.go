package storage

import (
	"errors"
	"testing"
	"time"
)

func TestChatRepository_NilDatabase(t *testing.T) {
	_, err := NewChatRepository(nil)
	if !errors.Is(err, ErrNilDatabase) {
		t.Fatalf("esperava ErrNilDatabase, obteve: %v", err)
	}
}

func TestChatRepo_Create_Success(t *testing.T) {
	db := newTestMigratedDB(t)
	regRepo, _ := NewRegistrationRepository(db)
	chatRepo, _ := NewChatRepository(db)

	_ = regRepo.Create(Registration{ID: "reg-owner-1", Status: RegistrationActive})

	chat := Chat{
		ID:                  "chat-1",
		OwnerRegistrationID: "reg-owner-1",
		Title:               "Discussão de Arquitetura",
		Status:              ChatActive,
		MetadataJSON:        `{"topic":"sqlite"}`,
	}

	if err := chatRepo.Create(chat); err != nil {
		t.Fatalf("falha ao criar chat: %v", err)
	}

	fetched, err := chatRepo.FindByID("chat-1")
	if err != nil {
		t.Fatalf("falha no FindByID: %v", err)
	}

	if fetched.ID != chat.ID {
		t.Errorf("esperava ID %s, obteve %s", chat.ID, fetched.ID)
	}
	if fetched.OwnerRegistrationID != chat.OwnerRegistrationID {
		t.Errorf("esperava Owner %s, obteve %s", chat.OwnerRegistrationID, fetched.OwnerRegistrationID)
	}
	if fetched.Title != chat.Title {
		t.Errorf("esperava Title %s, obteve %s", chat.Title, fetched.Title)
	}
	if fetched.Status != ChatActive {
		t.Errorf("esperava Status ChatActive (1), obteve %d", fetched.Status)
	}
	if fetched.MetadataJSON != chat.MetadataJSON {
		t.Errorf("esperava MetadataJSON %s, obteve %s", chat.MetadataJSON, fetched.MetadataJSON)
	}
}

func TestChatRepo_Create_InvalidOwner(t *testing.T) {
	db := newTestMigratedDB(t)
	chatRepo, _ := NewChatRepository(db)

	chat := Chat{
		ID:                  "chat-invalid-owner",
		OwnerRegistrationID: "owner-que-nao-existe",
		Title:               "Chat Sem Owner",
	}

	err := chatRepo.Create(chat)
	if !errors.Is(err, ErrInvalidOwner) {
		t.Fatalf("esperava ErrInvalidOwner ao referenciar owner inexistente, obteve: %v", err)
	}
}

func TestChatRepo_Create_DuplicateID(t *testing.T) {
	db := newTestMigratedDB(t)
	regRepo, _ := NewRegistrationRepository(db)
	chatRepo, _ := NewChatRepository(db)

	_ = regRepo.Create(Registration{ID: "reg-owner-1", Status: RegistrationActive})

	chat := Chat{ID: "chat-dup", OwnerRegistrationID: "reg-owner-1", Title: "Chat 1"}
	if err := chatRepo.Create(chat); err != nil {
		t.Fatalf("falha na primeira criação: %v", err)
	}

	err := chatRepo.Create(chat)
	if !errors.Is(err, ErrAlreadyExists) {
		t.Fatalf("esperava ErrAlreadyExists na duplicação de ID, obteve: %v", err)
	}
}

func TestChatRepo_Create_StatusDeletedProhibited(t *testing.T) {
	db := newTestMigratedDB(t)
	regRepo, _ := NewRegistrationRepository(db)
	chatRepo, _ := NewChatRepository(db)

	_ = regRepo.Create(Registration{ID: "reg-owner-1", Status: RegistrationActive})

	chat := Chat{
		ID:                  "chat-deleted-init",
		OwnerRegistrationID: "reg-owner-1",
		Title:               "Chat Deletado Inicialmente",
		Status:              ChatDeleted,
	}

	err := chatRepo.Create(chat)
	if !errors.Is(err, ErrInvalidArgument) {
		t.Fatalf("esperava ErrInvalidArgument ao tentar criar chat com status ChatDeleted, obteve: %v", err)
	}
}

func TestChatRepo_Create_EmptyIDOrOwner(t *testing.T) {
	db := newTestMigratedDB(t)
	chatRepo, _ := NewChatRepository(db)

	err := chatRepo.Create(Chat{ID: "", OwnerRegistrationID: "owner-1", Title: "T"})
	if !errors.Is(err, ErrInvalidArgument) {
		t.Fatalf("esperava ErrInvalidArgument para ID vazio, obteve: %v", err)
	}

	err = chatRepo.Create(Chat{ID: "c-1", OwnerRegistrationID: "", Title: "T"})
	if !errors.Is(err, ErrInvalidArgument) {
		t.Fatalf("esperava ErrInvalidArgument para Owner vazio, obteve: %v", err)
	}
}

func TestChatRepo_FindByID_NotFound(t *testing.T) {
	db := newTestMigratedDB(t)
	chatRepo, _ := NewChatRepository(db)

	_, err := chatRepo.FindByID("chat-inexistente")
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("esperava ErrNotFound, obteve: %v", err)
	}
}

func TestChatRepo_FindByID_EmptyID(t *testing.T) {
	db := newTestMigratedDB(t)
	chatRepo, _ := NewChatRepository(db)

	_, err := chatRepo.FindByID("")
	if !errors.Is(err, ErrInvalidArgument) {
		t.Fatalf("esperava ErrInvalidArgument, obteve: %v", err)
	}
}

func TestChatRepo_Update_Success(t *testing.T) {
	db := newTestMigratedDB(t)
	regRepo, _ := NewRegistrationRepository(db)
	chatRepo, _ := NewChatRepository(db)

	_ = regRepo.Create(Registration{ID: "reg-owner-1", Status: RegistrationActive})

	chat := Chat{ID: "chat-up", OwnerRegistrationID: "reg-owner-1", Title: "Título Original", Status: ChatActive}
	if err := chatRepo.Create(chat); err != nil {
		t.Fatalf("falha ao criar chat: %v", err)
	}

	initial, _ := chatRepo.FindByID("chat-up")

	// Sleep > 1 segundo para virada de segundo do RFC3339
	time.Sleep(1005 * time.Millisecond)

	updated := Chat{
		ID:                  "chat-up",
		OwnerRegistrationID: "reg-owner-tentativa-mudanca", // Deve ser ignorado
		Title:               "Título Atualizado",
		Status:              ChatArchived,
		MetadataJSON:        `{"updated":true}`,
	}

	if err := chatRepo.Update(updated); err != nil {
		t.Fatalf("falha no Update: %v", err)
	}

	fetched, err := chatRepo.FindByID("chat-up")
	if err != nil {
		t.Fatalf("falha ao buscar chat atualizado: %v", err)
	}

	if fetched.Title != "Título Atualizado" {
		t.Errorf("esperava novo título, obteve: %s", fetched.Title)
	}
	if fetched.Status != ChatArchived {
		t.Errorf("esperava status ChatArchived (2), obteve: %d", fetched.Status)
	}
	if fetched.OwnerRegistrationID != "reg-owner-1" {
		t.Errorf("OwnerRegistrationID não deveria ter mudado. Esperava reg-owner-1, obteve: %s", fetched.OwnerRegistrationID)
	}
	if fetched.MetadataJSON != `{"updated":true}` {
		t.Errorf("esperava novo MetadataJSON, obteve: %s", fetched.MetadataJSON)
	}
	if fetched.UpdatedAt == initial.UpdatedAt {
		t.Errorf("updated_at deveria ter sido atualizado (inicial %s vs atual %s)", initial.UpdatedAt, fetched.UpdatedAt)
	}
}

func TestChatRepo_Update_StatusDeletedProhibited(t *testing.T) {
	db := newTestMigratedDB(t)
	regRepo, _ := NewRegistrationRepository(db)
	chatRepo, _ := NewChatRepository(db)

	_ = regRepo.Create(Registration{ID: "reg-owner-1", Status: RegistrationActive})
	_ = chatRepo.Create(Chat{ID: "chat-up-del", OwnerRegistrationID: "reg-owner-1", Title: "Original"})

	err := chatRepo.Update(Chat{ID: "chat-up-del", Status: ChatDeleted, Title: "Tentativa"})
	if !errors.Is(err, ErrInvalidArgument) {
		t.Fatalf("esperava ErrInvalidArgument ao tentar atribuir ChatDeleted no Update, obteve: %v", err)
	}
}

func TestChatRepo_Update_DeletedChat(t *testing.T) {
	db := newTestMigratedDB(t)
	regRepo, _ := NewRegistrationRepository(db)
	chatRepo, _ := NewChatRepository(db)

	_ = regRepo.Create(Registration{ID: "reg-owner-1", Status: RegistrationActive})
	_ = chatRepo.Create(Chat{ID: "chat-to-del", OwnerRegistrationID: "reg-owner-1", Title: "Original"})
	_ = chatRepo.Delete("chat-to-del")

	err := chatRepo.Update(Chat{ID: "chat-to-del", Status: ChatActive, Title: "Ressuscitando"})
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("esperava ErrNotFound ao tentar Update em chat soft-deleted, obteve: %v", err)
	}
}

func TestChatRepo_Delete_SoftDelete_And_Idempotency(t *testing.T) {
	db := newTestMigratedDB(t)
	regRepo, _ := NewRegistrationRepository(db)
	chatRepo, _ := NewChatRepository(db)

	_ = regRepo.Create(Registration{ID: "reg-owner-1", Status: RegistrationActive})
	_ = chatRepo.Create(Chat{ID: "chat-soft-del", OwnerRegistrationID: "reg-owner-1", Title: "Chat Exemplo"})

	// Primeira chamada de Delete
	if err := chatRepo.Delete("chat-soft-del"); err != nil {
		t.Fatalf("falha na primeira chamada do Delete: %v", err)
	}

	// FindByID deve retornar ErrNotFound
	_, err := chatRepo.FindByID("chat-soft-del")
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("esperava ErrNotFound no FindByID após Delete, obteve: %v", err)
	}

	// Segunda chamada de Delete deve ser idempotente (retornar nil)
	if err := chatRepo.Delete("chat-soft-del"); err != nil {
		t.Fatalf("esperava nil na segunda chamada de Delete (idempotente), obteve: %v", err)
	}

	// Confirma que a linha física no SQL ainda existe com status = 3
	var status int
	err = db.QueryRow(`SELECT status FROM chats WHERE id = 'chat-soft-del'`).Scan(&status)
	if err != nil {
		t.Fatalf("falha ao consultar registro físico no banco: %v", err)
	}
	if status != 3 {
		t.Fatalf("esperava status físico == 3 (ChatDeleted), obteve: %d", status)
	}
}

func TestChatRepo_Delete_NotFound(t *testing.T) {
	db := newTestMigratedDB(t)
	chatRepo, _ := NewChatRepository(db)

	err := chatRepo.Delete("chat-que-nunca-existiu")
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("esperava ErrNotFound ao deletar chat inexistente, obteve: %v", err)
	}
}

func TestChatRepo_ListByOwner_ActiveOnly_And_Ordering(t *testing.T) {
	db := newTestMigratedDB(t)
	regRepo, _ := NewRegistrationRepository(db)
	chatRepo, _ := NewChatRepository(db)

	_ = regRepo.Create(Registration{ID: "owner-list", Status: RegistrationActive})

	_ = chatRepo.Create(Chat{ID: "chat-a", OwnerRegistrationID: "owner-list", Title: "Chat A", Status: ChatActive})
	time.Sleep(1005 * time.Millisecond)
	_ = chatRepo.Create(Chat{ID: "chat-b", OwnerRegistrationID: "owner-list", Title: "Chat B", Status: ChatArchived})
	time.Sleep(1005 * time.Millisecond)
	_ = chatRepo.Create(Chat{ID: "chat-c", OwnerRegistrationID: "owner-list", Title: "Chat C", Status: ChatActive})
	_ = chatRepo.Create(Chat{ID: "chat-d", OwnerRegistrationID: "owner-list", Title: "Chat D", Status: ChatActive})
	_ = chatRepo.Delete("chat-d") // Deletado não deve aparecer em nenhuma lista

	// Teste 1: ActiveOnly
	actives, err := chatRepo.ListByOwner("owner-list", ChatFilterActiveOnly)
	if err != nil {
		t.Fatalf("falha ao listar chats ativos: %v", err)
	}
	if len(actives) != 2 {
		t.Fatalf("esperava 2 chats ativos, obteve %d", len(actives))
	}
	if actives[0].ID != "chat-c" || actives[1].ID != "chat-a" {
		t.Errorf("ordem incorreta em ActiveOnly: [%s, %s]", actives[0].ID, actives[1].ID)
	}

	// Teste 2: IncludeArchived
	all, err := chatRepo.ListByOwner("owner-list", ChatFilterIncludeArchived)
	if err != nil {
		t.Fatalf("falha ao listar chats incluindo arquivados: %v", err)
	}
	if len(all) != 3 {
		t.Fatalf("esperava 3 chats (2 ativos + 1 arquivado), obteve %d", len(all))
	}
	if all[0].ID != "chat-c" || all[1].ID != "chat-b" || all[2].ID != "chat-a" {
		t.Errorf("ordem incorreta em IncludeArchived: [%s, %s, %s]", all[0].ID, all[1].ID, all[2].ID)
	}

	// Teste 3: Atualizar chat-a com virada de segundo para que ele vá determinística e isoladamente para o topo
	time.Sleep(1005 * time.Millisecond)
	_ = chatRepo.Update(Chat{ID: "chat-a", Title: "Chat A Renomeado", Status: ChatActive})

	activesUpdated, _ := chatRepo.ListByOwner("owner-list", ChatFilterActiveOnly)
	if activesUpdated[0].ID != "chat-a" {
		t.Errorf("esperava chat-a no topo da lista após update (updated_at DESC), obteve %s", activesUpdated[0].ID)
	}
}
