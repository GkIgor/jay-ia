# Especificação Técnica de Implementação — Task 08: Permission Evaluator Engine

**Projeto:** Jay Core — Fase 3.5  
**Pacote Go:** `core/internal/permission`  
**Arquivo principal:** `evaluator.go`  
**Dependências obrigatórias:** `core/internal/storage` (Task 06 - SharedRuleRepository)  
**Status:** Aguardando aprovação

---

## 1. Contexto

As Tasks 01 a 07 estabeleceram a infraestrutura de persistência SQLite completa do Jay Core (`Registration`, `Chat`, `Message`, `SharedRule`, `Tool`).

A Task 08 introduz o **`Permission Evaluator Engine`** (`core/internal/permission`), responsável por avaliar declarativamente e em memória se um registro requisitante tem autorização para realizar uma ação sobre um recurso pertencente a outra identidade no Jay Core.

---

## 2. Princípio de Separação de Responsabilidades

> **O Evaluator apenas calcula e avalia permissões em memória a partir das regras recuperadas do `SharedRuleRepository`. Ele NÃO realiza I/O direto além da leitura de regras e NÃO altera estados de banco de dados.**

---

## 3. Modelo de Avaliação e Regras de Decisão

A avaliação de acesso segue a ordem de precedência estrita abaixo:

1. **Propriedade Implícita (Ownership Rule)**: Se `RequesterID == ResourceOwnerID`, acesso imediato **PERMITIDO (true)**.
2. **Consulta de Regras**: Se `RequesterID != ResourceOwnerID`, busca as regras cadastradas pelo `ResourceOwnerID` via `SharedRuleRepository.ListByRegistration`.
3. **Filtro de Escopo**: Descarta regras que não sejam para o escopo solicitado ou `ScopeAll`.
4. **Casamento de Padrão (Pattern Match)**: Avalia o recurso contra `Exact`, `Prefix`, `Wildcard` ou `Regex`.
5. **Máscara de Ações (Bitmask)**: Confirma se `(rule.AllowedActions & requestedAction) == requestedAction`.
6. **Default Deny**: Se nenhuma regra casar e autorizar, o acesso é **NEGADO (false)**.

---

## 4. Algoritmos de Casamento de Padrão (Pattern Matching)

| `MatchType` | Constante | Algoritmo Go Aplicado |
|---|---|---|
| **Exact** | `MatchExact (1)` | `rule.Pattern == resourceID` |
| **Prefix** | `MatchPrefix (2)` | `strings.HasPrefix(resourceID, rule.Pattern)` |
| **Wildcard** | `MatchWildcard (3)` | `filepath.Match(rule.Pattern, resourceID)` |
| **Regex** | `MatchRegex (4)` | `regexp.MatchString(rule.Pattern, resourceID)` com cache via `sync.Map` |

---

## 5. Estrutura do Pacote `core/internal/permission`

```
core/internal/permission/
├── evaluator.go        [NOVO — Motor de Avaliação de Permissões]
└── evaluator_test.go   [NOVO — Suíte Completa de Testes Unitários]
```

---

## 6. Critérios de Aceite da Task

- [ ] Pacote `core/internal/permission` criado sem dependências circulares.
- [ ] `evaluator.go` compilando sem erros (`go build ./...`).
- [ ] `go vet ./...` e `go test ./...` sem falhas em todo o repositório.
- [ ] Atendimento de acesso imediato para o próprio proprietário (`RequesterID == OwnerID`).
- [ ] Cache de compilação de expressões regulares via `sync.Map`.
- [ ] Tratamento gracioso de erros de sintaxe de Regex sem gerar panics.
- [ ] 100% dos testes unitários da suíte aprovados.
