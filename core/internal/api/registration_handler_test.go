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

func setupTestRPCEnvironment(t *testing.T) (*Router, *service.RegistrationService) {
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
		t.Fatalf("falha ao executar DDL de teste: %v", err)
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

	svc, err := service.NewRegistrationService(regRepo, ruleRepo, evaluator)
	if err != nil {
		t.Fatalf("falha ao instanciar RegistrationService: %v", err)
	}

	handler, err := NewRegistrationHandler(svc)
	if err != nil {
		t.Fatalf("falha ao instanciar RegistrationHandler: %v", err)
	}

	router := NewRouter()
	handler.RegisterRoutes(router)

	return router, svc
}

func TestRegistrationHandler_RegisterClient_RPC(t *testing.T) {
	router, _ := setupTestRPCEnvironment(t)

	reqPayload := ipc.RegisterClientRequest{
		ClientID: "client_cpp",
		Metadata: `{"platform":"linux"}`,
	}
	reqEnv, _ := ipc.NewRequestEnvelope(ipc.MsgRegisterClient, "client_cpp", reqPayload)
	rawBytes, _ := json.Marshal(reqEnv)

	respBytes := router.Dispatch(context.Background(), rawBytes)
	var respEnv ipc.ResponseEnvelope
	if err := json.Unmarshal(respBytes, &respEnv); err != nil {
		t.Fatalf("falha ao desserializar resposta JSON: %v", err)
	}

	if respEnv.Status != ipc.ErrSuccess {
		t.Fatalf("esperava Status 0, obteve %d", respEnv.Status)
	}

	var regResp ipc.RegisterClientResponse
	if err := ipc.UnmarshalPayload(respEnv.Payload, &regResp); err != nil {
		t.Fatalf("falha ao desserializar payload de resposta: %v", err)
	}

	if regResp.Registration.ID != "client_cpp" || regResp.Registration.Status != 1 {
		t.Errorf("dados de registro incorretos: %+v", regResp.Registration)
	}
}

func TestRegistrationHandler_GetRegistration_RPC(t *testing.T) {
	router, svc := setupTestRPCEnvironment(t)
	ctx := context.Background()

	_, _ = svc.RegisterClient(ctx, "client_cpp", "")

	reqPayload := ipc.GetRegistrationRequest{
		RegistrationID: "client_cpp",
	}
	reqEnv, _ := ipc.NewRequestEnvelope(ipc.MsgGetRegistration, "client_cpp", reqPayload)
	rawBytes, _ := json.Marshal(reqEnv)

	respBytes := router.Dispatch(context.Background(), rawBytes)
	var respEnv ipc.ResponseEnvelope
	_ = json.Unmarshal(respBytes, &respEnv)

	if respEnv.Status != ipc.ErrSuccess {
		t.Fatalf("esperava Status 0, obteve %d", respEnv.Status)
	}

	var getResp ipc.GetRegistrationResponse
	_ = ipc.UnmarshalPayload(respEnv.Payload, &getResp)
	if getResp.Registration.ID != "client_cpp" {
		t.Errorf("esperava registration_id 'client_cpp', obteve %s", getResp.Registration.ID)
	}
}

func TestRegistrationHandler_UnregisterClient_RPC(t *testing.T) {
	router, svc := setupTestRPCEnvironment(t)
	ctx := context.Background()

	_, _ = svc.RegisterClient(ctx, "client_to_del", "")

	reqPayload := ipc.UnregisterClientRequest{
		ClientID: "client_to_del",
	}
	reqEnv, _ := ipc.NewRequestEnvelope(ipc.MsgUnregisterClient, "client_to_del", reqPayload)
	rawBytes, _ := json.Marshal(reqEnv)

	respBytes := router.Dispatch(context.Background(), rawBytes)
	var respEnv ipc.ResponseEnvelope
	_ = json.Unmarshal(respBytes, &respEnv)

	if respEnv.Status != ipc.ErrSuccess {
		t.Fatalf("esperava Status 0 no desregistro, obteve %d", respEnv.Status)
	}
}

func TestRegistrationHandler_ListRegistrations_RPC(t *testing.T) {
	router, svc := setupTestRPCEnvironment(t)
	ctx := context.Background()

	_, _ = svc.RegisterClient(ctx, "client_1", "")
	_, _ = svc.RegisterClient(ctx, "client_2", "")

	reqEnv, _ := ipc.NewRequestEnvelope(ipc.MsgListRegistrations, "client_1", nil)
	rawBytes, _ := json.Marshal(reqEnv)

	respBytes := router.Dispatch(context.Background(), rawBytes)
	var respEnv ipc.ResponseEnvelope
	_ = json.Unmarshal(respBytes, &respEnv)

	if respEnv.Status != ipc.ErrSuccess {
		t.Fatalf("esperava Status 0, obteve %d", respEnv.Status)
	}

	var listResp ipc.ListRegistrationsResponse
	_ = ipc.UnmarshalPayload(respEnv.Payload, &listResp)
	if listResp.Total != 2 || len(listResp.Registrations) != 2 {
		t.Fatalf("esperava 2 registros na lista, obteve %d", listResp.Total)
	}
}

func TestRegistrationHandler_UpdateSharedRules_RPC(t *testing.T) {
	router, svc := setupTestRPCEnvironment(t)
	ctx := context.Background()

	_, _ = svc.RegisterClient(ctx, "owner_1", "")

	reqPayload := ipc.UpdateSharedRulesRequest{
		Rules: []ipc.SharedRulePayload{
			{TargetScope: 0, Pattern: "client_*", MatchType: 2, AllowedActions: 15},
		},
	}
	reqEnv, _ := ipc.NewRequestEnvelope(ipc.MsgUpdateSharedRules, "owner_1", reqPayload)
	rawBytes, _ := json.Marshal(reqEnv)

	respBytes := router.Dispatch(context.Background(), rawBytes)
	var respEnv ipc.ResponseEnvelope
	_ = json.Unmarshal(respBytes, &respEnv)

	if respEnv.Status != ipc.ErrSuccess {
		t.Fatalf("esperava Status 0, obteve %d", respEnv.Status)
	}

	var rulesResp ipc.UpdateSharedRulesResponse
	_ = ipc.UnmarshalPayload(respEnv.Payload, &rulesResp)
	if rulesResp.AppliedRulesCount != 1 {
		t.Errorf("esperava 1 regra aplicada, obteve %d", rulesResp.AppliedRulesCount)
	}
}
