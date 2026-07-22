# Especificação Técnica de Implementação — Task 14: Tool Service & Handlers

**Projeto:** Jay Core — Fase 3.5  
**Pacote Go:** `core/internal/service` e `core/internal/api`  
**Arquivos principais:** `core/internal/service/tool_service.go`, `core/internal/api/tool_handler.go`, `core/internal/api/tool_mapper.go`  
**Dependências obrigatórias:** `sdk/ipc` (Task 09), `core/internal/permission` (Task 08), `core/internal/storage` (Tasks 07), Router RPC (Task 10), Convenções de Handlers (Task 11)  
**Status:** Aguardando aprovação

---

## 1. Contexto

As Tasks 11, 12 e 13 implementaram os serviços de recursos para Registros, Chats e Mensagens.

A Task 14 introduz o **Serviço de Ferramentas (`ToolService`)** e seus adaptadores RPC (`ToolHandler`), gerenciando o catálogo de capacidades registradas e versionadas pelos consumidores para utilização pelos Agentes do Core:

- `MsgRegisterTool (400)`: Registro e atualização idempotente (Upsert semântico) de ferramentas com versionamento SemVer e proteção contra sequestro de identidade (*Hijack Prevention*).
- `MsgUnregisterTool (401)`: Descadastramento e remoção física da ferramenta do catálogo.
- `MsgGetTool (402)`: Consulta individual dos detalhes e do JSON Schema de uma ferramenta.
- `MsgListTools (403)`: Listagem das ferramentas disponíveis no catálogo global ou filtradas por registro proprietário.

---

## 2. Princípios Arquiteturais e Separação em Camadas

```
[ Socket IPC: bytes JSON ]
           │
           ▼
[ Router (core/internal/api) ]
   - Valida envelopes JSON
   - Despacha por MessageType (400 a 403)
   - Traduz erros Go em ErrorCode IPC (ErrorMapper)
   - Isola Panics (recover)
           │
           ▼
[ ToolHandler (core/internal/api) ]
   - Desserializa Payloads dedicados (RegisterToolRequest, UnregisterToolRequest, etc.)
   - Invoca os métodos do ToolService
   - Converte entidades storage.Tool para ipc.ToolDTO (tool_mapper.go com toToolDTOs)
   - Constrói ipc.ResponseEnvelope
           │
           ▼
[ ToolService (core/internal/service) ]
   - Orquestra os Casos de Uso do Catálogo de Ferramentas
   - Aplica a regra de Hijack Prevention e autorização via PermissionEvaluator
   - Executa I/O através de ToolStore e SharedRuleStore
   - Retorna erros de domínio padrão Go
```

---

## 3. Critérios de Aceite da Task

- [ ] Separação estrita em 3 camadas (`Router` → `ToolHandler` → `ToolService` → Repositories/Evaluator).
- [ ] Mapeamento completo dos 4 comandos de ferramentas (400 a 403) com DTOs dedicados.
- [ ] Garantia de proteção contra Hijack (*Hijack Prevention* via `ErrOwnershipConflict`).
- [ ] Ocultação de segurança em `GetTool` (retorna `storage.ErrNotFound` em negações de autorização).
- [ ] Mappers `toToolDTO` e `toToolDTOs` centralizados em `tool_mapper.go`.
- [ ] `go vet ./...` e `go test ./...` sem falhas em todo o repositório.
- [ ] 100% dos testes unitários do Service e dos Handlers IPC aprovados (inclusive com `-race`).
