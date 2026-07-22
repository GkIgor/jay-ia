package api

import (
	"context"
	"errors"

	"github.com/GkIgor/jay-ia/core/internal/service"
	"github.com/GkIgor/jay-ia/sdk/ipc"
)

// ChatHandler atua como adaptador de protocolo RPC para o serviço de chats.
type ChatHandler struct {
	svc *service.ChatService
}

// NewChatHandler cria uma nova instância de ChatHandler.
func NewChatHandler(svc *service.ChatService) (*ChatHandler, error) {
	if svc == nil {
		return nil, errors.New("chat_handler: serviço de chat não pode ser nulo")
	}
	return &ChatHandler{svc: svc}, nil
}

// RegisterRoutes cadastra os 5 handlers de mensagens de Chat no Router RPC.
// Operação idempotente e thread-safe.
func (h *ChatHandler) RegisterRoutes(router *Router) {
	if router == nil {
		return
	}
	router.Register(ipc.MsgCreateChat, h.handleCreateChat)
	router.Register(ipc.MsgDeleteChat, h.handleDeleteChat)
	router.Register(ipc.MsgRenameChat, h.handleRenameChat)
	router.Register(ipc.MsgGetChat, h.handleGetChat)
	router.Register(ipc.MsgListChats, h.handleListChats)
}

func (h *ChatHandler) handleCreateChat(ctx context.Context, req *ipc.RequestEnvelope) (*ipc.ResponseEnvelope, error) {
	var payload ipc.CreateChatRequest
	if err := ipc.UnmarshalPayload(req.Payload, &payload); err != nil {
		return nil, err
	}

	chat, err := h.svc.CreateChat(ctx, req.ClientID, payload.Title, payload.Metadata)
	if err != nil {
		return nil, err
	}

	resp := ipc.CreateChatResponse{
		Chat: toChatDTO(chat, req.ClientID),
	}

	return ipc.NewResponseEnvelope(req.RequestID, req.Type, resp)
}

func (h *ChatHandler) handleDeleteChat(ctx context.Context, req *ipc.RequestEnvelope) (*ipc.ResponseEnvelope, error) {
	var payload ipc.DeleteChatRequest
	if err := ipc.UnmarshalPayload(req.Payload, &payload); err != nil {
		return nil, err
	}

	if err := h.svc.DeleteChat(ctx, req.ClientID, payload.ChatID); err != nil {
		return nil, err
	}

	resp := ipc.DeleteChatResponse{Success: true}
	return ipc.NewResponseEnvelope(req.RequestID, req.Type, resp)
}

func (h *ChatHandler) handleRenameChat(ctx context.Context, req *ipc.RequestEnvelope) (*ipc.ResponseEnvelope, error) {
	var payload ipc.RenameChatRequest
	if err := ipc.UnmarshalPayload(req.Payload, &payload); err != nil {
		return nil, err
	}

	chat, err := h.svc.RenameChat(ctx, req.ClientID, payload.ChatID, payload.NewTitle)
	if err != nil {
		return nil, err
	}

	resp := ipc.RenameChatResponse{
		Chat: toChatDTO(chat, req.ClientID),
	}

	return ipc.NewResponseEnvelope(req.RequestID, req.Type, resp)
}

func (h *ChatHandler) handleGetChat(ctx context.Context, req *ipc.RequestEnvelope) (*ipc.ResponseEnvelope, error) {
	var payload ipc.GetChatRequest
	if err := ipc.UnmarshalPayload(req.Payload, &payload); err != nil {
		return nil, err
	}

	chat, err := h.svc.GetChat(ctx, req.ClientID, payload.ChatID)
	if err != nil {
		return nil, err
	}

	resp := ipc.GetChatResponse{
		Chat: toChatDTO(chat, req.ClientID),
	}

	return ipc.NewResponseEnvelope(req.RequestID, req.Type, resp)
}

func (h *ChatHandler) handleListChats(ctx context.Context, req *ipc.RequestEnvelope) (*ipc.ResponseEnvelope, error) {
	var payload ipc.ListChatsRequest
	_ = ipc.UnmarshalPayload(req.Payload, &payload)

	chats, err := h.svc.ListChats(ctx, req.ClientID, payload.IncludeShared, payload.Limit)
	if err != nil {
		return nil, err
	}

	resp := ipc.ListChatsResponse{
		Chats: toChatDTOs(chats, req.ClientID),
		Total: len(chats),
	}

	return ipc.NewResponseEnvelope(req.RequestID, req.Type, resp)
}
