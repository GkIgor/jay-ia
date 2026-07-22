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

func setupTestToolService(t *testing.T) (*ToolService, *RegistrationService) {
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
	toolRepo, _ := storage.NewToolRepository(db)
	evaluator := permission.NewEvaluator()

	regSvc, _ := NewRegistrationService(regRepo, ruleRepo, evaluator)
	toolSvc, err := NewToolService(toolRepo, ruleRepo, evaluator)
	if err != nil {
		t.Fatalf("falha ao instanciar ToolService: %v", err)
	}

	return toolSvc, regSvc
}

func TestToolService_RegisterTool_Success(t *testing.T) {
	toolSvc, regSvc := setupTestToolService(t)
	ctx := context.Background()

	_, _ = regSvc.RegisterClient(ctx, "owner_a", "")

	tool, err := toolSvc.RegisterTool(ctx, "owner_a", ipc.RegisterToolRequest{
		ID:          "calculator",
		Name:        "Calculadora",
		Description: "Realiza operações matemáticas",
		Version:     "1.2.0",
		SchemaJSON:  `{"type":"object"}`,
	})
	if err != nil {
		t.Fatalf("falha ao registrar ferramenta: %v", err)
	}

	if tool.ID != "calculator" || tool.RegistrationID != "owner_a" || tool.Status != storage.ToolAvailable {
		t.Errorf("dados da ferramenta incorretos: %+v", tool)
	}
}

func TestToolService_RegisterTool_HijackPrevention(t *testing.T) {
	toolSvc, regSvc := setupTestToolService(t)
	ctx := context.Background()

	_, _ = regSvc.RegisterClient(ctx, "owner_a", "")
	_, _ = regSvc.RegisterClient(ctx, "attacker_b", "")

	_, _ = toolSvc.RegisterTool(ctx, "owner_a", ipc.RegisterToolRequest{
		ID:   "web_search",
		Name: "Busca Web",
	})

	// Tentativa de sequestro por attacker_b -> ErrOwnershipConflict
	_, err := toolSvc.RegisterTool(ctx, "attacker_b", ipc.RegisterToolRequest{
		ID:   "web_search",
		Name: "Busca Web Maliciosa",
	})

	if err != storage.ErrOwnershipConflict {
		t.Fatalf("esperava ErrOwnershipConflict no sequestro de ferramenta, obteve: %v", err)
	}
}

func TestToolService_UnregisterTool(t *testing.T) {
	toolSvc, regSvc := setupTestToolService(t)
	ctx := context.Background()

	_, _ = regSvc.RegisterClient(ctx, "owner_a", "")
	_, _ = toolSvc.RegisterTool(ctx, "owner_a", ipc.RegisterToolRequest{ID: "tool_1", Name: "Ferramenta 1"})

	if err := toolSvc.UnregisterTool(ctx, "owner_a", "tool_1"); err != nil {
		t.Fatalf("falha ao descadastrar ferramenta: %v", err)
	}

	_, err := toolSvc.GetTool(ctx, "owner_a", "tool_1")
	if err != storage.ErrNotFound {
		t.Fatalf("esperava ErrNotFound para ferramenta removida, obteve: %v", err)
	}
}

func TestToolService_GetTool_SecurityHiding(t *testing.T) {
	toolSvc, regSvc := setupTestToolService(t)
	ctx := context.Background()

	_, _ = regSvc.RegisterClient(ctx, "owner_a", "")
	_, _ = regSvc.RegisterClient(ctx, "guest_b", "")
	_, _ = toolSvc.RegisterTool(ctx, "owner_a", ipc.RegisterToolRequest{ID: "private_tool", Name: "Privada"})

	// Leitura por guest não autorizado -> ErrNotFound
	_, err := toolSvc.GetTool(ctx, "guest_b", "private_tool")
	if err != storage.ErrNotFound {
		t.Fatalf("esperava ErrNotFound para ocultação de segurança, obteve: %v", err)
	}
}

func TestToolService_ListTools(t *testing.T) {
	toolSvc, regSvc := setupTestToolService(t)
	ctx := context.Background()

	_, _ = regSvc.RegisterClient(ctx, "owner_a", "")
	_, _ = toolSvc.RegisterTool(ctx, "owner_a", ipc.RegisterToolRequest{ID: "tool_a", Name: "Tool A"})
	_, _ = toolSvc.RegisterTool(ctx, "owner_a", ipc.RegisterToolRequest{ID: "tool_b", Name: "Tool B"})

	list, err := toolSvc.ListTools(ctx, "owner_a", "")
	if err != nil {
		t.Fatalf("falha ao listar ferramentas: %v", err)
	}

	if len(list) != 2 {
		t.Fatalf("esperava 2 ferramentas, obteve %d", len(list))
	}
}
