package api

import (
	"context"
	"errors"

	"github.com/GkIgor/jay-ia/core/internal/service"
	"github.com/GkIgor/jay-ia/core/internal/storage"
	"github.com/GkIgor/jay-ia/sdk/ipc"
)

// RegistrationHandler atua como adaptador de protocolo RPC para o serviço de registros.
type RegistrationHandler struct {
	svc *service.RegistrationService
}

// NewRegistrationHandler cria uma nova instância de RegistrationHandler.
func NewRegistrationHandler(svc *service.RegistrationService) (*RegistrationHandler, error) {
	if svc == nil {
		return nil, errors.New("registration_handler: serviço de registro não pode ser nulo")
	}
	return &RegistrationHandler{svc: svc}, nil
}

// RegisterRoutes cadastra os 6 handlers de mensagens de Registration no Router RPC.
// Operação idempotente e thread-safe.
func (h *RegistrationHandler) RegisterRoutes(router *Router) {
	if router == nil {
		return
	}
	router.Register(ipc.MsgRegisterClient, h.handleRegisterClient)
	router.Register(ipc.MsgUnregisterClient, h.handleUnregisterClient)
	router.Register(ipc.MsgUpdateRegistration, h.handleUpdateRegistration)
	router.Register(ipc.MsgGetRegistration, h.handleGetRegistration)
	router.Register(ipc.MsgListRegistrations, h.handleListRegistrations)
	router.Register(ipc.MsgUpdateSharedRules, h.handleUpdateSharedRules)
}

func (h *RegistrationHandler) handleRegisterClient(ctx context.Context, req *ipc.RequestEnvelope) (*ipc.ResponseEnvelope, error) {
	var payload ipc.RegisterClientRequest
	if err := ipc.UnmarshalPayload(req.Payload, &payload); err != nil {
		return nil, err
	}

	clientID := payload.ClientID
	if clientID == "" {
		clientID = req.ClientID
	}

	reg, err := h.svc.RegisterClient(ctx, clientID, payload.Metadata)
	if err != nil {
		return nil, err
	}

	resp := ipc.RegisterClientResponse{
		Registration: toRegistrationDTO(reg),
	}

	return ipc.NewResponseEnvelope(req.RequestID, req.Type, resp)
}

func (h *RegistrationHandler) handleUnregisterClient(ctx context.Context, req *ipc.RequestEnvelope) (*ipc.ResponseEnvelope, error) {
	var payload ipc.UnregisterClientRequest
	_ = ipc.UnmarshalPayload(req.Payload, &payload)

	targetID := payload.ClientID
	if targetID == "" {
		targetID = req.ClientID
	}

	if err := h.svc.UnregisterClient(ctx, req.ClientID, targetID); err != nil {
		return nil, err
	}

	resp := ipc.UnregisterClientResponse{Success: true}
	return ipc.NewResponseEnvelope(req.RequestID, req.Type, resp)
}

func (h *RegistrationHandler) handleUpdateRegistration(ctx context.Context, req *ipc.RequestEnvelope) (*ipc.ResponseEnvelope, error) {
	var payload ipc.UpdateRegistrationRequest
	if err := ipc.UnmarshalPayload(req.Payload, &payload); err != nil {
		return nil, err
	}

	targetID := payload.ClientID
	if targetID == "" {
		targetID = req.ClientID
	}

	reg, err := h.svc.UpdateRegistration(ctx, req.ClientID, targetID, storage.RegistrationStatus(payload.Status), payload.Metadata)
	if err != nil {
		return nil, err
	}

	resp := ipc.UpdateRegistrationResponse{
		Registration: toRegistrationDTO(reg),
	}

	return ipc.NewResponseEnvelope(req.RequestID, req.Type, resp)
}

func (h *RegistrationHandler) handleGetRegistration(ctx context.Context, req *ipc.RequestEnvelope) (*ipc.ResponseEnvelope, error) {
	var payload ipc.GetRegistrationRequest
	_ = ipc.UnmarshalPayload(req.Payload, &payload)

	targetID := payload.RegistrationID
	if targetID == "" {
		targetID = req.ClientID
	}

	reg, err := h.svc.GetRegistration(ctx, req.ClientID, targetID)
	if err != nil {
		return nil, err
	}

	resp := ipc.GetRegistrationResponse{
		Registration: toRegistrationDTO(reg),
	}

	return ipc.NewResponseEnvelope(req.RequestID, req.Type, resp)
}

func (h *RegistrationHandler) handleListRegistrations(ctx context.Context, req *ipc.RequestEnvelope) (*ipc.ResponseEnvelope, error) {
	regs, err := h.svc.ListRegistrations(ctx, req.ClientID)
	if err != nil {
		return nil, err
	}

	resp := ipc.ListRegistrationsResponse{
		Registrations: toRegistrationDTOs(regs),
		Total:         len(regs),
	}

	return ipc.NewResponseEnvelope(req.RequestID, req.Type, resp)
}

func (h *RegistrationHandler) handleUpdateSharedRules(ctx context.Context, req *ipc.RequestEnvelope) (*ipc.ResponseEnvelope, error) {
	var payload ipc.UpdateSharedRulesRequest
	if err := ipc.UnmarshalPayload(req.Payload, &payload); err != nil {
		return nil, err
	}

	count, err := h.svc.UpdateSharedRules(ctx, req.ClientID, payload.Rules)
	if err != nil {
		return nil, err
	}

	resp := ipc.UpdateSharedRulesResponse{
		AppliedRulesCount: count,
	}

	return ipc.NewResponseEnvelope(req.RequestID, req.Type, resp)
}
