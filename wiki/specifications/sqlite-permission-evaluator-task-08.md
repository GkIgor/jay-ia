# Especificação Técnica de Implementação — Task 08: Permission Evaluator Engine

**Projeto:** Jay Core — Fase 3.5  
**Pacote Go:** `core/internal/permission`  
**Arquivo principal:** `evaluator.go`  
**Dependências obrigatórias:** `core/internal/storage` (Task 06 - SharedRuleRepository)  
**Status:** Aguardando aprovação

---

## 1. Contexto

As Tasks 01 a 07 estabeleceram a infraestrutura de persistência SQLite completa do Jay Core (`Registration`, `Chat`, `Message`, `SharedRule`, `Tool`).

A Task 08 introduz o **`Permission Evaluator Engine`** (`core/internal/permission`), um **motor puro de avaliação em memória** responsável por determinar de forma determinística e thread-safe se uma requisição de acesso deve ser permitida ou negada.

---

## 2. Princípio de Motor Puro & Desacoplamento de I/O

> **O `Evaluator` é um motor de autorização puro e stateless (em relação a I/O). Ele NÃO realiza NENHUMA operação de leitura de disco ou banco de dados e NÃO realiza nenhuma operação de escrita.**

O carregamento das regras via `SharedRuleRepository` é responsabilidade exclusiva da camada de serviço chamadora (`Service` / `Handler`). O `Evaluator` recebe a lista de `[]*storage.SharedRule` já em memória e calcula a decisão de autorização.

---

## 3. Modelo de Avaliação e Regras de Decisão

A avaliação de acesso pela função `Evaluate(rules []*storage.SharedRule, req AccessRequest) bool` segue a ordem de precedência estrita abaixo:

1. **Propriedade Implícita (Ownership Rule)**: Se `req.RequesterID == req.ResourceOwnerID`, o acesso é **imediatamente PERMITIDO (true)** em $O(1)$ sem avaliar nenhuma regra.
2. **Filtro de Escopo**: Descarta regras onde `rule.TargetScope != ScopeAll` e `rule.TargetScope != req.TargetScope`.
3. **Casamento de Padrão (Pattern Match)**: Avalia o recurso contra `Exact`, `Prefix`, `Wildcard` ou `Regex`.
4. **Máscara de Ações (Bitmask)**: Confirma se `(rule.AllowedActions & req.RequestedAction) == req.RequestedAction`.
5. **Curto-Circuito no Primeiro ALLOW**: Ao encontrar a primeira regra que satisfaça todas as condições acima, encerra a avaliação e retorna `true` imediatamente.
6. **Default Deny**: Se todas as regras forem avaliadas e nenhuma conceder permissão, retorna `false`.

---

## 4. Algoritmos de Casamento de Padrão (Pattern Matching)

| `MatchType` | Constante | Algoritmo Go Aplicado |
|---|---|---|
| **Exact** | `MatchExact (1)` | `rule.Pattern == req.ResourceID` |
| **Prefix** | `MatchPrefix (2)` | `strings.HasPrefix(req.ResourceID, rule.Pattern)` |
| **Wildcard** | `MatchWildcard (3)` | `filepath.Match(rule.Pattern, req.ResourceID)` |
| **Regex** | `MatchRegex (4)` | `compiled.MatchString(req.ResourceID)` com cache via `sync.Map` |

### 4.1. Cache e Tolerância a Erros de Regex
- **Cache de Compilação (`sync.Map`)**: Expressões regulares são compiladas via `regexp.Compile` e armazenadas em `sync.Map` no `Evaluator`.
- **Sintaxe Inválida de Regex**: Se um padrão de regex contiver erro de sintaxe, o `Evaluator` ignora a regra com segurança e prossegue a avaliação sem panics.

---

## 5. Complexidade e Concorrência

- **Ownership Check**: $O(1)$.
- **Rule Iteration**: $O(N)$ no pior caso.
- **Regex Cache Lookup**: $O(1)$ amortizado.
- **Thread Safety**: 100% thread-safe para invocações simultâneas por múltiplas goroutines.

---

## 6. Critérios de Aceite da Task

- [ ] Pacote `core/internal/permission` criado sem dependências de I/O.
- [ ] `evaluator.go` compilando sem erros (`go build ./...`).
- [ ] `go vet ./...` e `go test ./...` sem falhas em todo o repositório.
- [ ] Atendimento $O(1)$ para acesso do proprietário (`RequesterID == OwnerID`).
- [ ] Curto-circuito no primeiro ALLOW (`Short-circuit ALLOW`).
- [ ] Cache de compilação de regexes via `sync.Map` thread-safe.
- [ ] 100% dos testes unitários aprovados (inclusive com `-race`).
