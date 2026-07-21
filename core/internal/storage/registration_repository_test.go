package storage

import (
	"errors"
	"testing"
	"time"
)

func newTestMigratedDB(t *testing.T) *StorageEngine {
	t.Helper()
	engine, err := NewStorageEngine(Config{DatabasePath: ":memory:"})
	if err != nil {
		t.Fatalf("falha no NewStorageEngine: %v", err)
	}
	if err := engine.Open(); err != nil {
		t.Fatalf("falha no Open(): %v", err)
	}
	t.Cleanup(func() { _ = engine.Close() })

	migrator, err := NewMigrationEngine(engine.DB())
	if err != nil {
		t.Fatalf("falha no NewMigrationEngine: %v", err)
	}
	if err := migrator.Run(); err != nil {
		t.Fatalf("falha no Run(): %v", err)
	}

	return engine
}

func TestRegistrationRepository_NilDatabase(t *testing.T) {
	_, err := NewRegistrationRepository(nil)
	if !errors.Is(err, ErrNilDatabase) {
		t.Fatalf("esperava ErrNilDatabase, obteve: %v", err)
	}
}

func TestRegistrationRepo_Create_Success(t *testing.T) {
	engine := newTestMigratedDB(t)
	repo, err := NewRegistrationRepository(engine.DB())
	if err != nil {
		t.Fatalf("falha no NewRegistrationRepository: %v", err)
	}

	reg := Registration{
		ID:           "client-1",
		MetadataJSON: `{"version":"1.0"}`,
		Status:       RegistrationActive,
	}

	if err := repo.Create(reg); err != nil {
		t.Fatalf("falha ao executar Create: %v", err)
	}

	fetched, err := repo.FindByID("client-1")
	if err != nil {
		t.Fatalf("falha ao executar FindByID: %v", err)
	}

	if fetched.ID != reg.ID {
		t.Errorf("esperava ID %s, obteve %s", reg.ID, fetched.ID)
	}
	if fetched.MetadataJSON != reg.MetadataJSON {
		t.Errorf("esperava MetadataJSON %s, obteve %s", reg.MetadataJSON, fetched.MetadataJSON)
	}
	if fetched.Status != reg.Status {
		t.Errorf("esperava Status %d, obteve %d", reg.Status, fetched.Status)
	}
	if fetched.CreatedAt == "" || fetched.UpdatedAt == "" {
		t.Errorf("esperava timestamps preenchidos, obteve CreatedAt: %s, UpdatedAt: %s", fetched.CreatedAt, fetched.UpdatedAt)
	}
}

func TestRegistrationRepo_Create_DuplicateID(t *testing.T) {
	engine := newTestMigratedDB(t)
	repo, _ := NewRegistrationRepository(engine.DB())

	reg := Registration{ID: "client-1", Status: RegistrationActive}
	if err := repo.Create(reg); err != nil {
		t.Fatalf("falha na primeira inserção: %v", err)
	}

	err := repo.Create(reg)
	if !errors.Is(err, ErrAlreadyExists) {
		t.Fatalf("esperava ErrAlreadyExists, obteve: %v", err)
	}
}

func TestRegistrationRepo_Create_EmptyID(t *testing.T) {
	engine := newTestMigratedDB(t)
	repo, _ := NewRegistrationRepository(engine.DB())

	reg := Registration{ID: "", Status: RegistrationActive}
	err := repo.Create(reg)
	if !errors.Is(err, ErrInvalidArgument) {
		t.Fatalf("esperava ErrInvalidArgument, obteve: %v", err)
	}
}

func TestRegistrationRepo_Upsert_Insert(t *testing.T) {
	engine := newTestMigratedDB(t)
	repo, _ := NewRegistrationRepository(engine.DB())

	reg := Registration{
		ID:           "client-upsert-1",
		MetadataJSON: `{"env":"test"}`,
		Status:       RegistrationActive,
	}

	if err := repo.Upsert(reg); err != nil {
		t.Fatalf("falha no Upsert insert: %v", err)
	}

	fetched, err := repo.FindByID("client-upsert-1")
	if err != nil {
		t.Fatalf("falha ao buscar registro após Upsert: %v", err)
	}
	if fetched.MetadataJSON != reg.MetadataJSON {
		t.Errorf("esperava MetadataJSON %s, obteve %s", reg.MetadataJSON, fetched.MetadataJSON)
	}
}

func TestRegistrationRepo_Upsert_Update(t *testing.T) {
	engine := newTestMigratedDB(t)
	repo, _ := NewRegistrationRepository(engine.DB())

	reg := Registration{
		ID:           "client-upsert-2",
		MetadataJSON: `{"v":1}`,
		Status:       RegistrationActive,
	}
	if err := repo.Create(reg); err != nil {
		t.Fatalf("falha ao criar registro inicial: %v", err)
	}

	initial, err := repo.FindByID("client-upsert-2")
	if err != nil {
		t.Fatalf("falha ao buscar registro inicial: %v", err)
	}

	time.Sleep(10 * time.Millisecond)

	updatedReg := Registration{
		ID:           "client-upsert-2",
		MetadataJSON: `{"v":2}`,
		Status:       RegistrationInactive,
	}
	if err := repo.Upsert(updatedReg); err != nil {
		t.Fatalf("falha no Upsert update: %v", err)
	}

	fetched, err := repo.FindByID("client-upsert-2")
	if err != nil {
		t.Fatalf("falha ao buscar registro atualizado: %v", err)
	}

	if fetched.MetadataJSON != `{"v":2}` {
		t.Errorf("esperava MetadataJSON atualizado, obteve %s", fetched.MetadataJSON)
	}
	if fetched.Status != RegistrationInactive {
		t.Errorf("esperava Status alterado para Inactive, obteve %d", fetched.Status)
	}
	if fetched.CreatedAt != initial.CreatedAt {
		t.Errorf("CreatedAt não deveria ter mudado. Inicial: %s, Atual: %s", initial.CreatedAt, fetched.CreatedAt)
	}
}

