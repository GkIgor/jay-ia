package storage

import (
	"errors"
	"testing"
	"time"
)

func TestToolRepository_NilDatabase(t *testing.T) {
	_, err := NewToolRepository(nil)
	if !errors.Is(err, ErrNilDatabase) {
		t.Fatalf("esperava ErrNilDatabase, obteve: %v", err)
	}
}

func TestToolRepo_Register_Create_Success(t *testing.T) {
	engine := newTestMigratedDB(t)
	regRepo, _ := NewRegistrationRepository(engine.DB())
	toolRepo, _ := NewToolRepository(engine.DB())

	_ = regRepo.Create(Registration{ID: "reg-provider-1", Status: RegistrationActive})

	tool := Tool{
		ID:             "web_search",
		RegistrationID: "reg-provider-1",
		Name:           "Pesquisa Web",
		Description:    "Realiza pesquisas online",
		SchemaJSON:     `{"type":"object"}`,
	}

	if err := toolRepo.Register(tool); err != nil {
		t.Fatalf("falha ao registrar ferramenta: %v", err)
	}

	fetched, err := toolRepo.FindByID("web_search")
	if err != nil {
		t.Fatalf("falha no FindByID: %v", err)
	}

	if fetched.ID != "web_search" || fetched.RegistrationID != "reg-provider-1" || fetched.Name != "Pesquisa Web" {
		t.Errorf("dados incorretos na ferramenta registrada: %+v", fetched)
	}
	if fetched.Version != "1.0.0" {
		t.Errorf("esperava fallback de versão '1.0.0', obteve %s", fetched.Version)
	}
	if fetched.Status != ToolAvailable {
		t.Errorf("esperava fallback de status ToolAvailable (1), obteve %d", fetched.Status)
	}
	if fetched.CreatedAt == "" || fetched.UpdatedAt == "" {
		t.Errorf("timestamps deveriam estar preenchidos")
	}
}

func TestToolRepo_Register_Upsert_Update(t *testing.T) {
	engine := newTestMigratedDB(t)
	regRepo, _ := NewRegistrationRepository(engine.DB())
	toolRepo, _ := NewToolRepository(engine.DB())

	_ = regRepo.Create(Registration{ID: "reg-provider-1", Status: RegistrationActive})

	t1 := Tool{
		ID:             "calculator",
		RegistrationID: "reg-provider-1",
		Name:           "Calculadora V1",
		Description:    "Calcula expressões",
		Version:        "1.0.0",
		Status:         ToolAvailable,
	}
	_ = toolRepo.Register(t1)

	initial, _ := toolRepo.FindByID("calculator")
	time.Sleep(1005 * time.Millisecond)

	t2 := Tool{
		ID:             "calculator",
		RegistrationID: "reg-provider-1", // Mesmo proprietário -> Deve atualizar
		Name:           "Calculadora V2",
		Description:    "Calcula expressões avançadas",
		Version:        "2.0.0",
		Status:         ToolAvailable,
	}

	if err := toolRepo.Register(t2); err != nil {
		t.Fatalf("falha ao atualizar ferramenta existente: %v", err)
	}

	fetched, err := toolRepo.FindByID("calculator")
	if err != nil {
		t.Fatalf("falha no FindByID após update: %v", err)
	}

	if fetched.Name != "Calculadora V2" || fetched.Version != "2.0.0" {
		t.Errorf("esperava dados atualizados, obteve name %s, version %s", fetched.Name, fetched.Version)
	}
	if fetched.CreatedAt != initial.CreatedAt {
		t.Errorf("created_at não deveria ter mudado. Inicial %s, Atual %s", initial.CreatedAt, fetched.CreatedAt)
	}
	if fetched.UpdatedAt == initial.UpdatedAt {
		t.Errorf("updated_at deveria ter sido atualizado")
	}
}

