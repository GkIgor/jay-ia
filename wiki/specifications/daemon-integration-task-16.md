# Especificação Técnica de Implementação — Task 16: Daemon Unix Socket Integration

**Projeto:** Jay Core — Fase 3.5  
**Pacote Go:** `core/cmd/jayd` e `core/internal/daemon`  
**Arquivos principais:** `core/cmd/jayd/main.go`, `core/internal/daemon/daemon.go`  
**Dependências obrigatórias:** `sdk/ipc` (Task 09), `core/internal/storage` (Tasks 04-07), `core/internal/service` (Tasks 11-15), `core/internal/api` (Tasks 10-15), `core/internal/ipc` (Task 03)  
**Status:** Aguardando aprovação

---

## 1. Contexto

As Tasks 04 a 15 implementaram toda a arquitetura em camadas do Jay Core:
- Camada de Persistência SQLite (`storage`)
- Sistema Declarativo de Permissões (`permission`)
- Motor de LLM (`llm`)
- Serviços de Aplicação (`service`): `RegistrationService`, `ChatService`, `MessageService`, `ToolService`, `ProcessorService`
- Roteador RPC e Handlers (`api`): `Router`, `RegistrationHandler`, `ChatHandler`, `MessageHandler`, `ToolHandler`, `ProcessorHandler`

A **Task 16** consolida a Fio Condutor da aplicação no executável headless `jayd` (`core/cmd/jayd/main.go`), integrando o servidor Unix Domain Socket com a suíte de serviços e repositórios.

---

## 2. Diagrama de Bootstrap e Arquitetura Unificada

```
[ Sinal SO: SIGINT/SIGTERM ]
            │
            ▼
[ jayd (core/cmd/jayd/main.go) ]
   - Carrega .env (godotenv)
   - Inicializa Daemon (daemon.NewDaemon)
   - Aguarda encerramento gracioso
            │
            ▼
[ Daemon (core/internal/daemon/daemon.go) ]
   ├── 1. StorageEngine (SQLite ~/.jay/jay.db com Pragmas & Migrations)
   ├── 2. Repositories (Registration, SharedRule, Chat, Message, Tool)
   ├── 3. PermissionEvaluator & LLMClient (OpenRouter / Gemini / Mock)
   ├── 4. Services (RegistrationSvc, ChatSvc, MessageSvc, ToolSvc, ProcessorSvc)
   ├── 5. Handlers & Router RPC (RegisterRoutes dos 5 Handlers no Router)
   └── 6. Unix Domain Socket Server (ipc.Server em XDG_RUNTIME_DIR ou ~/.jay/jay.sock)
```

---

## 3. Critérios de Aceite da Task

- [ ] Executável `jayd` compilado com sucesso (`go build ./core/cmd/jayd`).
- [ ] Bootstrap completo integrando Storage Engine (SQLite), Repositórios, Evaluator, Serviços, Handlers e Router RPC.
- [ ] Conexão transparente via Unix Domain Socket entre os envelopes do SDK IPC e o Router RPC.
- [ ] Encerramento gracioso em sinais `SIGINT` e `SIGTERM` sem vazamento de socket ou corrupção de banco.
- [ ] `go vet ./...` e `go test ./...` sem falhas em todo o repositório.
- [ ] 100% dos testes aprovados sob concorrência (`-race`).
