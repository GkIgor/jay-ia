package storage

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"
)

// RuleScope representa o escopo de recursos afetados pela regra de compartilhamento.
type RuleScope int

const (
	ScopeAll      RuleScope = 0
	ScopeChats    RuleScope = 1
	ScopeMessages RuleScope = 2
	ScopeTools    RuleScope = 3
)

// MatchType representa o algoritmo de casamento aplicado ao padrão (pattern).
type MatchType int

const (
	MatchExact    MatchType = 1
	MatchPrefix   MatchType = 2
	MatchWildcard MatchType = 3
	MatchRegex    MatchType = 4
)

// PermissionAction representa a máscara de bits das ações permitidas pela regra.
type PermissionAction int

const (
	ActionRead    PermissionAction = 1
	ActionWrite   PermissionAction = 2
	ActionExecute PermissionAction = 4
	ActionAdmin   PermissionAction = 8
	ActionAll     PermissionAction = 15 // (1 | 2 | 4 | 8)
)

// SharedRule representa uma regra de compartilhamento declarada por uma identidade lógica (Registration).
// Persiste na tabela `shared_rules`.
type SharedRule struct {
	ID             int64
	RegistrationID string
	TargetScope    RuleScope
	Pattern        string
	MatchType      MatchType
	AllowedActions PermissionAction
	CreatedAt      string
}

// SharedRuleRepository implementa as operações de persistência sobre a tabela `shared_rules`.
type SharedRuleRepository struct {
	db *sql.DB
}

// NewSharedRuleRepository instancia o repositório de regras de compartilhamento.
// Retorna ErrNilDatabase se db for nil.
func NewSharedRuleRepository(db *sql.DB) (*SharedRuleRepository, error) {
	if db == nil {
		return nil, ErrNilDatabase
	}
	return &SharedRuleRepository{db: db}, nil
}

// ReplaceRules substitui atomicamente todas as regras de compartilhamento associadas a um registrationID.
// O parâmetro registrationID é a única fonte da verdade e sobrescreve o campo RegistrationID das structs.
// Retorna a quantidade de regras aplicadas e nil em caso de sucesso.
func (r *SharedRuleRepository) ReplaceRules(registrationID string, rules []SharedRule) (int, error) {
	cleanRegID := strings.TrimSpace(registrationID)
	if cleanRegID == "" {
		return 0, ErrInvalidArgument
	}

	for _, rule := range rules {
		if strings.TrimSpace(rule.Pattern) == "" {
			return 0, ErrInvalidArgument
		}
	}

	tx, err := r.db.Begin()
	if err != nil {
		return 0, fmt.Errorf("storage: falha ao iniciar transação no ReplaceRules: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	// 1. Remove todas as regras anteriores do registro
	_, err = tx.Exec(`DELETE FROM shared_rules WHERE registration_id = ?;`, cleanRegID)
	if err != nil {
		return 0, mapSharedRuleError(err)
	}

	// 2. Insere a nova lista de regras
	queryInsert := `
		INSERT INTO shared_rules (registration_id, target_scope, pattern, match_type, allowed_actions)
		VALUES (?, ?, ?, ?, ?);
	`
	for _, rule := range rules {
		matchType := rule.MatchType
		if matchType == 0 {
			matchType = MatchExact
		}
		allowedActions := rule.AllowedActions
		if allowedActions == 0 {
			allowedActions = ActionRead
		}

		_, err := tx.Exec(queryInsert, cleanRegID, rule.TargetScope, strings.TrimSpace(rule.Pattern), matchType, allowedActions)
		if err != nil {
			return 0, mapSharedRuleError(err)
		}
	}

	if err := tx.Commit(); err != nil {
		return 0, fmt.Errorf("storage: falha ao confirmar transação no ReplaceRules: %w", err)
	}

	return len(rules), nil
}

// ListByRegistration retorna todas as regras de compartilhamento pertencentes a uma identidade registrada, ordenadas por id ASC.
func (r *SharedRuleRepository) ListByRegistration(registrationID string) ([]*SharedRule, error) {
	cleanRegID := strings.TrimSpace(registrationID)
	if cleanRegID == "" {
		return nil, ErrInvalidArgument
	}

	query := `
		SELECT id, registration_id, target_scope, pattern, match_type, allowed_actions, created_at
		FROM shared_rules
		WHERE registration_id = ?
		ORDER BY id ASC;
	`
	rows, err := r.db.Query(query, cleanRegID)
	if err != nil {
		return nil, fmt.Errorf("storage: falha ao listar regras de compartilhamento para %s: %w", cleanRegID, err)
	}
	defer rows.Close()

	rules := []*SharedRule{}
	for rows.Next() {
		var rule SharedRule
		err := rows.Scan(
			&rule.ID, &rule.RegistrationID, &rule.TargetScope, &rule.Pattern,
			&rule.MatchType, &rule.AllowedActions, &rule.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("storage: falha ao ler linha de regra de compartilhamento: %w", err)
		}
		rules = append(rules, &rule)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("storage: erro durante iteração de regras de compartilhamento: %w", err)
	}

	return rules, nil
}

// DeleteByRegistration remove todas as regras de compartilhamento associadas a um registrationID.
// É uma operação idempotente (retorna nil mesmo se nenhuma regra existia).
func (r *SharedRuleRepository) DeleteByRegistration(registrationID string) error {
	cleanRegID := strings.TrimSpace(registrationID)
	if cleanRegID == "" {
		return ErrInvalidArgument
	}

	query := `DELETE FROM shared_rules WHERE registration_id = ?;`
	_, err := r.db.Exec(query, cleanRegID)
	if err != nil {
		return mapSharedRuleError(err)
	}

	return nil
}

// mapSharedRuleError converte erros técnicos de banco em erros semânticos do recurso SharedRule.
func mapSharedRuleError(err error) error {
	techErr := translateSQLiteError(err)
	switch {
	case errors.Is(techErr, ErrForeignKeyViolation):
		return ErrInvalidRegistration
	default:
		return techErr
	}
}
