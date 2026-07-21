package storage

import (
	"errors"
	"testing"
)

func TestSharedRuleRepository_NilDatabase(t *testing.T) {
	_, err := NewSharedRuleRepository(nil)
	if !errors.Is(err, ErrNilDatabase) {
		t.Fatalf("esperava ErrNilDatabase, obteve: %v", err)
	}
}

func TestSharedRuleRepo_ReplaceRules_Success(t *testing.T) {
	engine := newTestMigratedDB(t)
	regRepo, _ := NewRegistrationRepository(engine.DB())
	ruleRepo, _ := NewSharedRuleRepository(engine.DB())

	_ = regRepo.Create(Registration{ID: "reg-owner", Status: RegistrationActive})

	rules := []SharedRule{
		{
			TargetScope:    ScopeChats,
			Pattern:        "slack_*",
			MatchType:      MatchPrefix,
			AllowedActions: ActionRead | ActionWrite,
		},
		{
			TargetScope:    ScopeMessages,
			Pattern:        "jay_client_cli",
			MatchType:      MatchExact,
			AllowedActions: ActionAll,
		},
	}

	count, err := ruleRepo.ReplaceRules("reg-owner", rules)
	if err != nil {
		t.Fatalf("falha no ReplaceRules: %v", err)
	}
	if count != 2 {
		t.Fatalf("esperava 2 regras inseridas, obteve %d", count)
	}

	fetched, err := ruleRepo.ListByRegistration("reg-owner")
	if err != nil {
		t.Fatalf("falha no ListByRegistration: %v", err)
	}
	if len(fetched) != 2 {
		t.Fatalf("esperava 2 regras na lista, obteve %d", len(fetched))
	}

	if fetched[0].RegistrationID != "reg-owner" || fetched[0].Pattern != "slack_*" || fetched[0].MatchType != MatchPrefix {
		t.Errorf("dados incorretos na regra 0: %+v", fetched[0])
	}
	if fetched[1].RegistrationID != "reg-owner" || fetched[1].Pattern != "jay_client_cli" || fetched[1].AllowedActions != ActionAll {
		t.Errorf("dados incorretos na regra 1: %+v", fetched[1])
	}
}

func TestSharedRuleRepo_ReplaceRules_AtomicityOverwrite(t *testing.T) {
	engine := newTestMigratedDB(t)
	regRepo, _ := NewRegistrationRepository(engine.DB())
	ruleRepo, _ := NewSharedRuleRepository(engine.DB())

	_ = regRepo.Create(Registration{ID: "reg-owner", Status: RegistrationActive})

	// Conjunto 1: 2 regras
	r1 := []SharedRule{
		{Pattern: "p1", TargetScope: ScopeChats},
		{Pattern: "p2", TargetScope: ScopeMessages},
	}
	_, _ = ruleRepo.ReplaceRules("reg-owner", r1)

	// Conjunto 2: Substitui por 1 nova regra
	r2 := []SharedRule{
		{Pattern: "p3_novo", TargetScope: ScopeAll},
	}
	count, err := ruleRepo.ReplaceRules("reg-owner", r2)
	if err != nil {
		t.Fatalf("falha ao sobrescrever regras: %v", err)
	}
	if count != 1 {
		t.Fatalf("esperava 1 regra inserida, obteve %d", count)
	}

	fetched, _ := ruleRepo.ListByRegistration("reg-owner")
	if len(fetched) != 1 {
		t.Fatalf("esperava apenas 1 regra no banco após sobrescrever, obteve %d", len(fetched))
	}
	if fetched[0].Pattern != "p3_novo" {
		t.Errorf("esperava novo pattern p3_novo, obteve %s", fetched[0].Pattern)
	}
}

func TestSharedRuleRepo_ReplaceRules_OverrideStructID(t *testing.T) {
	engine := newTestMigratedDB(t)
	regRepo, _ := NewRegistrationRepository(engine.DB())
	ruleRepo, _ := NewSharedRuleRepository(engine.DB())

	_ = regRepo.Create(Registration{ID: "reg-owner-real", Status: RegistrationActive})

	rules := []SharedRule{
		{
			RegistrationID: "reg-owner-falso-divergente", // Deve ser sobrescrito pelo parâmetro da função
			Pattern:        "pattern_test",
		},
	}

	_, err := ruleRepo.ReplaceRules("reg-owner-real", rules)
	if err != nil {
		t.Fatalf("falha no ReplaceRules: %v", err)
	}

	fetched, _ := ruleRepo.ListByRegistration("reg-owner-real")
	if len(fetched) != 1 {
		t.Fatalf("esperava 1 regra, obteve %d", len(fetched))
	}
	if fetched[0].RegistrationID != "reg-owner-real" {
		t.Fatalf("RegistrationID não foi sobrescrito pelo parâmetro. Obteve %s", fetched[0].RegistrationID)
	}
}

