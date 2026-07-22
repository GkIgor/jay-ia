package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"

	"github.com/GkIgor/jay-ia/core/internal/permission"
	"github.com/GkIgor/jay-ia/core/internal/storage"
	"github.com/GkIgor/jay-ia/sdk/ipc"
)

// MessageStore define a interface do repositório de mensagens consumida pela camada de serviço.
type MessageStore interface {
	Create(msg storage.Message) error
	FindByID(id string) (*storage.Message, error)
	ListByChat(chatID string, sinceSequenceNo int, limit int) ([]*storage.Message, error)
	Update(msg storage.Message) error
	Delete(id string) error
}

// MessageService encapsula os casos de uso do ciclo de vida de mensagens e regras de autorização.
type MessageService struct {
	msgRepo   MessageStore
	chatRepo  ChatStore
	ruleRepo  SharedRuleStore
	evaluator *permission.Evaluator
}

// NewMessageService cria uma nova instância de MessageService.
func NewMessageService(
	msgRepo MessageStore,
	chatRepo ChatStore,
	ruleRepo SharedRuleStore,
	evaluator *permission.Evaluator,
) (*MessageService, error) {
	if msgRepo == nil || chatRepo == nil || ruleRepo == nil || evaluator == nil {
		return nil, errors.New("message_service: dependências nulas não são permitidas")
	}
	return &MessageService{
		msgRepo:   msgRepo,
		chatRepo:  chatRepo,
		ruleRepo:  ruleRepo,
		evaluator: evaluator,
	}, nil
}

// CreateMessage insere uma mensagem no histórico de um Chat com incremento atômico de sequence_no.
func (s *MessageService) CreateMessage(ctx context.Context, requesterID string, req ipc.CreateMessageRequest) (*storage.Message, error) {
	cleanReq := strings.TrimSpace(requesterID)
	cleanChatID := strings.TrimSpace(req.ChatID)
	cleanContent := strings.TrimSpace(req.Content)
	cleanAuthorID := strings.TrimSpace(req.AuthorID)

	if cleanReq == "" || cleanChatID == "" || cleanContent == "" {
		return nil, storage.ErrInvalidArgument
	}
	if cleanAuthorID == "" {
		cleanAuthorID = cleanReq
	}

	// 1. Autorização: Requer ActionWrite no chat
	if err := s.authorizeChatAccess(ctx, cleanReq, cleanChatID, storage.ScopeMessages, storage.ActionWrite); err != nil {
		return nil, err
	}

	msgID := generateMessageUUIDv4()
	msg := storage.Message{
		ID:           msgID,
		ChatID:       cleanChatID,
		AuthorType:   storage.AuthorType(req.AuthorType),
		AuthorID:     cleanAuthorID,
		Role:         storage.MessageRole(req.Role),
		Content:      cleanContent,
		ContentType:  storage.MessageContentType(req.ContentType),
		Status:       storage.MessageSent,
		MetadataJSON: req.Metadata,
	}

	if msg.AuthorType == 0 {
		msg.AuthorType = storage.AuthorRegistration
	}
	if msg.Role == 0 {
		msg.Role = storage.RoleUser
	}
	if msg.ContentType == 0 {
		msg.ContentType = storage.ContentTypeTextPlain
	}
	if msg.MetadataJSON == "" {
		msg.MetadataJSON = "{}"
	}

	if err := s.msgRepo.Create(msg); err != nil {
		return nil, err
	}

	return s.msgRepo.FindByID(msgID)
}

// UpdateMessage edita o conteúdo de uma mensagem existente. Requer autoria ou ActionWrite no chat.
func (s *MessageService) UpdateMessage(ctx context.Context, requesterID string, messageID string, newContent string, metadataJSON string) (*storage.Message, error) {
	cleanReq := strings.TrimSpace(requesterID)
	cleanID := strings.TrimSpace(messageID)
	cleanContent := strings.TrimSpace(newContent)
	if cleanReq == "" || cleanID == "" || cleanContent == "" {
		return nil, storage.ErrInvalidArgument
	}

	msg, err := s.msgRepo.FindByID(cleanID)
	if err != nil {
		return nil, storage.ErrNotFound
	}
	if msg.Status == storage.MessageDeleted {
		return nil, storage.ErrNotFound
	}

	// Autorização: Autor da mensagem ou ActionWrite no chat
	if cleanReq != msg.AuthorID {
		if err := s.authorizeChatAccess(ctx, cleanReq, msg.ChatID, storage.ScopeMessages, storage.ActionWrite); err != nil {
			return nil, err
		}
	}

	msg.Content = cleanContent
	if metadataJSON != "" {
		msg.MetadataJSON = metadataJSON
	}
	msg.Status = storage.MessageEdited

	if err := s.msgRepo.Update(*msg); err != nil {
		return nil, err
	}

	return s.msgRepo.FindByID(cleanID)
}

