# Especificação Técnica de Implementação — Task 11: Registration Service & Handlers

**Projeto:** Jay Core — Fase 3.5  
**Pacote Go:** `core/internal/service` e `core/internal/api`  
**Arquivos principais:** `core/internal/service/registration_service.go`, `core/internal/api/registration_handler.go`  
**Dependências obrigatórias:** `sdk/ipc` (Task 09), `core/internal/permission` (Task 08), `core/internal/storage` (Tasks 03, 06), Router RPC (Task 10)  
**Status:** Aguardando aprovação

---

## 1. Contexto

As Tasks 08, 09 e 10 entregaram o motor de avaliação de permissões (`Permission Evaluator`), o pacote de protocolo IPC (`sdk/ipc`) com DTOs fortemente tipados e o roteador RPC genérico (`Router`).

A Task 11 introduz a **camada de aplicação e caso de uso (`Service`)** em conjunto com os **adaptadores RPC (`Handler`)** para o módulo de **Registros e Identidades Lógicas (`Registration`)**:

- `MsgRegisterClient (100)`: Auto-registro e re-registro idempotente de identidades lógicas.
- `MsgUnregisterClient (101)`: Descadastramento e remoção física de registros.
- `MsgUpdateRegistration (102)`: Atualização de metadados e status de identidades.
- `MsgGetRegistration (103)`: Consulta individual de uma identidade lógica.
- `MsgListRegistrations (104)`: Listagem das identidades conhecidas pelo Core.
- `MsgUpdateSharedRules (105)`: Atualização atômica das regras de compartilhamento declaradas pelo proprietário.

---

## 2. Princípio da Arquitetura em Três Camadas (Router → Handler → Service → Repositories)

```
[ Socket IPC: bytes JSON ]
           │
           ▼
[ Router (core/internal/api) ]
   - Valida envelopes JSON
   - Despacha por MessageType
   - Traduz erros Go em ErrorCode IPC (ErrorMapper)
   - Isola Panics (recover)
           │
           ▼
[ RegistrationHandler (core/internal/api) ]
   - Desserializa Payloads dedicados (ipc.UnmarshalPayload)
   - Invoca os métodos do RegistrationService
   - Converte entidades para DTOs (DTO Mappers)
   - Constrói ipc.ResponseEnvelope
           │
           ▼
[ RegistrationService (core/internal/service) ]
   - Orquestra os Casos de Uso do Domínio
   - Aplica regras de autorização via PermissionEvaluator
   - Executa operações I/O nos Repositories (RegistrationRepository / SharedRuleRepository)
   - Retorna erros de domínio padrão Go
```

---

## 3. Critérios de Aceite da Task

- [ ] Separação limpa em 3 camadas (`Router` → `RegistrationHandler` → `RegistrationService` → Repositories/Evaluator).
- [ ] Mapeamento completo dos 6 comandos do protocolo IPC (100 a 105) com DTOs dedicados.
- [ ] Regras de autorização e caso de uso centralizados no `RegistrationService`.
- [ ] DTO Mappers isolados no `RegistrationHandler`.
- [ ] `go vet ./...` e `go test ./...` sem falhas em todo o repositório.
- [ ] 100% dos testes unitários do Service e dos Handlers IPC aprovados (inclusive com `-race`).
