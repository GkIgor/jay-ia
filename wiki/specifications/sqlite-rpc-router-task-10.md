# Especificação Técnica de Implementação — Task 10: Protocol RPC Router & Dispatcher

**Projeto:** Jay Core — Fase 3.5  
**Pacote Go:** `core/internal/api`  
**Arquivo principal:** `router.go`  
**Dependências obrigatórias:** `sdk/ipc` (Task 09), `core/internal/permission` (Task 08), `core/internal/storage` (Tasks 01-07)  
**Status:** Aguardando aprovação

---

## 1. Contexto

A Task 09 implementou o pacote `sdk/ipc`, estabelecendo os tipos numéricos de `MessageType`, `ErrorCode`, envelopes de requisição/resposta (`RequestEnvelope`, `ResponseEnvelope`) e DTOs fortemente tipados.

A Task 10 introduz o **`Protocol RPC Router & Dispatcher`** (`core/internal/api`), responsável por receber frames de bytes JSON do socket Unix, validar a estrutura do envelope, rotear e despachar a requisição para o `Handler` correspondente, traduzir erros de domínio, isolar panics e devolver a resposta formatada em JSON.

---

## 2. Princípios Arquiteturais do Roteador

> **O Roteador de RPC é o ponto de entrada único para requisições IPC no Core. Ele isola a camada de rede de socket das implementações de serviços de domínio, garantindo resiliência, validação centralizada, suporte a middlewares e isolamento de panics.**

---

## 3. Tipos Go, Middlewares e API do Router

```go
type Handler func(ctx context.Context, req *ipc.RequestEnvelope) (*ipc.ResponseEnvelope, error)

type Middleware func(next Handler) Handler

type Router struct {
    handlers    map[ipc.MessageType]Handler
    middlewares []Middleware
    mu          sync.RWMutex
}

func NewRouter() *Router
func (r *Router) Use(middlewares ...Middleware)
func (r *Router) Register(msgType ipc.MessageType, handler Handler)
func (r *Router) DispatchEnvelope(ctx context.Context, rawRequest []byte) *ipc.ResponseEnvelope
func (r *Router) Dispatch(ctx context.Context, rawRequest []byte) []byte
```

---

## 4. Mapeamento Centralizado de Erros de Domínio (`mapDomainErrorToIPC`)

| Erro de Domínio Go | Código `Status` (`ErrorCode`) | Mensagem Retornada |
|---|---|---|
| `storage.ErrNotFound` | `ErrNotFound (4004)` | `"registro não encontrado"` |
| `storage.ErrAlreadyExists` / `storage.ErrOwnershipConflict` | `ErrConflict (4009)` | `"conflito de recurso ou propriedade"` |
| `storage.ErrInvalidArgument` / `permission.ErrInvalidArgument` | `ErrInvalidFormat (4000)` | `"argumento inválido ou malformado"` |
| `storage.ErrDeleteRestricted` | `ErrForbidden (4003)` | `"operação bloqueada por dependência de recurso"` |
| `storage.ErrInvalidOwner` / `ErrInvalidChat` / `ErrInvalidRegistration` | `ErrNotFound (4004)` | `"recurso pai não existe"` |
| Panic ou Erro Interno | `ErrInternalDatabase (5000)` | `"erro interno de execução ao processar comando"` |

---

## 5. Critérios de Aceite da Task

- [ ] Pacote `core/internal/api` criado sem dependências circulares.
- [ ] `router.go` compilando sem erros (`go build ./...`).
- [ ] `go vet ./...` e `go test ./...` sem falhas em todo o repositório.
- [ ] Separação limpa de `DispatchEnvelope` e `Dispatch`.
- [ ] Tradução automática de erros de domínio Go (`storage.ErrNotFound`, etc.) em códigos `ErrorCode` IPC.
- [ ] Suporte a middlewares decoráveis (`Router.Use`).
- [ ] Captura e log de panics via `defer recover()` com stack trace.
- [ ] Liberação antecipada de `RWMutex` antes de executar o handler.
- [ ] 100% dos testes unitários da suíte aprovados (inclusive com `-race`).
