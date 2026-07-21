package storage

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"
)

// ToolStatus representa os estados possíveis de disponibilidade de uma ferramenta no catálogo.
type ToolStatus int

const (
	ToolAvailable  ToolStatus = 1
	ToolDisabled   ToolStatus = 2
	ToolDeprecated ToolStatus = 3
)

// Tool representa uma ferramenta funcional versionada registrada por um consumidor no Jay Core.
// Persiste na tabela `tools`.
type Tool struct {
	ID             string
	RegistrationID string
	Name           string
	Description    string
	Version        string
	SchemaJSON     string
	Status         ToolStatus
	CreatedAt      string
	UpdatedAt      string
}

// ToolRepository implementa as operações de persistência sobre a tabela `tools`.
type ToolRepository struct {
	db *sql.DB
}

// NewToolRepository instancia o repositório de ferramentas.
// Retorna ErrNilDatabase se db for nil.
func NewToolRepository(db *sql.DB) (*ToolRepository, error) {
	if db == nil {
		return nil, ErrNilDatabase
	}
	return &ToolRepository{db: db}, nil
}

// Register insere uma nova ferramenta ou atualiza uma ferramenta existente do mesmo proprietário (Upsert idempotente).
// Retorna ErrOwnershipConflict se a ferramenta já existir pertencendo a outro registration_id (proteção contra hijacking).
func (r *ToolRepository) Register(tool Tool) error {
	cleanID := strings.TrimSpace(tool.ID)
	cleanRegID := strings.TrimSpace(tool.RegistrationID)
	cleanName := strings.TrimSpace(tool.Name)

	if cleanID == "" || cleanRegID == "" || cleanName == "" {
		return ErrInvalidArgument
	}

	// Aplicação de valores padrão (fallbacks)
	if strings.TrimSpace(tool.Version) == "" {
		tool.Version = "1.0.0"
	}
	if tool.Status == 0 {
		tool.Status = ToolAvailable
	}
	if strings.TrimSpace(tool.SchemaJSON) == "" {
		tool.SchemaJSON = "{}"
	}

	now := time.Now().UTC().Format(time.RFC3339)
	query := `
		INSERT INTO tools (id, registration_id, name, description, version, schema_json, status, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET
			name         = excluded.name,
			description  = excluded.description,
			version      = excluded.version,
			schema_json  = excluded.schema_json,
			status       = excluded.status,
			updated_at   = excluded.updated_at
		WHERE tools.registration_id = excluded.registration_id;
	`
	res, err := r.db.Exec(query, cleanID, cleanRegID, cleanName, tool.Description, tool.Version, tool.SchemaJSON, tool.Status, now)
	if err != nil {
		return mapToolError(err)
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("storage: falha ao verificar linhas afetadas no Register da ferramenta: %w", err)
	}

	// Se RowsAffected == 0, significa que a ferramenta já existia mas pertencia a outro registration_id
	if rows == 0 {
		var existingOwner string
		err := r.db.QueryRow(`SELECT registration_id FROM tools WHERE id = ?;`, cleanID).Scan(&existingOwner)
		if err == nil && existingOwner != cleanRegID {
			return ErrOwnershipConflict
		}
	}

	return nil
}

// FindByID busca uma ferramenta pelo seu ID.
// Retorna ErrNotFound caso a ferramenta não exista.
func (r *ToolRepository) FindByID(id string) (*Tool, error) {
	cleanID := strings.TrimSpace(id)
	if cleanID == "" {
		return nil, ErrInvalidArgument
	}

	query := `
		SELECT id, registration_id, name, description, version, schema_json, status, created_at, updated_at
		FROM tools
		WHERE id = ?;
	`
	row := r.db.QueryRow(query, cleanID)

	var t Tool
	err := row.Scan(&t.ID, &t.RegistrationID, &t.Name, &t.Description, &t.Version, &t.SchemaJSON, &t.Status, &t.CreatedAt, &t.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("storage: falha ao buscar ferramenta %s: %w", cleanID, err)
	}

	return &t, nil
}

