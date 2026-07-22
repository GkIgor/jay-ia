# Especificação Técnica de Implementação — Task 12: Chat Service & Handlers

**Projeto:** Jay Core — Fase 3.5  
**Pacote Go:** `core/internal/service` e `core/internal/api`  
**Arquivos principais:** `core/internal/service/chat_service.go`, `core/internal/api/chat_handler.go`, `core/internal/api/chat_mapper.go`  
**Dependências obrigatórias:** `sdk/ipc` (Task 09), `core/internal/permission` (Task 08), `core/internal/storage` (Tasks 04, 06), Router RPC (Task 10), Convenções de Handlers (Task 11)  
**Status:** Aguardando aprovação

---

## 1. Contexto

A Task 11 estabeleceu o padrão de arquitetura em três camadas (`Router` → `Handler` → `Service` → `Repositories` & `Evaluator`) e implementou o módulo de Registros.

A Task 12 implementa o módulo de **Gerenciamento de Chats (`Chat`)**:

- `MsgCreateChat (200)`: Criação de novos containers de conversação mantidos pelo Core.
- `MsgDeleteChat (201)`: Exclusão lógica (*Soft Delete*, `status = 3`) preservando histórico e dados estatísticos.
- `MsgRenameChat (202)`: Atualização do título do chat.
- `MsgGetChat (203)`: Consulta individual de detalhes do chat.
- `MsgListChats (204)`: Listagem dos chats pertencentes ao consumidor, com suporte opcional à inclusão de chats compartilhados (`include_shared = true`).

---

## 2. Princípios Arquiteturais e Separação em Camadas

```
[ Socket IPC: bytes JSON ]
           │
           ▼
[ Router (core/internal/api) ]
   - Valida envelopes JSON
   - Despacha por MessageType (200 a 204)
   - Traduz erros Go em ErrorCode IPC (ErrorMapper)
   - Isola Panics (recover)
           │
           ▼
[ ChatHandler (core/internal/api) ]
   - Desserializa Payloads dedicados (CreateChatRequest, DeleteChatRequest, etc.)
   - Invoca os métodos do ChatService
   - Converte entidades storage.Chat para ipc.ChatDTO (chat_mapper.go)
   - Constrói ipc.ResponseEnvelope
           │
           ▼
[ ChatService (core/internal/service) ]
   - Orquestra os Casos de Uso de Chats
   - Valida autorização do requisitante via PermissionEvaluator
   - Executa I/O através de ChatStore e SharedRuleStore
   - Retorna erros de domínio padrão Go
```

---

## 3. Critérios de Aceite da Task

- [ ] Separação estrita nas 3 camadas (`Router` → `ChatHandler` → `ChatService` → Repositories/Evaluator).
- [ ] Mapeamento dos 5 comandos de Chat (200 a 204) com DTOs dedicados.
- [ ] Ocultação de segurança em `GetChat` (retorna `storage.ErrNotFound` para não autorizados).
- [ ] Suporte a *Soft Delete* preservando dados do chat.
- [ ] Suporte a consulta de chats compartilhados via `include_shared = true`.
- [ ] `go vet ./...` e `go test ./...` sem falhas em todo o repositório.
- [ ] 100% dos testes unitários do Service e dos Handlers IPC aprovados (inclusive com `-race`).
