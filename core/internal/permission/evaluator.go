package permission

import (
	"path/filepath"
	"regexp"
	"strings"
	"sync"

	"github.com/GkIgor/jay-ia/core/internal/storage"
)

// AccessRequest representa os parâmetros de uma verificação de permissão de acesso.
type AccessRequest struct {
	RequesterID     string
	ResourceOwnerID string
	TargetScope     storage.RuleScope
	ResourceID      string
	RequestedAction storage.PermissionAction
}

// Evaluator é o motor puro de autorização e verificação de regras de compartilhamento em memória.
// É thread-safe e seguro para invocações concorrentes.
type Evaluator struct {
	regexCache sync.Map // cache de *regexp.Regexp por padrão de string
}

// NewEvaluator instancia um novo motor de avaliação de permissões.
func NewEvaluator() *Evaluator {
	return &Evaluator{}
}

// Evaluate avalia se a requisição de acesso é permitida (true) ou negada (false) com base nas regras fornecidas.
// Retorna ErrInvalidArgument se RequesterID, ResourceOwnerID ou ResourceID forem vazios.
func (e *Evaluator) Evaluate(rules []*storage.SharedRule, req AccessRequest) (bool, error) {
	cleanRequester := strings.TrimSpace(req.RequesterID)
	cleanOwner := strings.TrimSpace(req.ResourceOwnerID)
	cleanResource := strings.TrimSpace(req.ResourceID)

	if cleanRequester == "" || cleanOwner == "" || cleanResource == "" {
		return false, ErrInvalidArgument
	}

	// 1. Regra de Propriedade Implícita (Ownership Rule): o proprietário sempre tem acesso total
	if cleanRequester == cleanOwner {
		return true, nil
	}

	// 2. Avaliação de Regras de Compartilhamento em Memória
	for _, rule := range rules {
		if rule == nil {
			continue
		}

		// Filtro de Escopo: a regra deve se aplicar ao escopo específico ou a ScopeAll (0)
		if rule.TargetScope != storage.ScopeAll && rule.TargetScope != req.TargetScope {
			continue
		}

		// Casamento de Padrão (Pattern Match)
		matched := e.matchPattern(rule.MatchType, rule.Pattern, cleanResource)
		if !matched {
			continue
		}

		// Validação Bitwise de Ação: a regra deve conter todas as ações solicitadas
		if (rule.AllowedActions & req.RequestedAction) == req.RequestedAction {
			// Curto-circuito no primeiro ALLOW
			return true, nil
		}
	}

	// 3. Default Deny: se nenhuma regra autorizar a ação, o acesso é negado
	return false, nil
}

// matchPattern executa o algoritmo de casamento correspondente ao MatchType.
func (e *Evaluator) matchPattern(matchType storage.MatchType, pattern string, resourceID string) bool {
	cleanPattern := strings.TrimSpace(pattern)
	if cleanPattern == "" {
		return false
	}

	switch matchType {
	case storage.MatchExact:
		return cleanPattern == resourceID

	case storage.MatchPrefix:
		return strings.HasPrefix(resourceID, cleanPattern)

	case storage.MatchWildcard:
		matched, err := filepath.Match(cleanPattern, resourceID)
		if err != nil {
			return false
		}
		return matched

	case storage.MatchRegex:
		var re *regexp.Regexp
		if cached, ok := e.regexCache.Load(cleanPattern); ok {
			re = cached.(*regexp.Regexp)
		} else {
			compiled, err := regexp.Compile(cleanPattern)
			if err != nil {
				// Padrão de regex inválido: ignora a regra com segurança sem panic
				return false
			}
			e.regexCache.Store(cleanPattern, compiled)
			re = compiled
		}
		return re.MatchString(resourceID)

	default:
		return false
	}
}
