package service

import (
	"context"
	"errors"
	"strings"
	"sync"
	"time"

	"github.com/GkIgor/jay-ia/core/internal/llm"
	"github.com/GkIgor/jay-ia/core/internal/permission"
	"github.com/GkIgor/jay-ia/core/internal/storage"
	"github.com/GkIgor/jay-ia/core/internal/tools"
)

const (
	DefaultAgentID      = "jay_agent"
	DefaultHistoryLimit = 100
	DefaultLLMTimeout   = 30 * time.Second
)

// ProcessorService orquestra o ciclo de inferência de IA (LLM) sobre o histórico de conversas dos chats.
type ProcessorService struct {
	msgRepo   MessageStore
	chatRepo  ChatStore
	toolRepo  ToolStore
	ruleRepo  SharedRuleStore
	evaluator *permission.Evaluator
	llmClient llm.Client
	chatLocks sync.Map
}

// NewProcessorService instancia o serviço de processamento de chats com a IA.
func NewProcessorService(
	msgRepo MessageStore,
	chatRepo ChatStore,
	toolRepo ToolStore,
	ruleRepo SharedRuleStore,
	evaluator *permission.Evaluator,
	llmClient llm.Client,
) (*ProcessorService, error) {
	if msgRepo == nil || chatRepo == nil || toolRepo == nil || ruleRepo == nil || evaluator == nil || llmClient == nil {
		return nil, errors.New("processor_service: dependências nulas não são permitidas")
	}
	return &ProcessorService{
		msgRepo:   msgRepo,
		chatRepo:  chatRepo,
		toolRepo:  toolRepo,
		ruleRepo:  ruleRepo,
		evaluator: evaluator,
		llmClient: llmClient,
	}, nil
}

// ProcessChat executa um ciclo de inferência da LLM sobre um chat, persisto a resposta do assistente.
func (s *ProcessorService) ProcessChat(ctx context.Context, requesterID string, chatID string) (*storage.Message, error) {
	cleanReq := strings.TrimSpace(requesterID)
	cleanChatID := strings.TrimSpace(chatID)

	if cleanReq == "" || cleanChatID == "" {
		return nil, storage.ErrInvalidArgument
	}

	// 1. Autorização & Validação Preliminar (Fora do Lock)
	chat, err := s.chatRepo.FindByID(cleanChatID)
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
			TargetScope:     storage.ScopeMessages,
			ResourceID:      cleanChatID,
			RequestedAction: storage.ActionWrite,
		})
		if err != nil || !allowed {
			return nil, storage.ErrForbidden
		}
	}

	// 2. Adquisição do Mutex por Chat para serialização da inferência naquele chatID
	lockVal, _ := s.chatLocks.LoadOrStore(cleanChatID, &sync.Mutex{})
	chatMutex := lockVal.(*sync.Mutex)
	chatMutex.Lock()
	defer chatMutex.Unlock()

	// 3. Carrega Histórico Recente de Mensagens
	history, err := s.msgRepo.ListByChat(cleanChatID, 0, DefaultHistoryLimit)
	if err != nil {
		return nil, err
	}

	llmHistory := toLLMMessages(history)

	// 4. Consulta e Mapeamento de Ferramentas Autorizadas para o Requisitante
	candidateTools, err := s.toolRepo.ListAvailable()
	if err != nil {
		candidateTools = []*storage.Tool{}
	}

	authorizedTools := make([]*storage.Tool, 0, len(candidateTools))
	for _, tool := range candidateTools {
		if tool.RegistrationID == cleanReq {
			authorizedTools = append(authorizedTools, tool)
			continue
		}

		rules, err := s.ruleRepo.ListByRegistration(tool.RegistrationID)
		if err != nil || len(rules) == 0 {
			continue
		}

		allowed, err := s.evaluator.Evaluate(rules, permission.AccessRequest{
			RequesterID:     cleanReq,
			ResourceOwnerID: tool.RegistrationID,
			TargetScope:     storage.ScopeTools,
			ResourceID:      tool.ID,
			RequestedAction: storage.ActionRead,
		})
		if err == nil && allowed {
			authorizedTools = append(authorizedTools, tool)
		}
	}

	llmTools := toLLMTools(authorizedTools)

	// 5. Aplicação de Contexto com Timeout Padrão
	procCtx, cancel := context.WithTimeout(ctx, DefaultLLMTimeout)
	defer cancel()

	// 6. Invocação da LLM
	resp, err := s.llmClient.GenerateContent(procCtx, llmHistory, llmTools)
	if err != nil {
		return nil, err
	}

	if resp == nil || strings.TrimSpace(resp.Text) == "" {
		return nil, errors.New("processor_service: LLM não retornou conteúdo de texto")
	}

	// 7. Persistência da Resposta do Assistente
	msgID := generateMessageUUIDv4()
	agentMsg := storage.Message{
		ID:           msgID,
		ChatID:       cleanChatID,
		AuthorType:   storage.AuthorAgent,
		AuthorID:     DefaultAgentID,
		Role:         storage.RoleAssistant,
		Content:      strings.TrimSpace(resp.Text),
		ContentType:  storage.ContentTypeTextPlain,
		Status:       storage.MessageSent,
		MetadataJSON: "{}",
	}

	if err := s.msgRepo.Create(agentMsg); err != nil {
		// Ausência de transação distribuída: loga e propaga a falha de escrita local
		return nil, err
	}

	return s.msgRepo.FindByID(msgID)
}

// toLLMMessages converte o histórico do banco storage.Message em []llm.Message para a LLM.
func toLLMMessages(records []*storage.Message) []llm.Message {
	result := make([]llm.Message, 0, len(records))
	for _, rec := range records {
		if rec.Status == storage.MessageDeleted {
			continue
		}

		var role llm.Role
		if rec.AuthorType == storage.AuthorRegistration || rec.Role == storage.RoleUser {
			role = llm.RoleUser
		} else {
			role = llm.RoleModel
		}

		result = append(result, llm.Message{
			Role: role,
			Parts: []llm.Part{
				{Text: rec.Content},
			},
		})
	}
	return result
}

// toLLMTools converte entidades storage.Tool em []tools.Definition para o cliente LLM.
func toLLMTools(records []*storage.Tool) []tools.Definition {
	result := make([]tools.Definition, 0, len(records))
	for _, t := range records {
		result = append(result, tools.Definition{
			Name:        t.Name,
			Description: t.Description,
		})
	}
	return result
}
