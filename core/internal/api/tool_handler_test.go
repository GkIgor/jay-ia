package api

import (
	"context"
	"database/sql"
	"encoding/json"
	"testing"

	_ "modernc.org/sqlite"

	"github.com/GkIgor/jay-ia/core/internal/permission"
	"github.com/GkIgor/jay-ia/core/internal/service"
	"github.com/GkIgor/jay-ia/core/internal/storage"
	"github.com/GkIgor/jay-ia/sdk/ipc"
)

func setupTestToolRPCEnvironment(t *testing.T) (*Router, *service.ToolService, *service.RegistrationService) {
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

	regSvc, _ := service.NewRegistrationService(regRepo, ruleRepo, evaluator)
	toolSvc, _ := service.NewToolService(toolRepo, ruleRepo, evaluator)

	toolHandler, _ := NewToolHandler(toolSvc)

	router := NewRouter()
	toolHandler.RegisterRoutes(router)

	return router, toolSvc, regSvc
}

func TestToolHandler_RegisterTool_RPC(t *testing.T) {
	router, _, regSvc := setupTestToolRPCEnvironment(t)
	ctx := context.Background()

	_, _ = regSvc.RegisterClient(ctx, "client_cpp", "")

	reqPayload := ipc.RegisterToolRequest{
		ID:          "calculator",
		Name:        "Calculadora",
		Description: "Realiza cálculos matemáticos",
		Version:     "1.0.0",
		SchemaJSON:  `{"type":"object"}`,
	}
	reqEnv, _ := ipc.NewRequestEnvelope(ipc.MsgRegisterTool, "client_cpp", reqPayload)
	rawBytes, _ := json.Marshal(reqEnv)

	respBytes := router.Dispatch(ctx, rawBytes)
	var respEnv ipc.ResponseEnvelope
	if err := json.Unmarshal(respBytes, &respEnv); err != nil {
		t.Fatalf("falha ao desserializar resposta JSON: %v", err)
	}

	if respEnv.Status != ipc.ErrSuccess {
		t.Fatalf("esperava Status 0, obteve %d", respEnv.Status)
	}

	var regResp ipc.RegisterToolResponse
	if err := ipc.UnmarshalPayload(respEnv.Payload, &regResp); err != nil {
		t.Fatalf("falha ao desserializar payload de resposta: %v", err)
	}

	if regResp.Tool.ID != "calculator" || regResp.Tool.RegistrationID != "client_cpp" {
		t.Errorf("dados de ferramenta incorretos: %+v", regResp.Tool)
	}
}

func TestToolHandler_GetTool_RPC(t *testing.T) {
	router, toolSvc, regSvc := setupTestToolRPCEnvironment(t)
	ctx := context.Background()

	_, _ = regSvc.RegisterClient(ctx, "client_cpp", "")
	_, _ = toolSvc.RegisterTool(ctx, "client_cpp", ipc.RegisterToolRequest{ID: "tool_1", Name: "Ferramenta 1"})

	reqPayload := ipc.GetToolRequest{
		ToolID: "tool_1",
	}
	reqEnv, _ := ipc.NewRequestEnvelope(ipc.MsgGetTool, "client_cpp", reqPayload)
	rawBytes, _ := json.Marshal(reqEnv)

	respBytes := router.Dispatch(ctx, rawBytes)
	var respEnv ipc.ResponseEnvelope
	_ = json.Unmarshal(respBytes, &respEnv)

	if respEnv.Status != ipc.ErrSuccess {
		t.Fatalf("esperava Status 0, obteve %d", respEnv.Status)
	}

	var getResp ipc.GetToolResponse
	_ = ipc.UnmarshalPayload(respEnv.Payload, &getResp)
	if getResp.Tool.ID != "tool_1" {
		t.Errorf("ID de ferramenta incorreto: %s", getResp.Tool.ID)
	}
}

func TestToolHandler_UnregisterTool_RPC(t *testing.T) {
	router, toolSvc, regSvc := setupTestToolRPCEnvironment(t)
	ctx := context.Background()

	_, _ = regSvc.RegisterClient(ctx, "client_cpp", "")
	_, _ = toolSvc.RegisterTool(ctx, "client_cpp", ipc.RegisterToolRequest{ID: "tool_1", Name: "Ferramenta 1"})

	reqPayload := ipc.UnregisterToolRequest{
		ToolID: "tool_1",
	}
	reqEnv, _ := ipc.NewRequestEnvelope(ipc.MsgUnregisterTool, "client_cpp", reqPayload)
	rawBytes, _ := json.Marshal(reqEnv)

	respBytes := router.Dispatch(ctx, rawBytes)
	var respEnv ipc.ResponseEnvelope
	_ = json.Unmarshal(respBytes, &respEnv)

	if respEnv.Status != ipc.ErrSuccess {
		t.Fatalf("esperava Status 0 na exclusão, obteve %d", respEnv.Status)
	}

	var unregResp ipc.UnregisterToolResponse
	_ = ipc.UnmarshalPayload(respEnv.Payload, &unregResp)
	if !unregResp.Success {
		t.Errorf("esperava success = true")
	}
}

func TestToolHandler_ListTools_RPC(t *testing.T) {
	router, toolSvc, regSvc := setupTestToolRPCEnvironment(t)
	ctx := context.Background()

	_, _ = regSvc.RegisterClient(ctx, "client_cpp", "")
	_, _ = toolSvc.RegisterTool(ctx, "client_cpp", ipc.RegisterToolRequest{ID: "t1", Name: "T1"})
	_, _ = toolSvc.RegisterTool(ctx, "client_cpp", ipc.RegisterToolRequest{ID: "t2", Name: "T2"})

	reqPayload := ipc.ListToolsRequest{}
	reqEnv, _ := ipc.NewRequestEnvelope(ipc.MsgListTools, "client_cpp", reqPayload)
	rawBytes, _ := json.Marshal(reqEnv)

	respBytes := router.Dispatch(ctx, rawBytes)
	var respEnv ipc.ResponseEnvelope
	_ = json.Unmarshal(respBytes, &respEnv)

	if respEnv.Status != ipc.ErrSuccess {
		t.Fatalf("esperava Status 0 na listagem, obteve %d", respEnv.Status)
	}

	var listResp ipc.ListToolsResponse
	_ = ipc.UnmarshalPayload(respEnv.Payload, &listResp)
	if len(listResp.Tools) != 2 {
		t.Fatalf("esperava 2 ferramentas na lista, obteve %d", len(listResp.Tools))
	}
}
