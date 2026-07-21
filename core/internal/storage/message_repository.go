package storage

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"
)

// AuthorType representa o tipo da entidade autora de uma mensagem.
type AuthorType int

const (
	AuthorRegistration AuthorType = 1
	AuthorAgent        AuthorType = 2
	AuthorTool         AuthorType = 3
	AuthorSystem       AuthorType = 4
)

// MessageRole representa o papel da mensagem no diálogo.
type MessageRole int

const (
	RoleUser      MessageRole = 1
	RoleAssistant MessageRole = 2
	RoleSystem    MessageRole = 3
	RoleTool      MessageRole = 4
)

// MessageContentType representa a tipagem de conteúdo armazenado na mensagem.
type MessageContentType int

const (
	ContentTypeTextPlain  MessageContentType = 1
	ContentTypeMarkdown   MessageContentType = 2
	ContentTypeJSON       MessageContentType = 3
	ContentTypeToolCall   MessageContentType = 4
	ContentTypeToolResult MessageContentType = 5
)

// MessageStatus representa os estados possíveis do ciclo de vida de uma mensagem.
type MessageStatus int

const (
	MessageSent    MessageStatus = 1
	MessageEdited  MessageStatus = 2
	MessageDeleted MessageStatus = 3
)

// Message representa uma mensagem persistida no histórico de um Chat.
type Message struct {
	ID           string
	ChatID       string
	AuthorType   AuthorType
	AuthorID     string
	Role         MessageRole
	Content      string
	ContentType  MessageContentType
	Status       MessageStatus
	SequenceNo   int
	MetadataJSON string
	CreatedAt    string
	UpdatedAt    string
}

// MessageRepository implementa as operações de persistência sobre a tabela `messages`.
type MessageRepository struct {
	db *sql.DB
}

// NewMessageRepository instancia o repositório de mensagens.
// Retorna ErrNilDatabase se db for nil.
func NewMessageRepository(db *sql.DB) (*MessageRepository, error) {
	if db == nil {
		return nil, ErrNilDatabase
	}
	return &MessageRepository{db: db}, nil
}

