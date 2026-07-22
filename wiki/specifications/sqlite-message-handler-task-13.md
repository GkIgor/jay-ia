# Especificação Técnica de Implementação — Task 13: Message Service & Handlers (CRUD sem IA)

**Projeto:** Jay Core — Fase 3.5  
**Pacote Go:** `core/internal/service` e `core/internal/api`  
**Arquivos principais:** `core/internal/service/message_service.go`, `core/internal/api/message_handler.go`, `core/internal/api/message_mapper.go`  
**Dependências obrigatórias:** `sdk/ipc` (Task 09), `core/internal/permission` (Task 08), `core/internal/storage` (Tasks 05, 06), Router RPC (Task 10), Convenções de Handlers (Task 11)  
**Status:** Aguardando aprovação

---

## 1. Contexto

A Task 12 entregou o `ChatService` e `ChatHandler`, permitindo a gestão dos containers de conversação.

A Task 13 implementa o **CRUD puro do Serviço de Mensagens (`MessageService`)** desacoplado da execução do motor de IA (que será acoplado na Task 15 via `Chat Processing Service`):

- `MsgCreateMessage (300)`: Inserção atômica de mensagens com controle sequencial (`sequence_no`), autoria composta (`AuthorType`, `AuthorID`) e marcações de papel (`Role`).
- `MsgUpdateMessage (301)`: Edição do conteúdo de mensagens (`status = 2` / `MessageEdited`).
- `MsgDeleteMessage (302)`: Exclusão lógica (*Soft Delete*, `status = 3` / `MessageDeleted`) preservando a integridade dos sequenciais do chat.
- `MsgGetMessages (303)`: Consulta por histórico de mensagens usando o Modelo **Pull** (`since_sequence_no` e `limit`).

---

## 2. Princípios Arquiteturais e Separação em Camadas

```
[ Socket IPC: bytes JSON ]
           │
           ▼
[ Router (core/internal/api) ]
   - Valida envelopes JSON
   - Despacha por MessageType (300 a 303)
   - Traduz erros Go em ErrorCode IPC (ErrorMapper)
   - Isola Panics (recover)
           │
           ▼
[ MessageHandler (core/internal/api) ]
   - Desserializa Payloads dedicados (CreateMessageRequest, UpdateMessageRequest, GetMessagesRequest, etc.)
   - Invoca os métodos do MessageService
   - Converte entidades storage.Message para ipc.MessageDTO (message_mapper.go com toMessageDTOs)
   - Constrói ipc.ResponseEnvelope
           │
           ▼
[ MessageService (core/internal/service) ]
   - Orquestra o CRUD de Mensagens
   - Valida autorização de acesso ao Chat correspondente via PermissionEvaluator (helpers internos)
   - Executa I/O através de MessageStore, ChatStore e SharedRuleStore
   - Mantém o incremento atômico de sequence_no
```

---

## 3. Critérios de Aceite da Task

- [ ] Separação estrita em 3 camadas (`Router` → `MessageHandler` → `MessageService` → Repositories/Evaluator).
- [ ] Mapeamento completo dos 4 comandos de mensagem (300 a 303) com DTOs dedicados.
- [ ] Incremento sequencial atômico de `sequence_no` por chat.
- [ ] Semântica estrita do Pull Model (`sequence_no > since_sequence_no`, ordenação `sequence_no ASC`, `has_more` via `limit + 1`).
- [ ] Ocultação do conteúdo de mensagens em Soft Delete (`content = ""`).
- [ ] Ocultação de segurança em `GetMessages` (retorna `storage.ErrNotFound` em negações de autorização).
- [ ] Mappers `toMessageDTO` e `toMessageDTOs` centralizados em `message_mapper.go`.
- [ ] `go vet ./...` e `go test ./...` sem falhas em todo o repositório.
- [ ] 100% dos testes unitários do Service e dos Handlers IPC aprovados (inclusive com `-race`).
