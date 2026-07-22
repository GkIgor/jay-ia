package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/GkIgor/jay-ia/core/internal/permission"
	"github.com/GkIgor/jay-ia/core/internal/storage"
)

// ChatStore define a interface do repositório de chats consumida pela camada de serviço.
type ChatStore interface {
	Create(chat storage.Chat) error
	FindByID(id string) (*storage.Chat, error)
	ListByOwner(ownerRegistrationID string, filter storage.ChatFilter) ([]*storage.Chat, error)
	Update(chat storage.Chat) error
	Delete(id string) error
}

// ChatService encapsula os casos de uso de domínio e regras de autorização do módulo de Chats.
type ChatService struct {
	chatRepo  ChatStore
	regRepo   RegistrationStore
	ruleRepo  SharedRuleStore
	evaluator *permission.Evaluator
}

// NewChatService cria uma nova instância de ChatService.
func NewChatService(
	chatRepo ChatStore,
	regRepo RegistrationStore,
	ruleRepo SharedRuleStore,
	evaluator *permission.Evaluator,
) (*ChatService, error) {
	if chatRepo == nil || regRepo == nil || ruleRepo == nil || evaluator == nil {
		return nil, errors.New("chat_service: dependências nulas não são permitidas")
	}
	return &ChatService{
		chatRepo:  chatRepo,
		regRepo:   regRepo,
		ruleRepo:  ruleRepo,
		evaluator: evaluator,
	}, nil
}

// CreateChat cria um novo chat para o proprietário especificado com UUID v4 gerado.
func (s *ChatService) CreateChat(ctx context.Context, ownerRegistrationID string, title string, metadataJSON string) (*storage.Chat, error) {
	cleanOwner := strings.TrimSpace(ownerRegistrationID)
	cleanTitle := strings.TrimSpace(title)
	if cleanOwner == "" {
		return nil, storage.ErrInvalidArgument
	}
	if cleanTitle == "" {
		cleanTitle = "Novo Chat"
	}
	if metadataJSON == "" {
		metadataJSON = "{}"
	}

	chatID := generateUUIDv4()
	chat := storage.Chat{
		ID:                  chatID,
		OwnerRegistrationID: cleanOwner,
		Title:               cleanTitle,
		Status:              storage.ChatActive,
		MetadataJSON:        metadataJSON,
	}

	if err := s.chatRepo.Create(chat); err != nil {
		return nil, err
	}

	return s.chatRepo.FindByID(chatID)
}

// DeleteChat executa o Soft Delete em um chat. Apenas o proprietário ou autorizado com ActionAdmin pode deletar.
func (s *ChatService) DeleteChat(ctx context.Context, requesterID string, chatID string) error {
	cleanReq := strings.TrimSpace(requesterID)
	cleanID := strings.TrimSpace(chatID)
	if cleanReq == "" || cleanID == "" {
		return storage.ErrInvalidArgument
	}

	chat, err := s.chatRepo.FindByID(cleanID)
	if err != nil {
		return storage.ErrNotFound
	}
	if chat.Status == storage.ChatDeleted {
		return storage.ErrNotFound
	}

	if cleanReq != chat.OwnerRegistrationID {
		rules, err := s.ruleRepo.ListByRegistration(chat.OwnerRegistrationID)
		if err != nil && !errors.Is(err, storage.ErrNotFound) {
			return err
		}
		allowed, err := s.evaluator.Evaluate(rules, permission.AccessRequest{
			RequesterID:     cleanReq,
			ResourceOwnerID: chat.OwnerRegistrationID,
			TargetScope:     storage.ScopeChats,
			ResourceID:      cleanID,
			RequestedAction: storage.ActionAdmin,
		})
		if err != nil || !allowed {
			return storage.ErrForbidden
		}
	}

	return s.chatRepo.Delete(cleanID)
}

// RenameChat atualiza o título de um chat existente. Requer autorização ActionWrite se não for o proprietário.
func (s *ChatService) RenameChat(ctx context.Context, requesterID string, chatID string, newTitle string) (*storage.Chat, error) {
	cleanReq := strings.TrimSpace(requesterID)
	cleanID := strings.TrimSpace(chatID)
	cleanTitle := strings.TrimSpace(newTitle)
	if cleanReq == "" || cleanID == "" || cleanTitle == "" {
		return nil, storage.ErrInvalidArgument
	}

	chat, err := s.chatRepo.FindByID(cleanID)
	if err != nil {
		return nil, storage.ErrNotFound
	}
	if chat.Status == storage.ChatDeleted {
		return nil, storage.ErrNotFound
	}

	if cleanReq != chat.OwnerRegistrationID {
		rules, err := s.ruleRepo.ListByRegistration(chat.OwnerRegistrationID)
		if err != nil && !errors.Is(err, storage.ErrNotFound) {
			return nil, err
		}
		allowed, err := s.evaluator.Evaluate(rules, permission.AccessRequest{
			RequesterID:     cleanReq,
			ResourceOwnerID: chat.OwnerRegistrationID,
			TargetScope:     storage.ScopeChats,
			ResourceID:      cleanID,
			RequestedAction: storage.ActionWrite,
		})
		if err != nil || !allowed {
			return nil, storage.ErrForbidden
		}
	}

	chat.Title = cleanTitle
	if err := s.chatRepo.Update(*chat); err != nil {
		return nil, err
	}

	return s.chatRepo.FindByID(cleanID)
}

