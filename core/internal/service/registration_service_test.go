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

func setupTestService(t *testing.T) (*RegistrationService, *storage.RegistrationRepository, *storage.SharedRuleRepository) {
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
	`
	if _, err := db.Exec(ddl); err != nil {
		t.Fatalf("falha ao criar tabelas de teste: %v", err)
	}

	regRepo, err := storage.NewRegistrationRepository(db)
	if err != nil {
		t.Fatalf("falha ao instanciar RegistrationRepository: %v", err)
	}

	ruleRepo, err := storage.NewSharedRuleRepository(db)
	if err != nil {
		t.Fatalf("falha ao instanciar SharedRuleRepository: %v", err)
	}

	evaluator := permission.NewEvaluator()

	svc, err := NewRegistrationService(regRepo, ruleRepo, evaluator)
	if err != nil {
		t.Fatalf("falha ao instanciar RegistrationService: %v", err)
	}

	return svc, regRepo, ruleRepo
}

func TestRegistrationService_RegisterClient(t *testing.T) {
	svc, _, _ := setupTestService(t)
	ctx := context.Background()

	reg, err := svc.RegisterClient(ctx, "client_a", `{"version":"1.0"}`)
	if err != nil {
		t.Fatalf("falha ao registrar cliente: %v", err)
	}
	if reg.ID != "client_a" || reg.Status != storage.RegistrationActive {
		t.Errorf("dados incorretos no registro: %+v", reg)
	}

	// Idempotência no re-registro
	regUpdated, err := svc.RegisterClient(ctx, "client_a", `{"version":"1.1"}`)
	if err != nil {
		t.Fatalf("falha no re-registro idempotente: %v", err)
	}
	if regUpdated.MetadataJSON != `{"version":"1.1"}` {
		t.Errorf("metadados não atualizados no re-registro: %s", regUpdated.MetadataJSON)
	}
}

func TestRegistrationService_UnregisterClient(t *testing.T) {
	svc, _, _ := setupTestService(t)
	ctx := context.Background()

	_, _ = svc.RegisterClient(ctx, "client_a", "")

	// Cliente A se desregistra
	if err := svc.UnregisterClient(ctx, "client_a", "client_a"); err != nil {
		t.Fatalf("falha ao desregistrar próprio cliente: %v", err)
	}

	// Verifica se foi removido
	_, err := svc.GetRegistration(ctx, "client_a", "client_a")
	if err == nil {
		t.Errorf("esperava erro ao buscar cliente removido")
	}
}

func TestRegistrationService_UnregisterClient_Forbidden(t *testing.T) {
	svc, _, _ := setupTestService(t)
	ctx := context.Background()

	_, _ = svc.RegisterClient(ctx, "client_a", "")
	_, _ = svc.RegisterClient(ctx, "client_b", "")

	// Cliente B tenta desregistrar Cliente A sem permissão de admin
	err := svc.UnregisterClient(ctx, "client_b", "client_a")
	if err == nil {
		t.Fatalf("esperava erro de autorização ao tentar desregistrar outro cliente")
	}
}

func TestRegistrationService_GetRegistration_SecurityHiding(t *testing.T) {
	svc, _, _ := setupTestService(t)
	ctx := context.Background()

	_, _ = svc.RegisterClient(ctx, "client_owner", "")
	_, _ = svc.RegisterClient(ctx, "client_other", "")

	// Consulta pelo próprio dono
	reg, err := svc.GetRegistration(ctx, "client_owner", "client_owner")
	if err != nil || reg.ID != "client_owner" {
		t.Fatalf("dono deveria conseguir consultar o próprio registro")
	}

	// Consulta por outro cliente sem permissão -> deve retornar ErrNotFound por segurança
	_, err = svc.GetRegistration(ctx, "client_other", "client_owner")
	if err == nil || err != storage.ErrNotFound {
		t.Fatalf("esperava ErrNotFound para ocultação de segurança, obteve: %v", err)
	}
}

func TestRegistrationService_UpdateSharedRules(t *testing.T) {
	svc, _, _ := setupTestService(t)
	ctx := context.Background()

	_, _ = svc.RegisterClient(ctx, "owner_1", "")

	rulesPayload := []ipc.SharedRulePayload{
		{TargetScope: 0, Pattern: "owner_", MatchType: 2, AllowedActions: 15}, // MatchPrefix (2) com prefixo "owner_"
	}

	count, err := svc.UpdateSharedRules(ctx, "owner_1", rulesPayload)
	if err != nil {
		t.Fatalf("falha ao atualizar regras de compartilhamento: %v", err)
	}
	if count != 1 {
		t.Errorf("esperava 1 regra aplicada, obteve %d", count)
	}

	// Consulta como outro cliente que busca o recurso owner_1 (cujo ID possui o prefixo owner_)
	_, _ = svc.RegisterClient(ctx, "client_2", "")
	reg, err := svc.GetRegistration(ctx, "client_2", "owner_1")
	if err != nil || reg.ID != "owner_1" {
		t.Fatalf("client_2 deveria conseguir ler owner_1 graças à regra compartilhada: %v", err)
	}
}
