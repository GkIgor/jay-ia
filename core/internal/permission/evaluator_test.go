package permission

import (
	"errors"
	"fmt"
	"sync"
	"testing"

	"github.com/GkIgor/jay-ia/core/internal/storage"
)

func TestEvaluator_OwnershipRule(t *testing.T) {
	eval := NewEvaluator()

	req := AccessRequest{
		RequesterID:     "user-1",
		ResourceOwnerID: "user-1", // Proprietário == Requisitante
		TargetScope:     storage.ScopeChats,
		ResourceID:      "chat-123",
		RequestedAction: storage.ActionAdmin,
	}

	// Mesmo sem regras cadastradas (slice nil), deve retornar true imediatamente
	allowed, err := eval.Evaluate(nil, req)
	if err != nil {
		t.Fatalf("falha ao avaliar permissão do proprietário: %v", err)
	}
	if !allowed {
		t.Fatalf("esperava permissão concedida (ALLOW) para o proprietário")
	}
}

func TestEvaluator_DefaultDeny(t *testing.T) {
	eval := NewEvaluator()

	req := AccessRequest{
		RequesterID:     "user-guest",
		ResourceOwnerID: "user-owner",
		TargetScope:     storage.ScopeChats,
		ResourceID:      "chat-123",
		RequestedAction: storage.ActionRead,
	}

	// Sem regras fornecidas
	allowed, err := eval.Evaluate([]*storage.SharedRule{}, req)
	if err != nil {
		t.Fatalf("falha ao avaliar requisição: %v", err)
	}
	if allowed {
		t.Fatalf("esperava permissão negada (Default Deny)")
	}
}

func TestEvaluator_MatchExact_Success(t *testing.T) {
	eval := NewEvaluator()

	rules := []*storage.SharedRule{
		{
			RegistrationID: "user-owner",
			TargetScope:    storage.ScopeChats,
			Pattern:        "chat-secret-1",
			MatchType:      storage.MatchExact,
			AllowedActions: storage.ActionRead | storage.ActionWrite,
		},
	}

	req := AccessRequest{
		RequesterID:     "user-guest",
		ResourceOwnerID: "user-owner",
		TargetScope:     storage.ScopeChats,
		ResourceID:      "chat-secret-1",
		RequestedAction: storage.ActionRead,
	}

	allowed, err := eval.Evaluate(rules, req)
	if err != nil || !allowed {
		t.Fatalf("esperava permissão concedida (MatchExact), obteve allowed=%v, err=%v", allowed, err)
	}
}

func TestEvaluator_MatchPrefix_Success(t *testing.T) {
	eval := NewEvaluator()

	rules := []*storage.SharedRule{
		{
			RegistrationID: "user-owner",
			TargetScope:    storage.ScopeMessages,
			Pattern:        "slack_",
			MatchType:      storage.MatchPrefix,
			AllowedActions: storage.ActionRead,
		},
	}

	req := AccessRequest{
		RequesterID:     "user-guest",
		ResourceOwnerID: "user-owner",
		TargetScope:     storage.ScopeMessages,
		ResourceID:      "slack_msg_99",
		RequestedAction: storage.ActionRead,
	}

	allowed, err := eval.Evaluate(rules, req)
	if err != nil || !allowed {
		t.Fatalf("esperava permissão concedida (MatchPrefix), obteve allowed=%v, err=%v", allowed, err)
	}
}

func TestEvaluator_MatchWildcard_Success(t *testing.T) {
	eval := NewEvaluator()

	rules := []*storage.SharedRule{
		{
			RegistrationID: "user-owner",
			TargetScope:    storage.ScopeTools,
			Pattern:        "web_*",
			MatchType:      storage.MatchWildcard,
			AllowedActions: storage.ActionExecute,
		},
	}

	req := AccessRequest{
		RequesterID:     "user-guest",
		ResourceOwnerID: "user-owner",
		TargetScope:     storage.ScopeTools,
		ResourceID:      "web_search",
		RequestedAction: storage.ActionExecute,
	}

	allowed, err := eval.Evaluate(rules, req)
	if err != nil || !allowed {
		t.Fatalf("esperava permissão concedida (MatchWildcard), obteve allowed=%v, err=%v", allowed, err)
	}
}

