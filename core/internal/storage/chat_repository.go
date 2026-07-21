package storage

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"
)

// ChatStatus representa os estados possíveis do ciclo de vida de um Chat.
type ChatStatus int

const (
	ChatActive   ChatStatus = 1
	ChatArchived ChatStatus = 2
	ChatDeleted  ChatStatus = 3
)

// ChatFilter representa as opções de filtragem para listagem de chats por proprietário.
type ChatFilter int

const (
	ChatFilterActiveOnly      ChatFilter = 0
	ChatFilterIncludeArchived ChatFilter = 1
)

// Chat representa um container de conversa no Jay Core.
// Persiste na tabela `chats`.
type Chat struct {
	ID                  string
	OwnerRegistrationID string
	Title               string
	Status              ChatStatus
	MetadataJSON        string
	CreatedAt           string
	UpdatedAt           string
}

// ChatRepository implementa as operações de persistência sobre a tabela `chats`.
type ChatRepository struct {
	db *sql.DB
}

// NewChatRepository instancia o repositório de chats.
// Retorna ErrNilDatabase se db for nil.
func NewChatRepository(db *sql.DB) (*ChatRepository, error) {
	if db == nil {
		return nil, ErrNilDatabase
	}
	return &ChatRepository{db: db}, nil
}

// Create insere um novo Chat no banco de dados.
// Retorna ErrInvalidArgument se ID ou OwnerRegistrationID forem vazios, ou se o status inicial for ChatDeleted.
func (r *ChatRepository) Create(chat Chat) error {
	if strings.TrimSpace(chat.ID) == "" || strings.TrimSpace(chat.OwnerRegistrationID) == "" {
		return ErrInvalidArgument
	}

	if chat.Status == ChatDeleted {
		return ErrInvalidArgument
	}

	if chat.Status == 0 {
		chat.Status = ChatActive
	}

	query := `INSERT INTO chats (id, owner_registration_id, title, status, metadata_json) VALUES (?, ?, ?, ?, ?);`
	_, err := r.db.Exec(query, chat.ID, chat.OwnerRegistrationID, chat.Title, chat.Status, chat.MetadataJSON)
	if err != nil {
		return mapChatError(err)
	}

	return nil
}

// FindByID busca um Chat ativo ou arquivado pelo seu ID.
// Retorna ErrNotFound se o chat não existir ou estiver com Soft Delete (status = ChatDeleted).
func (r *ChatRepository) FindByID(id string) (*Chat, error) {
	if strings.TrimSpace(id) == "" {
		return nil, ErrInvalidArgument
	}

	query := `SELECT id, owner_registration_id, title, status, metadata_json, created_at, updated_at FROM chats WHERE id = ? AND status != 3;`
	row := r.db.QueryRow(query, id)

	var chat Chat
	err := row.Scan(&chat.ID, &chat.OwnerRegistrationID, &chat.Title, &chat.Status, &chat.MetadataJSON, &chat.CreatedAt, &chat.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("storage: falha ao buscar chat %s: %w", id, err)
	}

	return &chat, nil
}

// ListByOwner retorna a lista de chats pertencentes a uma identidade lógica registrada,
// ordenados por updated_at DESC, created_at DESC. Chats com status ChatDeleted nunca são retornados.
func (r *ChatRepository) ListByOwner(ownerRegistrationID string, filter ChatFilter) ([]*Chat, error) {
	if strings.TrimSpace(ownerRegistrationID) == "" {
		return nil, ErrInvalidArgument
	}

	var query string
	switch filter {
	case ChatFilterIncludeArchived:
		query = `SELECT id, owner_registration_id, title, status, metadata_json, created_at, updated_at FROM chats WHERE owner_registration_id = ? AND status IN (1, 2) ORDER BY updated_at DESC, created_at DESC;`
	default:
		query = `SELECT id, owner_registration_id, title, status, metadata_json, created_at, updated_at FROM chats WHERE owner_registration_id = ? AND status = 1 ORDER BY updated_at DESC, created_at DESC;`
	}

	rows, err := r.db.Query(query, ownerRegistrationID)
	if err != nil {
		return nil, fmt.Errorf("storage: falha ao listar chats do proprietário %s: %w", ownerRegistrationID, err)
	}
	defer rows.Close()

	chats := []*Chat{}
	for rows.Next() {
		var chat Chat
		if err := rows.Scan(&chat.ID, &chat.OwnerRegistrationID, &chat.Title, &chat.Status, &chat.MetadataJSON, &chat.CreatedAt, &chat.UpdatedAt); err != nil {
			return nil, fmt.Errorf("storage: falha ao ler linha de chat: %w", err)
		}
		chats = append(chats, &chat)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("storage: erro durante iteração de chats: %w", err)
	}

	return chats, nil
}

// Update atualiza title, status e metadata_json de um chat existente.
// Proíbe a transição para ChatDeleted via Update.
// Retorna ErrNotFound se o chat não existir ou já estiver deletado.
func (r *ChatRepository) Update(chat Chat) error {
	if strings.TrimSpace(chat.ID) == "" {
		return ErrInvalidArgument
	}

	if chat.Status == ChatDeleted {
		return ErrInvalidArgument
	}

	now := time.Now().UTC().Format(time.RFC3339)
	query := `UPDATE chats SET title = ?, status = ?, metadata_json = ?, updated_at = ? WHERE id = ? AND status != 3;`
	res, err := r.db.Exec(query, chat.Title, chat.Status, chat.MetadataJSON, now, chat.ID)
	if err != nil {
		return mapChatError(err)
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("storage: falha ao verificar linhas afetadas no update: %w", err)
	}

	if rows == 0 {
		return ErrNotFound
	}

	return nil
}

// Delete realiza o Soft Delete de um chat configurando status = ChatDeleted (3) e atualizando updated_at.
// É uma operação idempotente: se o chat já estiver deletado (status = 3), retorna nil.
// Retorna ErrNotFound se o ID não existir no banco de dados.
func (r *ChatRepository) Delete(id string) error {
	if strings.TrimSpace(id) == "" {
		return ErrInvalidArgument
	}

	now := time.Now().UTC().Format(time.RFC3339)
	query := `UPDATE chats SET status = 3, updated_at = ? WHERE id = ? AND status != 3;`
	res, err := r.db.Exec(query, now, id)
	if err != nil {
		return mapChatError(err)
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("storage: falha ao verificar linhas afetadas no delete: %w", err)
	}

	if rows > 0 {
		return nil
	}

	// Se nenhuma linha foi alterada, verifica se o chat existe (já com status = 3) ou não existe
	var dummyStatus int
	err = r.db.QueryRow(`SELECT status FROM chats WHERE id = ?;`, id).Scan(&dummyStatus)
	if err != nil {
		if err == sql.ErrNoRows {
			return ErrNotFound
		}
		return fmt.Errorf("storage: falha ao verificar existência do chat %s: %w", id, err)
	}

	// Chat existe e já está deletado (status == 3) -> idempotência
	return nil
}

// mapChatError converte erros técnicos de banco em erros semânticos do recurso Chat.
func mapChatError(err error) error {
	techErr := translateSQLiteError(err)
	switch {
	case errors.Is(techErr, ErrUniqueViolation):
		return ErrAlreadyExists
	case errors.Is(techErr, ErrForeignKeyViolation):
		return ErrInvalidOwner
	default:
		return techErr
	}
}
