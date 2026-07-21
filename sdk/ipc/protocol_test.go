package ipc

import (
	"encoding/json"
	"errors"
	"testing"
)

func TestProtocol_NewRequestEnvelope(t *testing.T) {
	payload := CreateChatRequest{Title: "Chat Teste"}
	req, err := NewRequestEnvelope(MsgCreateChat, "jay_client_cpp", payload)
	if err != nil {
		t.Fatalf("falha ao criar RequestEnvelope: %v", err)
	}

	if req.ProtocolVersion != ProtocolVersionCurrent {
		t.Errorf("esperava ProtocolVersion %d, obteve %d", ProtocolVersionCurrent, req.ProtocolVersion)
	}
	if req.ClientID != "jay_client_cpp" {
		t.Errorf("esperava ClientID 'jay_client_cpp', obteve '%s'", req.ClientID)
	}
	if req.Type != MsgCreateChat {
		t.Errorf("esperava Type %d, obteve %d", MsgCreateChat, req.Type)
	}
	if req.RequestID == "" {
		t.Errorf("RequestID não deveria ser vazio")
	}

	var parsedPayload CreateChatRequest
	if err := UnmarshalPayload(req.Payload, &parsedPayload); err != nil {
		t.Fatalf("falha ao desserializar payload: %v", err)
	}
	if parsedPayload.Title != "Chat Teste" {
		t.Errorf("esperava Title 'Chat Teste', obteve '%s'", parsedPayload.Title)
	}
}

func TestProtocol_NilPayload_NormalizesToEmptyObject(t *testing.T) {
	req, err := NewRequestEnvelope(MsgListChats, "cli", nil)
	if err != nil {
		t.Fatalf("falha ao criar request com nil payload: %v", err)
	}

	if string(req.Payload) != "{}" {
		t.Fatalf("esperava Payload normalizado para '{}', obteve '%s'", string(req.Payload))
	}

	// Serializa o envelope completo para JSON
	bytes, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("falha ao serializar envelope: %v", err)
	}

	if string(bytes) == "" || string(req.Payload) == "null" {
		t.Fatalf("payload no JSON não pode ser null")
	}
}

func TestProtocol_NewResponseEnvelope(t *testing.T) {
	respPayload := CreateChatResponse{
		Chat: ChatDTO{ID: "chat-1", Title: "Chat 1", Status: ChatActive},
	}
	resp, err := NewResponseEnvelope("req-uuid-123", MsgCreateChat, respPayload)
	if err != nil {
		t.Fatalf("falha ao criar ResponseEnvelope: %v", err)
	}

	if resp.Status != ErrSuccess {
		t.Errorf("esperava Status ErrSuccess (0), obteve %d", resp.Status)
	}
	if resp.Error != nil {
		t.Errorf("Error deveria ser nil em resposta de sucesso")
	}

	var parsed CreateChatResponse
	if err := UnmarshalPayload(resp.Payload, &parsed); err != nil {
		t.Fatalf("falha ao desserializar payload da resposta: %v", err)
	}
	if parsed.Chat.ID != "chat-1" || parsed.Chat.Title != "Chat 1" {
		t.Errorf("dados incorretos na resposta: %+v", parsed.Chat)
	}
}

func TestProtocol_NewErrorResponseEnvelope(t *testing.T) {
	resp := NewErrorResponseEnvelope("req-uuid-123", MsgGetChat, ErrNotFound, "Chat não encontrado", "chat_id=chat-999")

	if resp.Status != ErrNotFound {
		t.Errorf("esperava Status ErrNotFound (4004), obteve %d", resp.Status)
	}
	if resp.Error == nil {
		t.Fatalf("Error não deveria ser nil em envelope de erro")
	}
	if resp.Error.Message != "Chat não encontrado" || resp.Error.Details != "chat_id=chat-999" {
		t.Errorf("ErrorInfo incorreto: %+v", resp.Error)
	}
	if string(resp.Payload) != "{}" {
		t.Errorf("Payload em resposta de erro deve ser '{}', obteve '%s'", string(resp.Payload))
	}
}

func TestProtocol_MarshalUnmarshalPayload(t *testing.T) {
	original := CreateMessageRequest{
		ChatID:       "chat-123",
		AuthorType:   AuthorRegistration,
		AuthorID:     "cli",
		Role:         RoleUser,
		Content:      "Test message",
		TriggerAgent: true,
	}

	raw, err := MarshalPayload(original)
	if err != nil {
		t.Fatalf("falha no MarshalPayload: %v", err)
	}

	var restored CreateMessageRequest
	if err := UnmarshalPayload(raw, &restored); err != nil {
		t.Fatalf("falha no UnmarshalPayload: %v", err)
	}

	if restored.ChatID != original.ChatID || restored.Content != original.Content || restored.TriggerAgent != original.TriggerAgent {
		t.Errorf("falha de simetria no payload: %+v", restored)
	}
}

func TestProtocol_ValidateRequestEnvelope(t *testing.T) {
	valid := &RequestEnvelope{
		ProtocolVersion: ProtocolVersionCurrent,
		RequestID:       "req-1",
		ClientID:        "client-1",
		Type:            MsgCreateChat,
	}
	if err := ValidateRequestEnvelope(valid); err != nil {
		t.Fatalf("esperava requisição válida, obteve erro: %v", err)
	}

	// Versão incompatível
	invalidVer := *valid
	invalidVer.ProtocolVersion = 99
	if err := ValidateRequestEnvelope(&invalidVer); !errors.Is(err, ErrInvalidProtocolVersion) {
		t.Fatalf("esperava ErrInvalidProtocolVersion, obteve: %v", err)
	}

	// RequestID vazio
	noReqID := *valid
	noReqID.RequestID = ""
	if err := ValidateRequestEnvelope(&noReqID); !errors.Is(err, ErrMissingRequestID) {
		t.Fatalf("esperava ErrMissingRequestID, obteve: %v", err)
	}

	// ClientID vazio
	noClientID := *valid
	noClientID.ClientID = ""
	if err := ValidateRequestEnvelope(&noClientID); !errors.Is(err, ErrMissingClientID) {
		t.Fatalf("esperava ErrMissingClientID, obteve: %v", err)
	}

	// Type desconhecido
	badType := *valid
	badType.Type = 999
	if err := ValidateRequestEnvelope(&badType); !errors.Is(err, ErrUnknownMessageType) {
		t.Fatalf("esperava ErrUnknownMessageType, obteve: %v", err)
	}
}
