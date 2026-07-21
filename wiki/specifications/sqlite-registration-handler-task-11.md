# Especificação Técnica de Implementação — Task 11: Registration Service & Handlers

**Projeto:** Jay Core — Fase 3.5  
**Pacote Go:** `core/internal/api`  
**Arquivo principal:** `registration_handler.go`  
**Dependências obrigatórias:** `sdk/ipc` (Task 09), `core/internal/permission` (Task 08), `core/internal/storage` (Tasks 03, 06), `core/internal/api` Router (Task 10)  
**Status:** Aguardando aprovação

---

## 1. Contexto

As Tasks 08, 09 e 10 entregaram o motor de avaliação de permissões (`Permission Evaluator`), o pacote de protocolo IPC com DTOs e envelopes fortemente tipados (`sdk/ipc`) e o roteador RPC genérico (`Router`).

A Task 11 abre a camada de **Serviços de Recursos & Handlers RPC**, implementando a lógica de negócio e os handlers RPC para o módulo de **Registros e Identidades Lógicas (`Registration`)**:

- `MsgRegisterClient (100)`: Registro e re-registro idempotente de identidades lógicas.
- `MsgUnregisterClient (101)`: Descadastramento e remoção de registros.
- `MsgUpdateRegistration (102)`: Atualização de metadados e status de identidades.
- `MsgGetRegistration (103)`: Consulta individual de uma identidade lógica.
- `MsgListRegistrations (104)`: Listagem das identidades conhecidas pelo Core.
- `MsgUpdateSharedRules (105)`: Atualização atômica das regras de compartilhamento declaradas.

---

## 2. Princípio de Separação de Responsabilidades

> **Os Handlers RPC são adaptadores de protocolo. Eles desserializam os payloads de `ipc.RequestEnvelope`, invocam a verificação de acesso do `Evaluator`, chamam os métodos de repositório (`RegistrationRepository` / `SharedRuleRepository`) e constroem envelopes `ipc.ResponseEnvelope`.**

---

## 3. Estrutura do Pacote e Injeção de Dependências

```go
type RegistrationHandler struct {
    regRepo   *storage.RegistrationRepository
    ruleRepo  *storage.SharedRuleRepository
    evaluator *permission.Evaluator
}

func NewRegistrationHandler(
    regRepo *storage.RegistrationRepository,
    ruleRepo *storage.SharedRuleRepository,
    evaluator *permission.Evaluator,
) (*RegistrationHandler, error)

func (h *RegistrationHandler) RegisterRoutes(router *Router)
```

---

## 4. Critérios de Aceite da Task

- [ ] `registration_handler.go` compilando sem erros (`go build ./...`).
- [ ] `go vet ./...` e `go test ./...` sem falhas em todo o repositório.
- [ ] Mapeamento completo dos 6 comandos numéricos do protocolo IPC (100 a 105).
- [ ] Auto-registro e re-registro idempotente funcionando via `RegisterClient`.
- [ ] Substituição atômica de regras de compartilhamento via `UpdateSharedRules`.
- [ ] 100% dos testes de integração RPC do módulo aprovados (inclusive com `-race`).
