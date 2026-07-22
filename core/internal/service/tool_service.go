package service

import (
	"context"
	"errors"
	"strings"

	"github.com/GkIgor/jay-ia/core/internal/permission"
	"github.com/GkIgor/jay-ia/core/internal/storage"
	"github.com/GkIgor/jay-ia/sdk/ipc"
)

// ToolStore define a interface do repositório de ferramentas consumida pela camada de serviço.
type ToolStore interface {
	Register(tool storage.Tool) error
	FindByID(id string) (*storage.Tool, error)
	ListByRegistration(registrationID string) ([]*storage.Tool, error)
	ListAvailable() ([]*storage.Tool, error)
	Delete(id string) error
}

// ToolService encapsula os casos de uso do catálogo de ferramentas e proteção contra sequestro de identidade (Hijack Prevention).
type ToolService struct {
	toolRepo  ToolStore
	ruleRepo  SharedRuleStore
	evaluator *permission.Evaluator
}

// NewToolService cria uma nova instância de ToolService.
func NewToolService(
	toolRepo ToolStore,
	ruleRepo SharedRuleStore,
	evaluator *permission.Evaluator,
) (*ToolService, error) {
	if toolRepo == nil || ruleRepo == nil || evaluator == nil {
		return nil, errors.New("tool_service: dependências nulas não são permitidas")
	}
	return &ToolService{
		toolRepo:  toolRepo,
		ruleRepo:  ruleRepo,
		evaluator: evaluator,
	}, nil
}

// RegisterTool cadastra ou atualiza (Upsert idempotente) uma ferramenta no catálogo com Hijack Prevention.
func (s *ToolService) RegisterTool(ctx context.Context, ownerRegistrationID string, req ipc.RegisterToolRequest) (*storage.Tool, error) {
	cleanOwner := strings.TrimSpace(ownerRegistrationID)
	cleanID := strings.TrimSpace(req.ID)
	cleanName := strings.TrimSpace(req.Name)

	if cleanOwner == "" || cleanID == "" || cleanName == "" {
		return nil, storage.ErrInvalidArgument
	}

	version := strings.TrimSpace(req.Version)
	if version == "" {
		version = "1.0.0"
	}
	schemaJSON := strings.TrimSpace(req.SchemaJSON)
	if schemaJSON == "" {
		schemaJSON = "{}"
	}

	tool := storage.Tool{
		ID:             cleanID,
		RegistrationID: cleanOwner,
		Name:           cleanName,
		Description:    req.Description,
		Version:        version,
		SchemaJSON:     schemaJSON,
		Status:         storage.ToolAvailable,
	}

	if err := s.toolRepo.Register(tool); err != nil {
		return nil, err
	}

	return s.toolRepo.FindByID(cleanID)
}

// UnregisterTool realiza a remoção física (Hard Delete) de uma ferramenta do catálogo.
func (s *ToolService) UnregisterTool(ctx context.Context, requesterID string, toolID string) error {
	cleanReq := strings.TrimSpace(requesterID)
	cleanID := strings.TrimSpace(toolID)
	if cleanReq == "" || cleanID == "" {
		return storage.ErrInvalidArgument
	}

	tool, err := s.toolRepo.FindByID(cleanID)
	if err != nil {
		return storage.ErrNotFound
	}

	// Autorização: Proprietário da ferramenta ou privilégio ActionAdmin
	if cleanReq != tool.RegistrationID {
		rules, err := s.ruleRepo.ListByRegistration(tool.RegistrationID)
		if err != nil && !errors.Is(err, storage.ErrNotFound) {
			return err
		}
		allowed, err := s.evaluator.Evaluate(rules, permission.AccessRequest{
			RequesterID:     cleanReq,
			ResourceOwnerID: tool.RegistrationID,
			TargetScope:     storage.ScopeTools,
			ResourceID:      cleanID,
			RequestedAction: storage.ActionAdmin,
		})
		if err != nil || !allowed {
			return storage.ErrForbidden
		}
	}

	return s.toolRepo.Delete(cleanID)
}

// GetTool busca o detalhe de uma ferramenta por ID. Retorna ErrNotFound em caso de não autorização (Ocultação de Segurança).
func (s *ToolService) GetTool(ctx context.Context, requesterID string, toolID string) (*storage.Tool, error) {
	cleanReq := strings.TrimSpace(requesterID)
	cleanID := strings.TrimSpace(toolID)
	if cleanReq == "" || cleanID == "" {
		return nil, storage.ErrInvalidArgument
	}

	tool, err := s.toolRepo.FindByID(cleanID)
	if err != nil {
		return nil, storage.ErrNotFound
	}

	if cleanReq != tool.RegistrationID {
		rules, err := s.ruleRepo.ListByRegistration(tool.RegistrationID)
		if err != nil && !errors.Is(err, storage.ErrNotFound) {
			return nil, storage.ErrNotFound
		}
		allowed, err := s.evaluator.Evaluate(rules, permission.AccessRequest{
			RequesterID:     cleanReq,
			ResourceOwnerID: tool.RegistrationID,
			TargetScope:     storage.ScopeTools,
			ResourceID:      cleanID,
			RequestedAction: storage.ActionRead,
		})
		if err != nil || !allowed {
			return nil, storage.ErrNotFound // Ocultação de Segurança
		}
	}

	return tool, nil
}

// ListTools lista as ferramentas disponíveis ou filtradas por proprietário, omitindo ferramentas não autorizadas.
func (s *ToolService) ListTools(ctx context.Context, requesterID string, filterRegistrationID string) ([]*storage.Tool, error) {
	cleanReq := strings.TrimSpace(requesterID)
	cleanFilter := strings.TrimSpace(filterRegistrationID)
	if cleanReq == "" {
		return nil, storage.ErrInvalidArgument
	}

	var candidateTools []*storage.Tool
	var err error

	if cleanFilter != "" {
		candidateTools, err = s.toolRepo.ListByRegistration(cleanFilter)
	} else {
		candidateTools, err = s.toolRepo.ListAvailable()
	}

	if err != nil {
		return nil, err
	}

	allowedTools := make([]*storage.Tool, 0, len(candidateTools))
	for _, tool := range candidateTools {
		if tool.RegistrationID == cleanReq {
			allowedTools = append(allowedTools, tool)
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
			allowedTools = append(allowedTools, tool)
		}
	}

	return allowedTools, nil
}
