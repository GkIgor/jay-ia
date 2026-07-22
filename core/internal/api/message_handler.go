package api

import (
	"context"
	"errors"

	"github.com/GkIgor/jay-ia/core/internal/service"
	"github.com/GkIgor/jay-ia/sdk/ipc"
)

// MessageHandler atua como adaptador de protocolo RPC para o serviço de mensagens.
type MessageHandler struct {
	svc *service.MessageService
}

// NewMessageHandler cria uma nova instância de MessageHandler.
func NewMessageHandler(svc *service.MessageService) (*MessageHandler, error) {
	if svc == nil {
		return nil, errors.New("message_handler: serviço de mensagem não pode ser nulo")
	}
	return &MessageHandler{svc: svc}, nil
}

// RegisterRoutes cadastra os 4 handlers de mensagens IPC no Router RPC.
// Operação idempotente e thread-safe.
func (h *MessageHandler) RegisterRoutes(router *Router) {
	if router == nil {
		return
	}
	router.Register(ipc.MsgCreateMessage, h.handleCreateMessage)
	router.Register(ipc.MsgUpdateMessage, h.handleUpdateMessage)
	router.Register(ipc.MsgDeleteMessage, h.handleDeleteMessage)
	router.Register(ipc.MsgGetMessages, h.handleGetMessages)
}

func (h *MessageHandler) handleCreateMessage(ctx context.Context, req *ipc.RequestEnvelope) (*ipc.ResponseEnvelope, error) {
	var payload ipc.CreateMessageRequest
	if err := ipc.UnmarshalPayload(req.Payload, &payload); err != nil {
		return nil, err
	}

	msg, err := h.svc.CreateMessage(ctx, req.ClientID, payload)
	if err != nil {
		return nil, err
	}

	resp := ipc.CreateMessageResponse{
		CreatedMessage: toMessageDTO(msg),
	}

	return ipc.NewResponseEnvelope(req.RequestID, req.Type, resp)
}

func (h *MessageHandler) handleUpdateMessage(ctx context.Context, req *ipc.RequestEnvelope) (*ipc.ResponseEnvelope, error) {
	var payload ipc.UpdateMessageRequest
	if err := ipc.UnmarshalPayload(req.Payload, &payload); err != nil {
		return nil, err
	}

	msg, err := h.svc.UpdateMessage(ctx, req.ClientID, payload.MessageID, payload.NewContent, payload.Metadata)
	if err != nil {
		return nil, err
	}

	resp := ipc.UpdateMessageResponse{
		Message: toMessageDTO(msg),
	}

	return ipc.NewResponseEnvelope(req.RequestID, req.Type, resp)
}

func (h *MessageHandler) handleDeleteMessage(ctx context.Context, req *ipc.RequestEnvelope) (*ipc.ResponseEnvelope, error) {
	var payload ipc.DeleteMessageRequest
	if err := ipc.UnmarshalPayload(req.Payload, &payload); err != nil {
		return nil, err
	}

	if err := h.svc.DeleteMessage(ctx, req.ClientID, payload.MessageID); err != nil {
		return nil, err
	}

	resp := ipc.DeleteMessageResponse{Success: true}
	return ipc.NewResponseEnvelope(req.RequestID, req.Type, resp)
}

func (h *MessageHandler) handleGetMessages(ctx context.Context, req *ipc.RequestEnvelope) (*ipc.ResponseEnvelope, error) {
	var payload ipc.GetMessagesRequest
	if err := ipc.UnmarshalPayload(req.Payload, &payload); err != nil {
		return nil, err
	}

	msgs, hasMore, err := h.svc.GetMessages(ctx, req.ClientID, payload.ChatID, payload.SinceSequenceNo, payload.Limit)
	if err != nil {
		return nil, err
	}

	resp := ipc.GetMessagesResponse{
		ChatID:   payload.ChatID,
		Messages: toMessageDTOs(msgs),
		HasMore:  hasMore,
	}

	return ipc.NewResponseEnvelope(req.RequestID, req.Type, resp)
}