func TestRegistrationRepo_FindByID_NotFound(t *testing.T) {
	engine := newTestMigratedDB(t)
	repo, _ := NewRegistrationRepository(engine.DB())

	_, err := repo.FindByID("inexistente")
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("esperava ErrNotFound, obteve: %v", err)
	}
}

func TestRegistrationRepo_FindByID_EmptyID(t *testing.T) {
	engine := newTestMigratedDB(t)
	repo, _ := NewRegistrationRepository(engine.DB())

	_, err := repo.FindByID("")
	if !errors.Is(err, ErrInvalidArgument) {
		t.Fatalf("esperava ErrInvalidArgument, obteve: %v", err)
	}
}

func TestRegistrationRepo_List_Empty(t *testing.T) {
	engine := newTestMigratedDB(t)
	repo, _ := NewRegistrationRepository(engine.DB())

	list, err := repo.List()
	if err != nil {
		t.Fatalf("falha ao listar banco vazio: %v", err)
	}
	if list == nil {
		t.Fatalf("esperava slice vazia não-nil, obteve nil")
	}
	if len(list) != 0 {
		t.Fatalf("esperava len 0, obteve %d", len(list))
	}
}

func TestRegistrationRepo_List_Multiple(t *testing.T) {
	engine := newTestMigratedDB(t)
	repo, _ := NewRegistrationRepository(engine.DB())

	reg1 := Registration{ID: "client-a", Status: RegistrationActive}
	reg2 := Registration{ID: "client-b", Status: RegistrationActive}
	reg3 := Registration{ID: "client-c", Status: RegistrationActive}

	_ = repo.Create(reg1)
	time.Sleep(5 * time.Millisecond)
	_ = repo.Create(reg2)
	time.Sleep(5 * time.Millisecond)
	_ = repo.Create(reg3)

	list, err := repo.List()
	if err != nil {
		t.Fatalf("falha no List: %v", err)
	}

	if len(list) != 3 {
		t.Fatalf("esperava len 3, obteve %d", len(list))
	}
	if list[0].ID != "client-a" || list[1].ID != "client-b" || list[2].ID != "client-c" {
		t.Fatalf("ordem de created_at incorreta: [%s, %s, %s]", list[0].ID, list[1].ID, list[2].ID)
	}
}

func TestRegistrationRepo_Delete_Success(t *testing.T) {
	engine := newTestMigratedDB(t)
	repo, _ := NewRegistrationRepository(engine.DB())

	reg := Registration{ID: "client-del", Status: RegistrationActive}
	_ = repo.Create(reg)

	if err := repo.Delete("client-del"); err != nil {
		t.Fatalf("falha no Delete: %v", err)
	}

	_, err := repo.FindByID("client-del")
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("esperava ErrNotFound após delete, obteve: %v", err)
	}
}

func TestRegistrationRepo_Delete_Idempotent(t *testing.T) {
	engine := newTestMigratedDB(t)
	repo, _ := NewRegistrationRepository(engine.DB())

	err := repo.Delete("inexistente")
	if err != nil {
		t.Fatalf("esperava nil no Delete de registro inexistente (idempotente), obteve: %v", err)
	}
}

func TestRegistrationRepo_Delete_EmptyID(t *testing.T) {
	engine := newTestMigratedDB(t)
	repo, _ := NewRegistrationRepository(engine.DB())

	err := repo.Delete("")
	if !errors.Is(err, ErrInvalidArgument) {
		t.Fatalf("esperava ErrInvalidArgument no Delete, obteve: %v", err)
	}
}

func TestRegistrationRepo_Delete_Restricted(t *testing.T) {
	engine := newTestMigratedDB(t)
	repo, _ := NewRegistrationRepository(engine.DB())

	reg := Registration{ID: "client-with-chat", Status: RegistrationActive}
	if err := repo.Create(reg); err != nil {
		t.Fatalf("falha ao criar registro: %v", err)
	}

	_, err := engine.DB().Exec(`INSERT INTO chats (id, owner_registration_id, title) VALUES ('chat-1', 'client-with-chat', 'Title')`)
	if err != nil {
		t.Fatalf("falha ao inserir chat de teste: %v", err)
	}

	err = repo.Delete("client-with-chat")
	if !errors.Is(err, ErrDeleteRestricted) {
		t.Fatalf("esperava ErrDeleteRestricted, obteve: %v", err)
	}
}

func TestRegistrationRepo_Create_After_Delete(t *testing.T) {
	engine := newTestMigratedDB(t)
	repo, _ := NewRegistrationRepository(engine.DB())

	reg := Registration{ID: "client-recycle", Status: RegistrationActive}
	if err := repo.Create(reg); err != nil {
		t.Fatalf("falha na primeira inserção: %v", err)
	}

	if err := repo.Delete("client-recycle"); err != nil {
		t.Fatalf("falha ao deletar registro: %v", err)
	}

	if err := repo.Create(reg); err != nil {
		t.Fatalf("falha ao recriar registro após delete: %v", err)
	}
}
