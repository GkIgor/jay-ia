# Especificação Técnica de Implementação — Task 16: Daemon Unix Socket Integration

**Projeto:** Jay Core — Fase 3.5  
**Pacote Go:** `core/cmd/jayd` e `core/internal/daemon`  
**Arquivos principais:** `core/cmd/jayd/main.go`, `core/internal/daemon/daemon.go`  
**Dependências obrigatórias:** `sdk/ipc` (Task 09), `core/internal/storage` (Tasks 04-07), `core/internal/service` (Tasks 11-15), `core/internal/api` (Tasks 10-15), `core/internal/ipc` (Task 03)  
**Status:** Aguardando aprovação

---

## 1. Contexto & Papel do Daemon como Composition Root

As Tasks 04 a 15 implementaram toda a arquitetura lógica e em camadas do Jay Core:
- Camada de Persistência SQLite (`storage`)
- Sistema Declarativo de Permissões (`permission`)
- Motor de LLM (`llm`)
- Serviços de Aplicação (`service`): `RegistrationService`, `ChatService`, `MessageService`, `ToolService`, `ProcessorService`
- Roteador RPC e Handlers (`api`): `Router`, `RegistrationHandler`, `ChatHandler`, `MessageHandler`, `ToolHandler`, `ProcessorHandler`

O pacote `core/internal/daemon` atua como o **Composition Root** exclusivo da aplicação, sendo o único local responsável por construir o grafo completo de dependências sem conter regras de negócio.

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
[ Daemon (core/internal/daemon/daemon.go) — Composition Root ]
   ├── buildStorage()          -> StorageEngine (SQLite ~/.jay/jay.db com Pragmas & Migrations)
   ├── buildRepositories()     -> Repositories (Registration, SharedRule, Chat, Message, Tool)
   ├── buildServices()         -> PermissionEvaluator & LLMClient + 5 Serviços de Aplicação
   ├── buildHandlersAndRouter()-> Router RPC com os 5 Handlers (19 rotas cadastradas)
   └── buildServer()           -> Unix Domain Socket Server (ipc.Server em XDG_RUNTIME_DIR/jay.sock)
```

---

## 3. Critérios de Aceite da Task

- [ ] Executável `jayd` compilado com sucesso (`go build ./core/cmd/jayd`).
- [ ] Bootstrap modularizado via builders privados (`buildStorage`, `buildRepositories`, `buildServices`, `buildHandlersAndRouter`, `buildServer`).
- [ ] Validação estrita de `LLM_PROVIDER` desconhecido com retorno de erro explícito.
- [ ] Conexão transparente via Unix Domain Socket entre os envelopes do SDK IPC e o Router RPC.
- [ ] Encerramento gracioso ordenado (`cancelCtx` → `server.Stop` → `engine.Close`).
- [ ] `go vet ./...` e `go test ./...` sem falhas em todo o repositório.
- [ ] 100% dos testes aprovados sob concorrência (`-race`).