// DeleteMessage realiza o Soft Delete de uma mensagem (status = MessageDeleted). Requer autoria ou ActionAdmin no chat.
func (s *MessageService) DeleteMessage(ctx context.Context, requesterID string, messageID string) error {
	cleanReq := strings.TrimSpace(requesterID)
	cleanID := strings.TrimSpace(messageID)
	if cleanReq == "" || cleanID == "" {
		return storage.ErrInvalidArgument
	}

	msg, err := s.msgRepo.FindByID(cleanID)
	if err != nil {
		return storage.ErrNotFound
	}
	if msg.Status == storage.MessageDeleted {
		return storage.ErrNotFound
	}

	// Autorização: Autor da mensagem ou ActionAdmin no chat (preserva a auditabilidade)
	if cleanReq != msg.AuthorID {
		if err := s.authorizeChatAccess(ctx, cleanReq, msg.ChatID, storage.ScopeMessages, storage.ActionAdmin); err != nil {
			return err
		}
	}

	return s.msgRepo.Delete(cleanID)
}

// GetMessages retorna o histórico de mensagens de um chat no modelo Pull (sequence_no > sinceSequenceNo).
func (s *MessageService) GetMessages(ctx context.Context, requesterID string, chatID string, sinceSequenceNo int, limit int) ([]*storage.Message, bool, error) {
	cleanReq := strings.TrimSpace(requesterID)
	cleanChatID := strings.TrimSpace(chatID)
	if cleanReq == "" || cleanChatID == "" {
		return nil, false, storage.ErrInvalidArgument
	}
	if limit <= 0 {
		limit = 50
	} else if limit > 100 {
		limit = 100
	}

	// Autorização: Requer ActionRead no chat
	if err := s.authorizeChatAccess(ctx, cleanReq, cleanChatID, storage.ScopeMessages, storage.ActionRead); err != nil {
		return nil, false, err
	}

	// Busca limit + 1 para calcular has_more
	records, err := s.msgRepo.ListByChat(cleanChatID, sinceSequenceNo, limit+1)
	if err != nil {
		return nil, false, err
	}

	hasMore := false
	if len(records) > limit {
		hasMore = true
		records = records[:limit]
	}

	return records, hasMore, nil
}

// authorizeChatAccess é um helper interno que centraliza a verificação de acesso ao Chat pai.
func (s *MessageService) authorizeChatAccess(ctx context.Context, requesterID string, chatID string, scope storage.RuleScope, action storage.PermissionAction) error {
	chat, err := s.chatRepo.FindByID(chatID)
	if err != nil {
		return storage.ErrNotFound
	}
	if chat.Status == storage.ChatDeleted {
		return storage.ErrNotFound
	}

	if requesterID == chat.OwnerRegistrationID {
		return nil
	}

	rules, err := s.ruleRepo.ListByRegistration(chat.OwnerRegistrationID)
	if err != nil && !errors.Is(err, storage.ErrNotFound) {
		return storage.ErrForbidden
	}

	allowed, err := s.evaluator.Evaluate(rules, permission.AccessRequest{
		RequesterID:     requesterID,
		ResourceOwnerID: chat.OwnerRegistrationID,
		TargetScope:     scope,
		ResourceID:      chatID,
		RequestedAction: action,
	})
	if err != nil || !allowed {
		if action == storage.ActionRead {
			return storage.ErrNotFound // Ocultação de segurança para leitura
		}
		return storage.ErrForbidden
	}

	return nil
}

// generateMessageUUIDv4 gera um UUID v4 para identificação única de mensagens.
func generateMessageUUIDv4() string {
	var uuid [16]byte
	_, err := rand.Read(uuid[:])
	if err != nil {
		return fmt.Sprintf("%x-%x-%x-%x-%x", uuid[0:4], uuid[4:6], uuid[6:8], uuid[8:10], uuid[10:16])
	}
	uuid[6] = (uuid[6] & 0x0f) | 0x40
	uuid[8] = (uuid[8] & 0x3f) | 0x80
	buf := make([]byte, 36)
	hex.Encode(buf[0:8], uuid[0:4])
	buf[8] = '-'
	hex.Encode(buf[9:13], uuid[4:6])
	buf[13] = '-'
	hex.Encode(buf[14:18], uuid[6:8])
	buf[18] = '-'
	hex.Encode(buf[19:23], uuid[8:10])
	buf[23] = '-'
	hex.Encode(buf[24:36], uuid[10:16])
	return string(buf)
}