// Create insere uma nova mensagem no banco de dados.
// Se SequenceNo == 0, calcula e atribui atômica e transacionalmente MAX(sequence_no) + 1.
// Se SequenceNo > 0 for fornecido e já existir no chat, retorna ErrAlreadyExists.
func (r *MessageRepository) Create(msg Message) error {
	if strings.TrimSpace(msg.ID) == "" || strings.TrimSpace(msg.ChatID) == "" || strings.TrimSpace(msg.AuthorID) == "" {
		return ErrInvalidArgument
	}

	if msg.Status == MessageDeleted {
		return ErrInvalidArgument
	}

	// Aplicação de valores padrão
	if msg.Status == 0 {
		msg.Status = MessageSent
	}
	if msg.ContentType == 0 {
		msg.ContentType = ContentTypeTextPlain
	}

	tx, err := r.db.Begin()
	if err != nil {
		return fmt.Errorf("storage: falha ao iniciar transação no Create mensagem: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	if msg.SequenceNo > 0 {
		// Verifica duplicação para sequence_no explícito
		var count int
		err := tx.QueryRow(`SELECT COUNT(1) FROM messages WHERE chat_id = ? AND sequence_no = ?;`, msg.ChatID, msg.SequenceNo).Scan(&count)
		if err != nil {
			return fmt.Errorf("storage: falha ao verificar unicidade do sequence_no explícito: %w", err)
		}
		if count > 0 {
			return ErrAlreadyExists
		}
	} else {
		// Atribuição automática de sequence_no
		var maxSeq int
		err := tx.QueryRow(`SELECT COALESCE(MAX(sequence_no), 0) + 1 FROM messages WHERE chat_id = ?;`, msg.ChatID).Scan(&maxSeq)
		if err != nil {
			return fmt.Errorf("storage: falha ao calcular próximo sequence_no: %w", err)
		}
		msg.SequenceNo = maxSeq
	}

	query := `
		INSERT INTO messages (id, chat_id, author_type, author_id, role, content, content_type, status, sequence_no, metadata_json)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?);
	`
	_, err = tx.Exec(query, msg.ID, msg.ChatID, msg.AuthorType, msg.AuthorID, msg.Role, msg.Content, msg.ContentType, msg.Status, msg.SequenceNo, msg.MetadataJSON)
	if err != nil {
		return mapMessageError(err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("storage: falha ao confirmar transação da mensagem: %w", err)
	}

	return nil
}

// FindByID busca uma mensagem ativa ou editada pelo seu ID.
// Retorna ErrNotFound se a mensagem não existir ou estiver com Soft Delete (status = MessageDeleted).
func (r *MessageRepository) FindByID(id string) (*Message, error) {
	if strings.TrimSpace(id) == "" {
		return nil, ErrInvalidArgument
	}

	query := `
		SELECT id, chat_id, author_type, author_id, role, content, content_type, status, sequence_no, metadata_json, created_at, updated_at
		FROM messages
		WHERE id = ? AND status != 3;
	`
	row := r.db.QueryRow(query, id)

	var msg Message
	err := row.Scan(
		&msg.ID, &msg.ChatID, &msg.AuthorType, &msg.AuthorID, &msg.Role,
		&msg.Content, &msg.ContentType, &msg.Status, &msg.SequenceNo,
		&msg.MetadataJSON, &msg.CreatedAt, &msg.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("storage: falha ao buscar mensagem %s: %w", id, err)
	}

	return &msg, nil
}

// ListByChat retorna a lista de mensagens de um chat com sequence_no maior que sinceSequenceNo.
// Suporta limitação de quantidade via limit (padrão 100, cap máximo de 500). Ordena por sequence_no ASC.
func (r *MessageRepository) ListByChat(chatID string, sinceSequenceNo int, limit int) ([]*Message, error) {
	if strings.TrimSpace(chatID) == "" {
		return nil, ErrInvalidArgument
	}

	if limit <= 0 {
		limit = 100
	} else if limit > 500 {
		limit = 500
	}

	query := `
		SELECT id, chat_id, author_type, author_id, role, content, content_type, status, sequence_no, metadata_json, created_at, updated_at
		FROM messages
		WHERE chat_id = ? AND sequence_no > ? AND status != 3
		ORDER BY sequence_no ASC
		LIMIT ?;
	`
	rows, err := r.db.Query(query, chatID, sinceSequenceNo, limit)
	if err != nil {
		return nil, fmt.Errorf("storage: falha ao listar mensagens do chat %s: %w", chatID, err)
	}
	defer rows.Close()

	msgs := []*Message{}
	for rows.Next() {
		var msg Message
		err := rows.Scan(
			&msg.ID, &msg.ChatID, &msg.AuthorType, &msg.AuthorID, &msg.Role,
			&msg.Content, &msg.ContentType, &msg.Status, &msg.SequenceNo,
			&msg.MetadataJSON, &msg.CreatedAt, &msg.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("storage: falha ao ler linha de mensagem: %w", err)
		}
		msgs = append(msgs, &msg)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("storage: erro durante iteração de mensagens: %w", err)
	}

	return msgs, nil
}

// Update atualiza content, content_type e metadata_json de uma mensagem existente.
// Transiciona o status de MessageSent (1) para MessageEdited (2).
// Retorna ErrNotFound se a mensagem não existir ou já estiver soft-deleted.
func (r *MessageRepository) Update(msg Message) error {
	if strings.TrimSpace(msg.ID) == "" {
		return ErrInvalidArgument
	}

	if msg.Status == MessageDeleted {
		return ErrInvalidArgument
	}

	now := time.Now().UTC().Format(time.RFC3339)
	query := `
		UPDATE messages
		SET content = ?, content_type = ?, metadata_json = ?, status = 2, updated_at = ?
		WHERE id = ? AND status != 3;
	`
	res, err := r.db.Exec(query, msg.Content, msg.ContentType, msg.MetadataJSON, now, msg.ID)
	if err != nil {
		return mapMessageError(err)
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("storage: falha ao verificar linhas afetadas no update da mensagem: %w", err)
	}

	if rows == 0 {
		return ErrNotFound
	}

	return nil
}

// Delete realiza o Soft Delete de uma mensagem configurando status = MessageDeleted (3) e atualizando updated_at.
// É uma operação 100% idempotente: retorna nil se a mensagem já estiver deletada (status = 3) ou não existir.
func (r *MessageRepository) Delete(id string) error {
	if strings.TrimSpace(id) == "" {
		return ErrInvalidArgument
	}

	now := time.Now().UTC().Format(time.RFC3339)
	query := `UPDATE messages SET status = 3, updated_at = ? WHERE id = ? AND status != 3;`
	_, err := r.db.Exec(query, now, id)
	if err != nil {
		return mapMessageError(err)
	}

	return nil
}

// mapMessageError converte erros técnicos de banco em erros semânticos do recurso Message.
func mapMessageError(err error) error {
	techErr := translateSQLiteError(err)
	switch {
	case errors.Is(techErr, ErrUniqueViolation):
		return ErrAlreadyExists
	case errors.Is(techErr, ErrForeignKeyViolation):
		return ErrInvalidChat
	default:
		return techErr
	}
}