// ListByRegistration retorna todas as ferramentas pertencentes a uma identidade registrada, ordenadas por name ASC.
func (r *ToolRepository) ListByRegistration(registrationID string) ([]*Tool, error) {
	cleanRegID := strings.TrimSpace(registrationID)
	if cleanRegID == "" {
		return nil, ErrInvalidArgument
	}

	query := `
		SELECT id, registration_id, name, description, version, schema_json, status, created_at, updated_at
		FROM tools
		WHERE registration_id = ?
		ORDER BY name ASC;
	`
	rows, err := r.db.Query(query, cleanRegID)
	if err != nil {
		return nil, fmt.Errorf("storage: falha ao listar ferramentas do registro %s: %w", cleanRegID, err)
	}
	defer rows.Close()

	tools := []*Tool{}
	for rows.Next() {
		var t Tool
		err := rows.Scan(&t.ID, &t.RegistrationID, &t.Name, &t.Description, &t.Version, &t.SchemaJSON, &t.Status, &t.CreatedAt, &t.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("storage: falha ao ler linha de ferramenta: %w", err)
		}
		tools = append(tools, &t)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("storage: erro durante iteração de ferramentas: %w", err)
	}

	return tools, nil
}

// ListAvailable retorna todas as ferramentas ativas (status == ToolAvailable (1)) no sistema, ordenadas por name ASC.
func (r *ToolRepository) ListAvailable() ([]*Tool, error) {
	query := `
		SELECT id, registration_id, name, description, version, schema_json, status, created_at, updated_at
		FROM tools
		WHERE status = 1
		ORDER BY name ASC;
	`
	rows, err := r.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("storage: falha ao listar ferramentas disponíveis: %w", err)
	}
	defer rows.Close()

	tools := []*Tool{}
	for rows.Next() {
		var t Tool
		err := rows.Scan(&t.ID, &t.RegistrationID, &t.Name, &t.Description, &t.Version, &t.SchemaJSON, &t.Status, &t.CreatedAt, &t.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("storage: falha ao ler linha de ferramenta disponível: %w", err)
		}
		tools = append(tools, &t)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("storage: erro durante iteração de ferramentas disponíveis: %w", err)
	}

	return tools, nil
}

// UpdateStatus atualiza o estado de disponibilidade (status) de uma ferramenta existente.
// Retorna ErrNotFound se a ferramenta não existir.
func (r *ToolRepository) UpdateStatus(id string, status ToolStatus) error {
	cleanID := strings.TrimSpace(id)
	if cleanID == "" || status < ToolAvailable || status > ToolDeprecated {
		return ErrInvalidArgument
	}

	now := time.Now().UTC().Format(time.RFC3339)
	query := `UPDATE tools SET status = ?, updated_at = ? WHERE id = ?;`
	res, err := r.db.Exec(query, status, now, cleanID)
	if err != nil {
		return mapToolError(err)
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("storage: falha ao verificar linhas afetadas no UpdateStatus: %w", err)
	}

	if rows == 0 {
		return ErrNotFound
	}

	return nil
}

// Delete realiza a remoção física (Hard Delete) de uma ferramenta pelo seu ID.
// É uma operação 100% idempotente: retorna nil se a ferramenta já não existia.
func (r *ToolRepository) Delete(id string) error {
	cleanID := strings.TrimSpace(id)
	if cleanID == "" {
		return ErrInvalidArgument
	}

	query := `DELETE FROM tools WHERE id = ?;`
	_, err := r.db.Exec(query, cleanID)
	if err != nil {
		return mapToolError(err)
	}

	return nil
}

// mapToolError converte erros técnicos de banco em erros semânticos do recurso Tool.
func mapToolError(err error) error {
	techErr := translateSQLiteError(err)
	switch {
	case errors.Is(techErr, ErrForeignKeyViolation):
		return ErrInvalidRegistration
	case errors.Is(techErr, ErrUniqueViolation):
		return ErrAlreadyExists
	default:
		return techErr
	}
}
