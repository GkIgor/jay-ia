# Especificação Técnica de Implementação — Task 10: Protocol RPC Router & Dispatcher

**Projeto:** Jay Core — Fase 3.5  
**Pacote Go:** `core/internal/api`  
**Arquivo principal:** `router.go`  
**Dependências obrigatórias:** `sdk/ipc` (Task 09), `core/internal/permission` (Task 08)  
**Status:** Aguardando aprovação

---

## 1. Contexto

A Task 09 implementou o pacote `sdk/ipc`, estabelecendo os tipos numéricos de `MessageType`, `ErrorCode`, envelopes de requisição/resposta (`RequestEnvelope`, `ResponseEnvelope`) e DTOs fortemente tipados.

A Task 10 introduz o **`Protocol RPC Router & Dispatcher`** (`core/internal/api`), responsável por receber frames de bytes JSON do socket Unix, validar a estrutura do envelope, rotear e despachar a requisição para o `Handler` correspondente, isolar falhas/panics e devolver a resposta formatada como payload JSON de resposta.

---

## 2. Princípios Arquiteturais do Roteador

> **O Roteador de RPC é o ponto de entrada único para requisições IPC no Core. Ele isola a camada de rede de socket das implementações de serviços de domínio, garantindo resiliência, validação centralizada e isolamento de panics.**

---

## 3. Tipos Go e Assinatura dos Handlers

```go
type Handler func(ctx context.Context, req *ipc.RequestEnvelope) (*ipc.ResponseEnvelope, error)

type Router struct {
    handlers map[ipc.MessageType]Handler
    mu       sync.RWMutex
}

func NewRouter() *Router
func (r *Router) Register(msgType ipc.MessageType, handler Handler)
func (r *Router) Dispatch(ctx context.Context, rawRequest []byte) []byte
```

---

## 4. Mapeamento de Erros e Respostas no Roteador

| Erro / Situação | Código `Status` (`ErrorCode`) | Mensagem Retornada |
|---|---|---|
| JSON malformado / inválido no frame | `ErrInvalidFormat (4000)` | `"formato JSON inválido"` |
| `protocol_version` incompatível | `ErrInvalidFormat (4000)` | `"versão do protocolo não suportada"` |
| `request_id` ou `client_id` ausentes | `ErrInvalidFormat (4000)` | `"request_id e client_id são obrigatórios"` |
| `MessageType` não cadastrado no Router | `ErrNotImplemented (5001)` | `"comando não suportado pelo servidor"` |
| Panic capturado durante a execução do Handler | `ErrInternalDatabase (5000)` | `"erro interno de execução ao processar comando"` |

---

## 5. Critérios de Aceite da Task

- [ ] Pacote `core/internal/api` criado sem dependências circulares.
- [ ] `router.go` compilando sem erros (`go build ./...`).
- [ ] `go vet ./...` e `go test ./...` sem falhas em todo o repositório.
- [ ] Roteamento dinâmico de `MessageType` para Handlers registrados.
- [ ] Captura de panics em handlers via `defer recover()` produzindo erro 5000.
- [ ] 100% dos testes unitários da suíte aprovados (inclusive com `-race`).