// GetChat busca os detalhes de um chat por ID. Retorna ErrNotFound em caso de não autorização (Ocultação de Segurança).
func (s *ChatService) GetChat(ctx context.Context, requesterID string, chatID string) (*storage.Chat, error) {
	cleanReq := strings.TrimSpace(requesterID)
	cleanID := strings.TrimSpace(chatID)
	if cleanReq == "" || cleanID == "" {
		return nil, storage.ErrInvalidArgument
	}

	chat, err := s.chatRepo.FindByID(cleanID)
	if err != nil {
		return nil, storage.ErrNotFound
	}
	if chat.Status == storage.ChatDeleted {
		return nil, storage.ErrNotFound
	}

	if cleanReq != chat.OwnerRegistrationID {
		rules, err := s.ruleRepo.ListByRegistration(chat.OwnerRegistrationID)
		if err != nil && !errors.Is(err, storage.ErrNotFound) {
			return nil, storage.ErrNotFound
		}
		allowed, err := s.evaluator.Evaluate(rules, permission.AccessRequest{
			RequesterID:     cleanReq,
			ResourceOwnerID: chat.OwnerRegistrationID,
			TargetScope:     storage.ScopeChats,
			ResourceID:      cleanID,
			RequestedAction: storage.ActionRead,
		})
		if err != nil || !allowed {
			// Ocultação de Segurança: retorna ErrNotFound para não vazamento de recurso
			return nil, storage.ErrNotFound
		}
	}

	return chat, nil
}

// ListChats retorna os chats pertencentes ao requisitante e opcionalmente chats compartilhados autorizados.
// Resultados são ordenados por created_at DESC.
func (s *ChatService) ListChats(ctx context.Context, requesterID string, includeShared bool, limit int) ([]*storage.Chat, error) {
	cleanReq := strings.TrimSpace(requesterID)
	if cleanReq == "" {
		return nil, storage.ErrInvalidArgument
	}
	if limit <= 0 {
		limit = 50
	}

	// 1. Busca chats próprios ativos
	ownChats, err := s.chatRepo.ListByOwner(cleanReq, storage.ChatFilterActiveOnly)
	if err != nil {
		return nil, err
	}

	chatsMap := make(map[string]*storage.Chat)
	for _, c := range ownChats {
		if c.Status != storage.ChatDeleted {
			chatsMap[c.ID] = c
		}
	}

	// 2. Se includeShared == true, busca chats de outros proprietários com permissão ActionRead
	if includeShared {
		allRegs, err := s.regRepo.List()
		if err == nil {
			for _, reg := range allRegs {
				if reg.ID == cleanReq {
					continue
				}
				rules, err := s.ruleRepo.ListByRegistration(reg.ID)
				if err != nil || len(rules) == 0 {
					continue
				}

				otherChats, err := s.chatRepo.ListByOwner(reg.ID, storage.ChatFilterActiveOnly)
				if err != nil {
					continue
				}

				for _, c := range otherChats {
					if c.Status == storage.ChatDeleted {
						continue
					}
					allowed, err := s.evaluator.Evaluate(rules, permission.AccessRequest{
						RequesterID:     cleanReq,
						ResourceOwnerID: reg.ID,
						TargetScope:     storage.ScopeChats,
						ResourceID:      c.ID,
						RequestedAction: storage.ActionRead,
					})
					if err == nil && allowed {
						chatsMap[c.ID] = c
					}
				}
			}
		}
	}

	// 3. Converte o mapa consolidado em slice
	result := make([]*storage.Chat, 0, len(chatsMap))
	for _, c := range chatsMap {
		result = append(result, c)
	}

	// 4. Ordena por created_at DESC (mais recente primeiro)
	sort.Slice(result, func(i, j int) bool {
		return result[i].CreatedAt > result[j].CreatedAt
	})

	// 5. Trunca para o limite solicitado
	if len(result) > limit {
		result = result[:limit]
	}

	return result, nil
}

// generateUUIDv4 gera um identificador no formato UUID v4.
func generateUUIDv4() string {
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
