package service

import (
	"context"
	"errors"
	"strings"

	"github.com/GkIgor/jay-ia/core/internal/permission"
	"github.com/GkIgor/jay-ia/core/internal/storage"
	"github.com/GkIgor/jay-ia/sdk/ipc"
)

// RegistrationStore define a interface do repositório de registros consumida pela camada de serviço.
type RegistrationStore interface {
	Upsert(reg storage.Registration) error
	FindByID(id string) (*storage.Registration, error)
	List() ([]*storage.Registration, error)
	Delete(id string) error
}

// SharedRuleStore define a interface do repositório de regras de compartilhamento consumida pela camada de serviço.
type SharedRuleStore interface {
	ReplaceRules(registrationID string, rules []storage.SharedRule) (int, error)
	ListByRegistration(registrationID string) ([]*storage.SharedRule, error)
}

// RegistrationService encapsula os casos de uso de domínio e regras de autorização do módulo de Registros.
type RegistrationService struct {
	regRepo   RegistrationStore
	ruleRepo  SharedRuleStore
	evaluator *permission.Evaluator
}

// NewRegistrationService cria uma nova instância do serviço de registros.
func NewRegistrationService(
	regRepo RegistrationStore,
	ruleRepo SharedRuleStore,
	evaluator *permission.Evaluator,
) (*RegistrationService, error) {
	if regRepo == nil || ruleRepo == nil || evaluator == nil {
		return nil, errors.New("registration_service: dependências nulas não são permitidas")
	}
	return &RegistrationService{
		regRepo:   regRepo,
		ruleRepo:  ruleRepo,
		evaluator: evaluator,
	}, nil
}

// RegisterClient realiza o auto-registro ou re-registro idempotente de um consumidor.
func (s *RegistrationService) RegisterClient(ctx context.Context, clientID string, metadataJSON string) (*storage.Registration, error) {
	cleanID := strings.TrimSpace(clientID)
	if cleanID == "" {
		return nil, storage.ErrInvalidArgument
	}

	reg := storage.Registration{
		ID:           cleanID,
		MetadataJSON: metadataJSON,
		Status:       storage.RegistrationActive,
	}

	if err := s.regRepo.Upsert(reg); err != nil {
		return nil, err
	}

	return s.regRepo.FindByID(cleanID)
}

// UnregisterClient remove um registro existente. Apenas o proprietário ou um admin pode descadastrar.
func (s *RegistrationService) UnregisterClient(ctx context.Context, requesterID string, targetID string) error {
	cleanReq := strings.TrimSpace(requesterID)
	cleanTarget := strings.TrimSpace(targetID)
	if cleanReq == "" || cleanTarget == "" {
		return storage.ErrInvalidArgument
	}

	if cleanReq != cleanTarget {
		rules, err := s.ruleRepo.ListByRegistration(cleanTarget)
		if err != nil && !errors.Is(err, storage.ErrNotFound) {
			return err
		}
		allowed, err := s.evaluator.Evaluate(rules, permission.AccessRequest{
			RequesterID:     cleanReq,
			ResourceOwnerID: cleanTarget,
			TargetScope:     storage.ScopeAll,
			ResourceID:      cleanTarget,
			RequestedAction: storage.ActionAdmin,
		})
		if err != nil || !allowed {
			return storage.ErrDeleteRestricted
		}
	}

	return s.regRepo.Delete(cleanTarget)
}

// UpdateRegistration atualiza o status ou metadados de um registro existente.
func (s *RegistrationService) UpdateRegistration(ctx context.Context, requesterID string, targetID string, status storage.RegistrationStatus, metadataJSON string) (*storage.Registration, error) {
	cleanReq := strings.TrimSpace(requesterID)
	cleanTarget := strings.TrimSpace(targetID)
	if cleanReq == "" || cleanTarget == "" {
		return nil, storage.ErrInvalidArgument
	}

	if cleanReq != cleanTarget {
		rules, err := s.ruleRepo.ListByRegistration(cleanTarget)
		if err != nil && !errors.Is(err, storage.ErrNotFound) {
			return nil, err
		}
		allowed, err := s.evaluator.Evaluate(rules, permission.AccessRequest{
			RequesterID:     cleanReq,
			ResourceOwnerID: cleanTarget,
			TargetScope:     storage.ScopeAll,
			ResourceID:      cleanTarget,
			RequestedAction: storage.ActionAdmin,
		})
		if err != nil || !allowed {
			return nil, storage.ErrDeleteRestricted
		}
	}

	reg := storage.Registration{
		ID:           cleanTarget,
		Status:       status,
		MetadataJSON: metadataJSON,
	}

	if err := s.regRepo.Upsert(reg); err != nil {
		return nil, err
	}

	return s.regRepo.FindByID(cleanTarget)
}

// GetRegistration busca um registro por ID. Retorna ErrNotFound se o requisitante não possuir acesso (ocultação de segurança).
func (s *RegistrationService) GetRegistration(ctx context.Context, requesterID string, targetID string) (*storage.Registration, error) {
	cleanReq := strings.TrimSpace(requesterID)
	cleanTarget := strings.TrimSpace(targetID)
	if cleanReq == "" || cleanTarget == "" {
		return nil, storage.ErrInvalidArgument
	}

	if cleanReq != cleanTarget {
		rules, err := s.ruleRepo.ListByRegistration(cleanTarget)
		if err != nil && !errors.Is(err, storage.ErrNotFound) {
			return nil, storage.ErrNotFound
		}
		allowed, err := s.evaluator.Evaluate(rules, permission.AccessRequest{
			RequesterID:     cleanReq,
			ResourceOwnerID: cleanTarget,
			TargetScope:     storage.ScopeAll,
			ResourceID:      cleanTarget,
			RequestedAction: storage.ActionRead,
		})
		if err != nil || !allowed {
			// Decisão de Segurança: Oculta a existência do recurso retornando ErrNotFound em vez de ErrForbidden
			return nil, storage.ErrNotFound
		}
	}

	return s.regRepo.FindByID(cleanTarget)
}

// ListRegistrations retorna a lista de todos os registros ativos conhecidos pelo Core.
func (s *RegistrationService) ListRegistrations(ctx context.Context, requesterID string) ([]*storage.Registration, error) {
	if strings.TrimSpace(requesterID) == "" {
		return nil, storage.ErrInvalidArgument
	}
	return s.regRepo.List()
}

// UpdateSharedRules substitui atomicamente as regras de compartilhamento pertencentes ao consumidor declarante.
func (s *RegistrationService) UpdateSharedRules(ctx context.Context, requesterID string, rulesPayload []ipc.SharedRulePayload) (int, error) {
	cleanReq := strings.TrimSpace(requesterID)
	if cleanReq == "" {
		return 0, storage.ErrInvalidArgument
	}

	// Garante que o registro do proprietário existe
	_, err := s.regRepo.FindByID(cleanReq)
	if err != nil {
		return 0, err
	}

	domainRules := make([]storage.SharedRule, 0, len(rulesPayload))
	for _, rp := range rulesPayload {
		cleanPattern := strings.TrimSpace(rp.Pattern)
		if cleanPattern == "" {
			return 0, permission.ErrInvalidArgument
		}
		domainRules = append(domainRules, storage.SharedRule{
			RegistrationID: cleanReq,
			TargetScope:    storage.RuleScope(rp.TargetScope),
			Pattern:        cleanPattern,
			MatchType:      storage.MatchType(rp.MatchType),
			AllowedActions: storage.PermissionAction(rp.AllowedActions),
		})
	}

	return s.ruleRepo.ReplaceRules(cleanReq, domainRules)
}
