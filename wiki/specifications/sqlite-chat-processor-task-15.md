# Especificação Técnica de Implementação — Task 15: Chat Processing Service (`MsgProcessChat`)

**Projeto:** Jay Core — Fase 3.5  
**Pacote Go:** `core/internal/service` e `core/internal/api`  
**Arquivos principais:** `core/internal/service/processor_service.go`, `core/internal/api/processor_handler.go`, `core/internal/api/processor_mapper.go`  
**Dependências obrigatórias:** `sdk/ipc` (Task 09), `core/internal/llm` (Task 02), `core/internal/permission` (Task 08), `core/internal/storage` (Tasks 04-07), Router RPC (Task 10), Convenções de Handlers (Task 11)  
**Status:** Aguardando aprovação

---

## 1. Contexto

As Tasks 11 a 14 entregaram os serviços de recursos para Registros, Chats, Mensagens (CRUD puro) e Ferramentas.

A Task 15 conecta a camada de IPC ao **Orquestrador de IA / Motor de LLM (`llm.Client`)**, implementando o **Serviço de Processamento de Chat (`ProcessorService`)**:

- `MsgProcessChat (350)`: Aciona o ciclo de inferência da IA sobre um Chat específico. O serviço carrega o histórico de mensagens ativas, consulta as ferramentas disponíveis autorizadas para o requisitante, invoca a LLM e persiste atômica e sequencialmente a resposta gerada pelo agente (`AuthorAgent` / `RoleAssistant`).
- Integração com `trigger_agent = true` no `MsgCreateMessage (300)`: Quando `CreateMessage` é invocado com `trigger_agent = true`, o `MessageService` invoca o `ProcessorService.ProcessChat` e preenche o campo `ProcessedMessage` no `CreateMessageResponse`.

---

## 2. Princípios Arquiteturais e Separação de Responsabilidades

```
[ Socket IPC: bytes JSON ]
           │
           ▼
[ Router (core/internal/api) ]
   - Valida envelopes JSON
   - Despacha MsgProcessChat (350)
   - Traduz erros em ErrorCode IPC (ErrorMapper)
   - Isola Panics (recover)
           │
           ▼
[ ProcessorHandler (core/internal/api) ]
   - Desserializa ProcessChatRequest
   - Invoca ProcessorService.ProcessChat
   - Converte a mensagem gerada em ipc.MessageDTO (processor_mapper.go)
   - Constrói ProcessChatResponse
           │
           ▼
[ ProcessorService (core/internal/service) ]
   - Valida autorização preliminar antes do lock
   - Serializa a inferência por chat (chatLocks sync.Map mutexes)
   - Carrega histórico do Chat e converte via toLLMMessages(...)
   - Consulta e converte ferramentas autorizadas via toLLMTools(...)
   - Aplica timeout padrão (30s) e respeita o contexto de cancelamento (ctx.Context)
   - Executa llmClient.GenerateContent(...)
   - Persiste a mensagem gerada pelo agente (AuthorAgent/RoleAssistant)
```

---

## 3. Critérios de Aceite da Task

- [ ] Separação estrita em 3 camadas (`Router` → `ProcessorHandler` → `ProcessorService` → `llm.Client`).
- [ ] Suporte completo ao comando `MsgProcessChat (350)`.
- [ ] Mutex de concorrência por `chatID` via `sync.Map` com autorização prévia fora do lock.
- [ ] Mappers isolados `toLLMMessages` e `toLLMTools`.
- [ ] Filtragem estrita de ferramentas autorizadas por `requesterID`.
- [ ] Persistência de mensagens geradas pela IA com autoria `AuthorAgent` e papel `RoleAssistant`.
- [ ] Testes unitários com `llm.MockClient` validados sem dependência de rede ou API Keys de terceiros.
- [ ] `go vet ./...` e `go test ./...` sem falhas em todo o repositório.
- [ ] 100% dos testes aprovados sob concorrência (`-race`).
