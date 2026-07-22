package api

import (
	"context"
	"errors"

	"github.com/GkIgor/jay-ia/core/internal/service"
	"github.com/GkIgor/jay-ia/sdk/ipc"
)

// ProcessorHandler atua como adaptador de protocolo RPC para o serviço de processamento de chats com a IA.
type ProcessorHandler struct {
	svc *service.ProcessorService
}

// NewProcessorHandler cria uma nova instância de ProcessorHandler.
func NewProcessorHandler(svc *service.ProcessorService) (*ProcessorHandler, error) {
	if svc == nil {
		return nil, errors.New("processor_handler: serviço de processamento não pode ser nulo")
	}
	return &ProcessorHandler{svc: svc}, nil
}

// RegisterRoutes cadastra o handler de MsgProcessChat no Router RPC.
// Operação idempotente e thread-safe.
func (h *ProcessorHandler) RegisterRoutes(router *Router) {
	if router == nil {
		return
	}
	router.Register(ipc.MsgProcessChat, h.handleProcessChat)
}

func (h *ProcessorHandler) handleProcessChat(ctx context.Context, req *ipc.RequestEnvelope) (*ipc.ResponseEnvelope, error) {
	var payload ipc.ProcessChatRequest
	if err := ipc.UnmarshalPayload(req.Payload, &payload); err != nil {
		return nil, err
	}

	msg, err := h.svc.ProcessChat(ctx, req.ClientID, payload.ChatID)
	if err != nil {
		return nil, err
	}

	resp := toProcessChatResponse(msg)
	return ipc.NewResponseEnvelope(req.RequestID, req.Type, resp)
}
