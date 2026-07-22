package api

import (
	"context"
	"errors"

	"github.com/GkIgor/jay-ia/core/internal/service"
	"github.com/GkIgor/jay-ia/sdk/ipc"
)

// ToolHandler atua como adaptador de protocolo RPC para o catálogo de ferramentas.
type ToolHandler struct {
	svc *service.ToolService
}

// NewToolHandler cria uma nova instância de ToolHandler.
func NewToolHandler(svc *service.ToolService) (*ToolHandler, error) {
	if svc == nil {
		return nil, errors.New("tool_handler: serviço de ferramenta não pode ser nulo")
	}
	return &ToolHandler{svc: svc}, nil
}

// RegisterRoutes cadastra os 4 handlers de ferramentas no Router RPC.
// Operação idempotente e thread-safe.
func (h *ToolHandler) RegisterRoutes(router *Router) {
	if router == nil {
		return
	}
	router.Register(ipc.MsgRegisterTool, h.handleRegisterTool)
	router.Register(ipc.MsgUnregisterTool, h.handleUnregisterTool)
	router.Register(ipc.MsgGetTool, h.handleGetTool)
	router.Register(ipc.MsgListTools, h.handleListTools)
}

func (h *ToolHandler) handleRegisterTool(ctx context.Context, req *ipc.RequestEnvelope) (*ipc.ResponseEnvelope, error) {
	var payload ipc.RegisterToolRequest
	if err := ipc.UnmarshalPayload(req.Payload, &payload); err != nil {
		return nil, err
	}

	tool, err := h.svc.RegisterTool(ctx, req.ClientID, payload)
	if err != nil {
		return nil, err
	}

	resp := ipc.RegisterToolResponse{
		Tool: toToolDTO(tool),
	}

	return ipc.NewResponseEnvelope(req.RequestID, req.Type, resp)
}

func (h *ToolHandler) handleUnregisterTool(ctx context.Context, req *ipc.RequestEnvelope) (*ipc.ResponseEnvelope, error) {
	var payload ipc.UnregisterToolRequest
	if err := ipc.UnmarshalPayload(req.Payload, &payload); err != nil {
		return nil, err
	}

	if err := h.svc.UnregisterTool(ctx, req.ClientID, payload.ToolID); err != nil {
		return nil, err
	}

	resp := ipc.UnregisterToolResponse{Success: true}
	return ipc.NewResponseEnvelope(req.RequestID, req.Type, resp)
}

func (h *ToolHandler) handleGetTool(ctx context.Context, req *ipc.RequestEnvelope) (*ipc.ResponseEnvelope, error) {
	var payload ipc.GetToolRequest
	if err := ipc.UnmarshalPayload(req.Payload, &payload); err != nil {
		return nil, err
	}

	tool, err := h.svc.GetTool(ctx, req.ClientID, payload.ToolID)
	if err != nil {
		return nil, err
	}

	resp := ipc.GetToolResponse{
		Tool: toToolDTO(tool),
	}

	return ipc.NewResponseEnvelope(req.RequestID, req.Type, resp)
}

func (h *ToolHandler) handleListTools(ctx context.Context, req *ipc.RequestEnvelope) (*ipc.ResponseEnvelope, error) {
	var payload ipc.ListToolsRequest
	_ = ipc.UnmarshalPayload(req.Payload, &payload)

	tools, err := h.svc.ListTools(ctx, req.ClientID, payload.RegistrationID)
	if err != nil {
		return nil, err
	}

	resp := ipc.ListToolsResponse{
		Tools: toToolDTOs(tools),
	}

	return ipc.NewResponseEnvelope(req.RequestID, req.Type, resp)
}