func TestEvaluator_MatchRegex_Success(t *testing.T) {
	eval := NewEvaluator()

	rules := []*storage.SharedRule{
		{
			RegistrationID: "user-owner",
			TargetScope:    storage.ScopeChats,
			Pattern:        `^chat_[0-9]+$`,
			MatchType:      storage.MatchRegex,
			AllowedActions: storage.ActionRead,
		},
	}

	req := AccessRequest{
		RequesterID:     "user-guest",
		ResourceOwnerID: "user-owner",
		TargetScope:     storage.ScopeChats,
		ResourceID:      "chat_42",
		RequestedAction: storage.ActionRead,
	}

	allowed, err := eval.Evaluate(rules, req)
	if err != nil || !allowed {
		t.Fatalf("esperava permissão concedida (MatchRegex), obteve allowed=%v, err=%v", allowed, err)
	}
}

func TestEvaluator_MatchRegex_InvalidSyntax(t *testing.T) {
	eval := NewEvaluator()

	rules := []*storage.SharedRule{
		{
			RegistrationID: "user-owner",
			TargetScope:    storage.ScopeChats,
			Pattern:        `[invalid_regex(`, // Expressão regular malformada
			MatchType:      storage.MatchRegex,
			AllowedActions: storage.ActionRead,
		},
	}

	req := AccessRequest{
		RequesterID:     "user-guest",
		ResourceOwnerID: "user-owner",
		TargetScope:     storage.ScopeChats,
		ResourceID:      "chat_1",
		RequestedAction: storage.ActionRead,
	}

	// Não deve dar panic nem erro crítico; deve ignorar a regra e aplicar Default Deny
	allowed, err := eval.Evaluate(rules, req)
	if err != nil {
		t.Fatalf("não esperava erro ao processar regex inválida: %v", err)
	}
	if allowed {
		t.Fatalf("esperava permissão negada (regex inválida deve ser ignorada)")
	}
}

func TestEvaluator_FirstMatchShortCircuit(t *testing.T) {
	eval := NewEvaluator()

	rules := []*storage.SharedRule{
		{
			RegistrationID: "user-owner",
			TargetScope:    storage.ScopeAll,
			Pattern:        "resource-1",
			MatchType:      storage.MatchExact,
			AllowedActions: storage.ActionRead,
		},
		{
			RegistrationID: "user-owner",
			TargetScope:    storage.ScopeAll,
			Pattern:        "[broken_regex", // Regra seguinte quebrada
			MatchType:      storage.MatchRegex,
			AllowedActions: storage.ActionAll,
		},
	}

	req := AccessRequest{
		RequesterID:     "user-guest",
		ResourceOwnerID: "user-owner",
		TargetScope:     storage.ScopeChats,
		ResourceID:      "resource-1",
		RequestedAction: storage.ActionRead,
	}

	allowed, err := eval.Evaluate(rules, req)
	if err != nil || !allowed {
		t.Fatalf("esperava curto-circuito na primeira regra válida, obteve allowed=%v, err=%v", allowed, err)
	}
}

func TestEvaluator_ActionBitmask_Denied(t *testing.T) {
	eval := NewEvaluator()

	rules := []*storage.SharedRule{
		{
			RegistrationID: "user-owner",
			TargetScope:    storage.ScopeChats,
			Pattern:        "chat-1",
			MatchType:      storage.MatchExact,
			AllowedActions: storage.ActionRead, // Apenas leitura
		},
	}

	req := AccessRequest{
		RequesterID:     "user-guest",
		ResourceOwnerID: "user-owner",
		TargetScope:     storage.ScopeChats,
		ResourceID:      "chat-1",
		RequestedAction: storage.ActionWrite, // Solicita escrita
	}

	allowed, err := eval.Evaluate(rules, req)
	if err != nil {
		t.Fatalf("falha ao avaliar requisição: %v", err)
	}
	if allowed {
		t.Fatalf("esperava permissão negada por incompatibilidade de bitmask de ações")
	}
}