func TestSharedRuleRepo_ReplaceRules_EmptyList(t *testing.T) {
	engine := newTestMigratedDB(t)
	regRepo, _ := NewRegistrationRepository(engine.DB())
	ruleRepo, _ := NewSharedRuleRepository(engine.DB())

	_ = regRepo.Create(Registration{ID: "reg-owner", Status: RegistrationActive})

	r1 := []SharedRule{{Pattern: "rule-1"}}
	_, _ = ruleRepo.ReplaceRules("reg-owner", r1)

	// Substitui por lista vazia
	count, err := ruleRepo.ReplaceRules("reg-owner", []SharedRule{})
	if err != nil {
		t.Fatalf("falha no ReplaceRules com lista vazia: %v", err)
	}
	if count != 0 {
		t.Fatalf("esperava 0 regras retornadas, obteve %d", count)
	}

	fetched, _ := ruleRepo.ListByRegistration("reg-owner")
	if len(fetched) != 0 {
		t.Fatalf("esperava 0 regras após limpeza com lista vazia, obteve %d", len(fetched))
	}
}

func TestSharedRuleRepo_ReplaceRules_InvalidRegistration(t *testing.T) {
	engine := newTestMigratedDB(t)
	ruleRepo, _ := NewSharedRuleRepository(engine.DB())

	rules := []SharedRule{{Pattern: "test"}}
	_, err := ruleRepo.ReplaceRules("reg-inexistente", rules)
	if !errors.Is(err, ErrInvalidRegistration) {
		t.Fatalf("esperava ErrInvalidRegistration ao referenciar registro inexistente, obteve: %v", err)
	}
}

func TestSharedRuleRepo_ReplaceRules_WhitespacePattern(t *testing.T) {
	engine := newTestMigratedDB(t)
	regRepo, _ := NewRegistrationRepository(engine.DB())
	ruleRepo, _ := NewSharedRuleRepository(engine.DB())

	_ = regRepo.Create(Registration{ID: "reg-owner", Status: RegistrationActive})

	rules := []SharedRule{{Pattern: "   "}} // Espaços em branco deve ser considerado inválido
	_, err := ruleRepo.ReplaceRules("reg-owner", rules)
	if !errors.Is(err, ErrInvalidArgument) {
		t.Fatalf("esperava ErrInvalidArgument para pattern com espaços em branco, obteve: %v", err)
	}
}

func TestSharedRuleRepo_ListByRegistration_Empty(t *testing.T) {
	engine := newTestMigratedDB(t)
	regRepo, _ := NewRegistrationRepository(engine.DB())
	ruleRepo, _ := NewSharedRuleRepository(engine.DB())

	_ = regRepo.Create(Registration{ID: "reg-owner", Status: RegistrationActive})

	fetched, err := ruleRepo.ListByRegistration("reg-owner")
	if err != nil {
		t.Fatalf("falha no ListByRegistration: %v", err)
	}
	if fetched == nil {
		t.Fatalf("esperava slice não-nil para registro sem regras, obteve nil")
	}
	if len(fetched) != 0 {
		t.Fatalf("esperava len 0, obteve %d", len(fetched))
	}
}

func TestSharedRuleRepo_DeleteByRegistration_Success(t *testing.T) {
	engine := newTestMigratedDB(t)
	regRepo, _ := NewRegistrationRepository(engine.DB())
	ruleRepo, _ := NewSharedRuleRepository(engine.DB())

	_ = regRepo.Create(Registration{ID: "reg-owner", Status: RegistrationActive})
	_, _ = ruleRepo.ReplaceRules("reg-owner", []SharedRule{{Pattern: "p1"}})

	if err := ruleRepo.DeleteByRegistration("reg-owner"); err != nil {
		t.Fatalf("falha no DeleteByRegistration: %v", err)
	}

	fetched, _ := ruleRepo.ListByRegistration("reg-owner")
	if len(fetched) != 0 {
		t.Fatalf("esperava 0 regras após delete, obteve %d", len(fetched))
	}

	// Segunda chamada de DeleteByRegistration deve ser idempotente
	if err := ruleRepo.DeleteByRegistration("reg-owner"); err != nil {
		t.Fatalf("esperava nil no DeleteByRegistration idempotente, obteve: %v", err)
	}
}

func TestSharedRuleRepo_CascadeOnRegistrationDelete(t *testing.T) {
	engine := newTestMigratedDB(t)
	regRepo, _ := NewRegistrationRepository(engine.DB())
	ruleRepo, _ := NewSharedRuleRepository(engine.DB())

	_ = regRepo.Create(Registration{ID: "reg-to-delete", Status: RegistrationActive})
	_, _ = ruleRepo.ReplaceRules("reg-to-delete", []SharedRule{{Pattern: "p1"}, {Pattern: "p2"}})

	// Remove o Registration pai
	if err := regRepo.Delete("reg-to-delete"); err != nil {
		t.Fatalf("falha ao deletar registration pai: %v", err)
	}

	// As shared_rules filhas devem ser apagadas via SQL CASCADE
	var count int
	err := engine.DB().QueryRow(`SELECT COUNT(1) FROM shared_rules WHERE registration_id = 'reg-to-delete'`).Scan(&count)
	if err != nil {
		t.Fatalf("falha ao verificar contagem física de shared_rules: %v", err)
	}
	if count != 0 {
		t.Fatalf("esperava 0 regras filhas após CASCADE delete no registration pai, obteve %d", count)
	}
}
