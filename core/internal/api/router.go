package api

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"runtime/debug"
	"sync"

	"github.com/GkIgor/jay-ia/core/internal/permission"
	"github.com/GkIgor/jay-ia/core/internal/storage"
	"github.com/GkIgor/jay-ia/sdk/ipc"
)

// Handler representa a função de processamento de um comando específico do protocolo IPC.
type Handler func(ctx context.Context, req *ipc.RequestEnvelope) (*ipc.ResponseEnvelope, error)

// Middleware encapsula e decora a execução de um Handler.
type Middleware func(next Handler) Handler

// Router gerencia o registro, middlewares e despacho de requisições IPC.
type Router struct {
	handlers    map[ipc.MessageType]Handler
	middlewares []Middleware
	mu          sync.RWMutex
}

// NewRouter instancia um novo Roteador RPC.
func NewRouter() *Router {
	return &Router{
		handlers:    make(map[ipc.MessageType]Handler),
		middlewares: make([]Middleware, 0),
	}
}

// Use adiciona um ou mais middlewares globais ao Roteador.
func (r *Router) Use(middlewares ...Middleware) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.middlewares = append(r.middlewares, middlewares...)
}

// Register associa um Handler a um MessageType.
// Se o comando já estiver cadastrado, sobrescreve o handler anterior (last write wins).
func (r *Router) Register(msgType ipc.MessageType, handler Handler) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.handlers[msgType] = handler
}

// DispatchEnvelope processa o frame bruto de requisição e retorna um *ipc.ResponseEnvelope estruturado.
func (r *Router) DispatchEnvelope(ctx context.Context, rawRequest []byte) *ipc.ResponseEnvelope {
	if len(rawRequest) == 0 {
		return ipc.NewErrorResponseEnvelope("", 0, ipc.ErrInvalidFormat, "envelope de requisição vazio", "")
	}

	var req ipc.RequestEnvelope
	if err := json.Unmarshal(rawRequest, &req); err != nil {
		return ipc.NewErrorResponseEnvelope("", 0, ipc.ErrInvalidFormat, "formato JSON inválido", err.Error())
	}

	if err := ipc.ValidateRequestEnvelope(&req); err != nil {
		switch {
		case errors.Is(err, ipc.ErrInvalidProtocolVersion):
			return ipc.NewErrorResponseEnvelope(req.RequestID, req.Type, ipc.ErrInvalidFormat, "versão do protocolo não suportada", err.Error())
		case errors.Is(err, ipc.ErrMissingRequestID), errors.Is(err, ipc.ErrMissingClientID):
			return ipc.NewErrorResponseEnvelope(req.RequestID, req.Type, ipc.ErrInvalidFormat, "request_id e client_id são obrigatórios", err.Error())
		case errors.Is(err, ipc.ErrUnknownMessageType):
			return ipc.NewErrorResponseEnvelope(req.RequestID, req.Type, ipc.ErrInvalidFormat, "tipo de mensagem inválido", err.Error())
		default:
			return ipc.NewErrorResponseEnvelope(req.RequestID, req.Type, ipc.ErrInvalidFormat, err.Error(), "")
		}
	}

	r.mu.RLock()
	h, exists := r.handlers[req.Type]
	mws := append([]Middleware(nil), r.middlewares...)
	r.mu.RUnlock()

	if !exists || h == nil {
		return ipc.NewErrorResponseEnvelope(req.RequestID, req.Type, ipc.ErrNotImplemented, "comando não suportado pelo servidor", fmt.Sprintf("type=%d", req.Type))
	}

	// Aplicação de middlewares
	finalHandler := h
	for i := len(mws) - 1; i >= 0; i-- {
		finalHandler = mws[i](finalHandler)
	}

	// Execução com isolamento e log de panics
	var resp *ipc.ResponseEnvelope
	var err error

	func() {
		defer func() {
			if p := recover(); p != nil {
				log.Printf("[RPC Router] PANIC ao processar comando type=%d, reqID=%s: %v\n%s", req.Type, req.RequestID, p, debug.Stack())
				resp = ipc.NewErrorResponseEnvelope(req.RequestID, req.Type, ipc.ErrInternalDatabase, "erro interno de execução ao processar comando", fmt.Sprintf("%v", p))
				err = nil
			}
		}()
		resp, err = finalHandler(ctx, &req)
	}()

	if resp != nil {
		return resp
	}

	if err != nil {
		errCode, msg := mapDomainErrorToIPC(err)
		return ipc.NewErrorResponseEnvelope(req.RequestID, req.Type, errCode, msg, err.Error())
	}

	return ipc.NewErrorResponseEnvelope(req.RequestID, req.Type, ipc.ErrInternalDatabase, "handler retornou resposta e erro nulos", "")
}

// Dispatch invoca DispatchEnvelope e serializa a resposta em []byte JSON.
func (r *Router) Dispatch(ctx context.Context, rawRequest []byte) []byte {
	env := r.DispatchEnvelope(ctx, rawRequest)
	bytes, err := json.Marshal(env)
	if err != nil {
		// Fallback seguro em caso de falha catastrófica de serialização JSON
		return []byte(`{"protocol_version":1,"request_id":"","type":0,"status":5000,"error":{"message":"erro interno ao serializar resposta"}}`)
	}
	return bytes
}

// mapDomainErrorToIPC converte erros de domínio Go conhecidos nos códigos de erro IPC.
func mapDomainErrorToIPC(err error) (ipc.ErrorCode, string) {
	if err == nil {
		return ipc.ErrSuccess, ""
	}
	switch {
	case errors.Is(err, storage.ErrNotFound):
		return ipc.ErrNotFound, "registro não encontrado"
	case errors.Is(err, storage.ErrAlreadyExists), errors.Is(err, storage.ErrOwnershipConflict):
		return ipc.ErrConflict, "conflito de recurso ou propriedade"
	case errors.Is(err, storage.ErrInvalidArgument), errors.Is(err, permission.ErrInvalidArgument):
		return ipc.ErrInvalidFormat, "argumento inválido ou malformado"
	case errors.Is(err, storage.ErrForbidden), errors.Is(err, storage.ErrDeleteRestricted):
		return ipc.ErrForbidden, "operação não autorizada ou bloqueada por dependência de recurso"
	case errors.Is(err, storage.ErrInvalidOwner), errors.Is(err, storage.ErrInvalidChat), errors.Is(err, storage.ErrInvalidRegistration):
		return ipc.ErrNotFound, "recurso pai não existe"
	default:
		return ipc.ErrInternalDatabase, err.Error()
	}
}