func TestEvaluator_ScopeFilter(t *testing.T) {
	eval := NewEvaluator()

	rules := []*storage.SharedRule{
		{
			RegistrationID: "user-owner",
			TargetScope:    storage.ScopeChats, // Apenas chats
			Pattern:        "item-1",
			MatchType:      storage.MatchExact,
			AllowedActions: storage.ActionAll,
		},
	}

	req := AccessRequest{
		RequesterID:     "user-guest",
		ResourceOwnerID: "user-owner",
		TargetScope:     storage.ScopeMessages, // Solicitando mensagens
		ResourceID:      "item-1",
		RequestedAction: storage.ActionRead,
	}

	allowed, err := eval.Evaluate(rules, req)
	if err != nil {
		t.Fatalf("falha ao avaliar requisição: %v", err)
	}
	if allowed {
		t.Fatalf("esperava permissão negada por incompatibilidade de escopo")
	}
}

func TestEvaluator_ScopeAll(t *testing.T) {
	eval := NewEvaluator()

	rules := []*storage.SharedRule{
		{
			RegistrationID: "user-owner",
			TargetScope:    storage.ScopeAll, // Escopo global
			Pattern:        "shared-*",
			MatchType:      storage.MatchWildcard,
			AllowedActions: storage.ActionRead,
		},
	}

	req := AccessRequest{
		RequesterID:     "user-guest",
		ResourceOwnerID: "user-owner",
		TargetScope:     storage.ScopeTools,
		ResourceID:      "shared-tool",
		RequestedAction: storage.ActionRead,
	}

	allowed, err := eval.Evaluate(rules, req)
	if err != nil || !allowed {
		t.Fatalf("esperava permissão concedida para ScopeAll, obteve allowed=%v, err=%v", allowed, err)
	}
}

func TestEvaluator_InvalidArguments(t *testing.T) {
	eval := NewEvaluator()

	_, err := eval.Evaluate(nil, AccessRequest{RequesterID: "", ResourceOwnerID: "owner", ResourceID: "res"})
	if !errors.Is(err, ErrInvalidArgument) {
		t.Fatalf("esperava ErrInvalidArgument para RequesterID vazio, obteve: %v", err)
	}

	_, err = eval.Evaluate(nil, AccessRequest{RequesterID: "req", ResourceOwnerID: "", ResourceID: "res"})
	if !errors.Is(err, ErrInvalidArgument) {
		t.Fatalf("esperava ErrInvalidArgument para ResourceOwnerID vazio, obteve: %v", err)
	}

	_, err = eval.Evaluate(nil, AccessRequest{RequesterID: "req", ResourceOwnerID: "owner", ResourceID: ""})
	if !errors.Is(err, ErrInvalidArgument) {
		t.Fatalf("esperava ErrInvalidArgument para ResourceID vazio, obteve: %v", err)
	}
}

func TestEvaluator_ConcurrentAccess(t *testing.T) {
	eval := NewEvaluator()

	rules := []*storage.SharedRule{
		{
			RegistrationID: "user-owner",
			TargetScope:    storage.ScopeAll,
			Pattern:        `^resource_[0-9]+$`,
			MatchType:      storage.MatchRegex,
			AllowedActions: storage.ActionRead,
		},
	}

	var wg sync.WaitGroup
	const numGoroutines = 100

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			req := AccessRequest{
				RequesterID:     "guest",
				ResourceOwnerID: "user-owner",
				TargetScope:     storage.ScopeChats,
				ResourceID:      fmt.Sprintf("resource_%d", id),
				RequestedAction: storage.ActionRead,
			}
			allowed, err := eval.Evaluate(rules, req)
			if err != nil || !allowed {
				t.Errorf("falha no acesso concorrente para goroutine %d: %v", id, err)
			}
		}(i)
	}

	wg.Wait()
}