func TestToolRepo_Register_HijackPrevention(t *testing.T) {
	engine := newTestMigratedDB(t)
	regRepo, _ := NewRegistrationRepository(engine.DB())
	toolRepo, _ := NewToolRepository(engine.DB())

	_ = regRepo.Create(Registration{ID: "reg-dono-legitimo", Status: RegistrationActive})
	_ = regRepo.Create(Registration{ID: "reg-invasor", Status: RegistrationActive})

	tLegit := Tool{
		ID:             "shared_exec",
		RegistrationID: "reg-dono-legitimo",
		Name:           "Executor Legítimo",
	}
	_ = toolRepo.Register(tLegit)

	// Invasor tenta re-registrar a ferramenta "shared_exec" sob o registration ID dele
	tHijack := Tool{
		ID:             "shared_exec",
		RegistrationID: "reg-invasor",
		Name:           "Executor Invasor",
	}

	err := toolRepo.Register(tHijack)
	if !errors.Is(err, ErrOwnershipConflict) {
		t.Fatalf("esperava ErrOwnershipConflict ao tentar re-registrar ferramenta sob outro proprietário, obteve: %v", err)
	}

	// Asserção dupla: confirma via FindByID que a ferramenta NADA alterou e continua com o dono legítimo
	fetched, err := toolRepo.FindByID("shared_exec")
	if err != nil {
		t.Fatalf("falha ao buscar ferramenta no banco: %v", err)
	}
	if fetched.RegistrationID != "reg-dono-legitimo" {
		t.Fatalf("VIOLAÇÃO DE PROPRIEDADE! A ferramenta foi alterada para %s", fetched.RegistrationID)
	}
	if fetched.Name != "Executor Legítimo" {
		t.Fatalf("VIOLAÇÃO DE DADOS! O nome foi alterado para %s", fetched.Name)
	}
}

func TestToolRepo_Register_InvalidRegistration(t *testing.T) {
	engine := newTestMigratedDB(t)
	toolRepo, _ := NewToolRepository(engine.DB())

	tool := Tool{
		ID:             "ghost_tool",
		RegistrationID: "reg-fantasma-inexistente",
		Name:           "Ferramenta Fantasma",
	}

	err := toolRepo.Register(tool)
	if !errors.Is(err, ErrInvalidRegistration) {
		t.Fatalf("esperava ErrInvalidRegistration ao referenciar registration inexistente, obteve: %v", err)
	}
}

func TestToolRepo_FindByID_NotFound(t *testing.T) {
	engine := newTestMigratedDB(t)
	toolRepo, _ := NewToolRepository(engine.DB())

	_, err := toolRepo.FindByID("tool-inexistente")
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("esperava ErrNotFound, obteve: %v", err)
	}
}

func TestToolRepo_ListByRegistration(t *testing.T) {
	engine := newTestMigratedDB(t)
	regRepo, _ := NewRegistrationRepository(engine.DB())
	toolRepo, _ := NewToolRepository(engine.DB())

	_ = regRepo.Create(Registration{ID: "reg-dev-1", Status: RegistrationActive})

	_ = toolRepo.Register(Tool{ID: "t-z", RegistrationID: "reg-dev-1", Name: "Zebra Tool"})
	_ = toolRepo.Register(Tool{ID: "t-a", RegistrationID: "reg-dev-1", Name: "Alpha Tool"})

	tools, err := toolRepo.ListByRegistration("reg-dev-1")
	if err != nil {
		t.Fatalf("falha no ListByRegistration: %v", err)
	}

	if len(tools) != 2 {
		t.Fatalf("esperava 2 ferramentas, obteve %d", len(tools))
	}

	// Deve ser ordenado por name ASC: Alpha Tool primeiro, depois Zebra Tool
	if tools[0].Name != "Alpha Tool" || tools[1].Name != "Zebra Tool" {
		t.Errorf("ordenação por name ASC incorreta: [%s, %s]", tools[0].Name, tools[1].Name)
	}
}

func TestToolRepo_ListAvailable(t *testing.T) {
	engine := newTestMigratedDB(t)
	regRepo, _ := NewRegistrationRepository(engine.DB())
	toolRepo, _ := NewToolRepository(engine.DB())

	_ = regRepo.Create(Registration{ID: "reg-cat", Status: RegistrationActive})

	_ = toolRepo.Register(Tool{ID: "t-avail", RegistrationID: "reg-cat", Name: "Ferramenta Ativa", Status: ToolAvailable})
	_ = toolRepo.Register(Tool{ID: "t-dis", RegistrationID: "reg-cat", Name: "Ferramenta Desativada", Status: ToolDisabled})
	_ = toolRepo.Register(Tool{ID: "t-dep", RegistrationID: "reg-cat", Name: "Ferramenta Obsoleta", Status: ToolDeprecated})

	avail, err := toolRepo.ListAvailable()
	if err != nil {
		t.Fatalf("falha no ListAvailable: %v", err)
	}

	if len(avail) != 1 {
		t.Fatalf("esperava apenas 1 ferramenta disponível (status == 1), obteve %d", len(avail))
	}
	if avail[0].ID != "t-avail" {
		t.Errorf("esperava t-avail na lista, obteve %s", avail[0].ID)
	}
}

