package storage

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"
)

// RegistrationStatus representa o estado de atividade de uma identidade lógica registrada.
type RegistrationStatus int

const (
	RegistrationActive    RegistrationStatus = 1
	RegistrationInactive  RegistrationStatus = 2
	RegistrationSuspended RegistrationStatus = 3
)

// Registration representa uma identidade lógica conhecida pelo Jay Core.
// Persiste na tabela `registrations`.
type Registration struct {
	ID           string
	MetadataJSON string
	Status       RegistrationStatus
	CreatedAt    string
	UpdatedAt    string
}

// RegistrationRepository implementa as operações de persistência sobre a tabela `registrations`.
type RegistrationRepository struct {
	db *sql.DB
}

// NewRegistrationRepository instancia o repositório.
// Retorna ErrNilDatabase se db for nil.
func NewRegistrationRepository(db *sql.DB) (*RegistrationRepository, error) {
	if db == nil {
		return nil, ErrNilDatabase
	}
	return &RegistrationRepository{db: db}, nil
}

// Create insere um novo Registration no banco.
// created_at e updated_at são preenchidos pelo banco via DEFAULT.
func (r *RegistrationRepository) Create(reg Registration) error {
	if strings.TrimSpace(reg.ID) == "" {
		return ErrInvalidArgument
	}

	query := `INSERT INTO registrations (id, metadata_json, status) VALUES (?, ?, ?);`
	_, err := r.db.Exec(query, reg.ID, reg.MetadataJSON, reg.Status)
	if err != nil {
		return mapRegistrationError(err)
	}

	return nil
}

// Upsert insere um novo Registration ou atualiza um existente em caso de conflito de ID.
// Em caso de update, metadata_json, status e updated_at são atualizados.
// created_at nunca é alterado pelo Upsert.
func (r *RegistrationRepository) Upsert(reg Registration) error {
	if strings.TrimSpace(reg.ID) == "" {
		return ErrInvalidArgument
	}

	now := time.Now().UTC().Format(time.RFC3339)
	query := `
		INSERT INTO registrations (id, metadata_json, status, updated_at)
		VALUES (?, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET
			metadata_json = excluded.metadata_json,
			status        = excluded.status,
			updated_at    = excluded.updated_at;
	`
	_, err := r.db.Exec(query, reg.ID, reg.MetadataJSON, reg.Status, now)
	if err != nil {
		return mapRegistrationError(err)
	}

	return nil
}

// FindByID busca um Registration pelo seu ID.
// Retorna ErrNotFound caso o registro não exista.
func (r *RegistrationRepository) FindByID(id string) (*Registration, error) {
	if strings.TrimSpace(id) == "" {
		return nil, ErrInvalidArgument
	}

	query := `SELECT id, metadata_json, status, created_at, updated_at FROM registrations WHERE id = ?;`
	row := r.db.QueryRow(query, id)

	var reg Registration
	err := row.Scan(&reg.ID, &reg.MetadataJSON, &reg.Status, &reg.CreatedAt, &reg.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("storage: falha ao buscar registration %s: %w", id, err)
	}

	return &reg, nil
}

// List retorna todas as identidades lógicas registradas, ordenadas por created_at ASC.
func (r *RegistrationRepository) List() ([]*Registration, error) {
	query := `SELECT id, metadata_json, status, created_at, updated_at FROM registrations ORDER BY created_at ASC;`
	rows, err := r.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("storage: falha ao listar registrations: %w", err)
	}
	defer rows.Close()

	regs := []*Registration{}
	for rows.Next() {
		var reg Registration
		if err := rows.Scan(&reg.ID, &reg.MetadataJSON, &reg.Status, &reg.CreatedAt, &reg.UpdatedAt); err != nil {
			return nil, fmt.Errorf("storage: falha ao ler linha de registration: %w", err)
		}
		regs = append(regs, &reg)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("storage: erro durante iteração de registrations: %w", err)
	}

	return regs, nil
}

// Delete remove fisicamente uma identidade lógica pelo seu ID.
// É uma operação idempotente: retorna nil mesmo se o registro já não existia.
// Retorna ErrDeleteRestricted se existirem dependências ativas via FK RESTRICT.
func (r *RegistrationRepository) Delete(id string) error {
	if strings.TrimSpace(id) == "" {
		return ErrInvalidArgument
	}

	query := `DELETE FROM registrations WHERE id = ?;`
	_, err := r.db.Exec(query, id)
	if err != nil {
		return mapRegistrationError(err)
	}

	return nil
}

// mapRegistrationError converte erros técnicos de banco em erros semânticos de Registration.
func mapRegistrationError(err error) error {
	techErr := translateSQLiteError(err)
	switch {
	case errors.Is(techErr, ErrUniqueViolation):
		return ErrAlreadyExists
	case errors.Is(techErr, ErrForeignKeyViolation):
		return ErrDeleteRestricted
	default:
		return techErr
	}
}
