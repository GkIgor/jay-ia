# Especificação Técnica de Implementação — Task 06: SharedRule Repository

**Projeto:** Jay Core — Fase 3.5  
**Pacote Go:** `core/internal/storage`  
**Arquivo principal:** `shared_rule_repository.go`  
**Dependências obrigatórias:** Task 01 (StorageEngine), Task 02 (MigrationEngine), Task 03 (RegistrationRepository), Task 04 (ChatRepository), Task 05 (MessageRepository)  
**Status:** Concluído

---

## 1. Contexto

As Tasks 01 a 05 entregaram a infraestrutura SQLite, o motor de migrações e os repositórios para `Registration`, `Chat` e `Message`.

A Task 06 introduz a entidade de controle de acesso declarativo: o **`SharedRuleRepository`**, responsável pela persistência das regras de compartilhamento (`SharedRule`) declaradas por cada identidade lógica registrada (`Registration`).

---

## 2. Princípio de Repository Purity & Desacoplamento

> **O SharedRule Repository apenas armazena e recupera regras de compartilhamento. Ele NÃO avalia permissões, NÃO executa o matching de padrões (patterns) e NÃO autoriza requisições.**

---

## 3. Substituição Atômica de Regras (`ReplaceRules`)

Conforme especificado no PRD (Seção 5.2) e no contrato IPC `MsgUpdateSharedRules (type = 105)`:
- A atualização de regras de um registro ocorre sempre por **substituição em bloco**.
- Quando um cliente envia uma nova lista de regras, todas as regras anteriores daquele `registration_id` são removidas e a nova lista é inserida dentro de uma **transação atômica do SQLite (`BEGIN TRANSACTION ... COMMIT`)**:

```sql
BEGIN;
DELETE FROM shared_rules WHERE registration_id = ?;
INSERT INTO shared_rules (registration_id, target_scope, pattern, match_type, allowed_actions)
VALUES (?, ?, ?, ?, ?);
COMMIT;
```

### 3.1. Garantia de Rollback Integral

> **Garantia Atômica:** Se qualquer inserção da nova lista falhar após a remoção das regras anteriores, a transação sofre **Rollback integral** e o conjunto original de regras permanece intacto no banco.

### 3.2. Fonte Única da Verdade para `registration_id`

Ao chamar `ReplaceRules(registrationID string, rules []SharedRule)`:
- O parâmetro `registrationID` da função é a **única fonte da verdade**. Qualquer valor presente no campo `rule.RegistrationID` das structs fornecidas na slice é **ignorado e sobrescrito** pelo parâmetro `registrationID`.

---

## 4. Tipos Go

### 4.1. Enums

```go
type RuleScope int
const (
    ScopeAll      RuleScope = 0
    ScopeChats    RuleScope = 1
    ScopeMessages RuleScope = 2
    ScopeTools    RuleScope = 3
)

type MatchType int
const (
    MatchExact    MatchType = 1
    MatchPrefix   MatchType = 2
    MatchWildcard MatchType = 3
    MatchRegex    MatchType = 4
)

type PermissionAction int
const (
    ActionRead    PermissionAction = 1
    ActionWrite   PermissionAction = 2
    ActionExecute PermissionAction = 4
    ActionAdmin   PermissionAction = 8
    ActionAll     PermissionAction = 15 // (1 | 2 | 4 | 8)
)
```

### 4.2. Struct `SharedRule`

```go
type SharedRule struct {
    ID             int64
    RegistrationID string
    TargetScope    RuleScope
    Pattern        string
    MatchType      MatchType
    AllowedActions PermissionAction
    CreatedAt      string
}
```

---

## 5. Critérios de Aceite da Task

- [x] Sentinela `ErrInvalidRegistration` adicionado a `errors.go`.
- [x] `shared_rule_repository.go` compilando sem erros (`go build ./...`).
- [x] `go vet ./...` e `go test ./...` sem falhas em todo o repositório.
- [x] Substituição atômica em `ReplaceRules` com Rollback integral em falha intermediária.
- [x] `registrationID` do parâmetro sobrescrevendo qualquer valor interno nas structs.
- [x] `strings.TrimSpace(pattern) == ""` rejeitado com `ErrInvalidArgument`.
- [x] Deleção em cascata via SQL (`ON DELETE CASCADE`) confirmada por teste unitário.
- [x] Ordenação `id ASC` em `ListByRegistration`.
