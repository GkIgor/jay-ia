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

- `MsgProcessChat (350)`: Aciona o ciclo de inferência da IA sobre um Chat específico. O serviço carrega o histórico de mensagens ativas, consulta as ferramentas disponíveis no catálogo, invoca a LLM e persiste atômica e sequencialmente a resposta gerada pelo agente (`AuthorAgent` / `RoleAssistant`).
- Suporte a `trigger_agent = true` no `MsgCreateMessage (300)`: Permite que os clientes enviem uma mensagem do usuário e disparem automaticamente o processamento do assistente na mesma requisição RPC.

---

## 2. Princípios Arquiteturais e Fluxo de Execução

```
[ Socket IPC: bytes JSON ]
           │
           ▼
[ Router (core/internal/api) ]
   - Valida envelopes JSON
   - Despacha MsgProcessChat (350)
   - Traduz erros em ErrorCode IPC
   - Isola Panics (recover)
           │
           ▼
[ ProcessorHandler (core/internal/api) ]
   - Desserializa ProcessChatRequest
   - Invoca ProcessorService.ProcessChat
   - Converte a mensagem gerada em ipc.MessageDTO
   - Constrói ProcessChatResponse
           │
           ▼
[ ProcessorService (core/internal/service) ]
   - Valida autorização de escrita no Chat
   - Carrega histórico do Chat via MessageStore
   - Converte mensagens de banco para []llm.Message
   - Consulta ferramentas disponíveis via ToolStore
   - Executa llmClient.GenerateContent(...)
   - Persiste a mensagem gerada pelo agente (AuthorAgent/RoleAssistant)
```

---

## 3. Critérios de Aceite da Task

- [ ] Separação estrita em 3 camadas (`Router` → `ProcessorHandler` → `ProcessorService` → `llm.Client`).
- [ ] Suporte completo ao comando `MsgProcessChat (350)`.
- [ ] Persistência de mensagens geradas pela IA com autoria `AuthorAgent` e papel `RoleAssistant`.
- [ ] Testes unitários com `llm.MockClient` validados sem dependência de rede ou API Keys de terceiros.
- [ ] `go vet ./...` e `go test ./...` sem falhas em todo o repositório.
- [ ] 100% dos testes aprovados sob concorrência (`-race`).