func TestToolRepo_UpdateStatus_Success(t *testing.T) {
	engine := newTestMigratedDB(t)
	regRepo, _ := NewRegistrationRepository(engine.DB())
	toolRepo, _ := NewToolRepository(engine.DB())

	_ = regRepo.Create(Registration{ID: "reg-1", Status: RegistrationActive})
	_ = toolRepo.Register(Tool{ID: "t-toggle", RegistrationID: "reg-1", Name: "Toggle Tool", Status: ToolAvailable})

	// Disponível -> Desativada
	if err := toolRepo.UpdateStatus("t-toggle", ToolDisabled); err != nil {
		t.Fatalf("falha ao atualizar status para ToolDisabled: %v", err)
	}

	fetched, _ := toolRepo.FindByID("t-toggle")
	if fetched.Status != ToolDisabled {
		t.Fatalf("esperava status ToolDisabled (2), obteve %d", fetched.Status)
	}

	// Desativada -> Reativada (Disponível)
	if err := toolRepo.UpdateStatus("t-toggle", ToolAvailable); err != nil {
		t.Fatalf("falha ao reativar para ToolAvailable: %v", err)
	}

	fetchedReactive, _ := toolRepo.FindByID("t-toggle")
	if fetchedReactive.Status != ToolAvailable {
		t.Fatalf("esperava status ToolAvailable (1), obteve %d", fetchedReactive.Status)
	}
}

func TestToolRepo_Delete_Success_And_Idempotent(t *testing.T) {
	engine := newTestMigratedDB(t)
	regRepo, _ := NewRegistrationRepository(engine.DB())
	toolRepo, _ := NewToolRepository(engine.DB())

	_ = regRepo.Create(Registration{ID: "reg-1", Status: RegistrationActive})
	_ = toolRepo.Register(Tool{ID: "t-del", RegistrationID: "reg-1", Name: "To Delete"})

	if err := toolRepo.Delete("t-del"); err != nil {
		t.Fatalf("falha no Delete: %v", err)
	}

	_, err := toolRepo.FindByID("t-del")
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("esperava ErrNotFound no FindByID após Delete, obteve: %v", err)
	}

	// Segunda chamada de Delete deve ser idempotente
	if err := toolRepo.Delete("t-del"); err != nil {
		t.Fatalf("esperava nil na segunda chamada de Delete (idempotente), obteve: %v", err)
	}

	// Delete de ID inexistente também deve ser idempotente
	if err := toolRepo.Delete("t-inexistente-que-nunca-existiu"); err != nil {
		t.Fatalf("esperava nil ao deletar ID inexistente (idempotente), obteve: %v", err)
	}
}

func TestToolRepo_CascadeOnRegistrationDelete(t *testing.T) {
	engine := newTestMigratedDB(t)
	regRepo, _ := NewRegistrationRepository(engine.DB())
	toolRepo, _ := NewToolRepository(engine.DB())

	_ = regRepo.Create(Registration{ID: "reg-owner-to-del", Status: RegistrationActive})
	_ = toolRepo.Register(Tool{ID: "t-child-1", RegistrationID: "reg-owner-to-del", Name: "Filha 1"})
	_ = toolRepo.Register(Tool{ID: "t-child-2", RegistrationID: "reg-owner-to-del", Name: "Filha 2"})

	// Exclui o Registration pai
	if err := regRepo.Delete("reg-owner-to-del"); err != nil {
		t.Fatalf("falha ao deletar registration pai: %v", err)
	}

	// As ferramentas filhas devem ser apagadas via SQL CASCADE
	var count int
	err := engine.DB().QueryRow(`SELECT COUNT(1) FROM tools WHERE registration_id = 'reg-owner-to-del';`).Scan(&count)
	if err != nil {
		t.Fatalf("falha ao verificar contagem física de ferramentas: %v", err)
	}
	if count != 0 {
		t.Fatalf("esperava 0 ferramentas filhas após CASCADE delete no registration pai, obteve %d", count)
	}
}
