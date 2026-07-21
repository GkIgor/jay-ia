package ipc

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

// ProtocolVersion representa o tipo numérico forte da versão do protocolo IPC.
type ProtocolVersion uint8

const ProtocolVersionCurrent ProtocolVersion = 1

// MessageType representa o tipo numérico forte da ação/comando do protocolo IPC.
type MessageType uint16

const (
	// Registros de Identidades Lógicas (100-199)
	MsgRegisterClient     MessageType = 100
	MsgUnregisterClient   MessageType = 101
	MsgUpdateRegistration MessageType = 102
	MsgGetRegistration    MessageType = 103
	MsgListRegistrations  MessageType = 104
	MsgUpdateSharedRules  MessageType = 105

	// Gerenciamento de Chats (200-299)
	MsgCreateChat MessageType = 200
	MsgDeleteChat MessageType = 201
	MsgRenameChat MessageType = 202
	MsgGetChat    MessageType = 203
	MsgListChats  MessageType = 204

	// Gerenciamento de Mensagens (300-349)
	MsgCreateMessage MessageType = 300
	MsgUpdateMessage MessageType = 301
	MsgDeleteMessage MessageType = 302
	MsgGetMessages   MessageType = 303

	// Processamento de Conversa com IA (350-399)
	MsgProcessChat MessageType = 350

	// Catálogo de Ferramentas (400-499)
	MsgRegisterTool   MessageType = 400
	MsgUnregisterTool MessageType = 401
	MsgGetTool        MessageType = 402
	MsgListTools      MessageType = 403

	// Sessões de Voz (500-599)
	MsgCreateVoiceSession MessageType = 500
	MsgGetVoiceSession    MessageType = 501
	MsgCloseVoiceSession  MessageType = 502
)

// ErrorCode representa os códigos de status e erro padronizados do protocolo IPC.
type ErrorCode uint16

const (
	ErrSuccess          ErrorCode = 0
	ErrInvalidFormat    ErrorCode = 4000
	ErrUnauthorized     ErrorCode = 4001
	ErrForbidden        ErrorCode = 4003
	ErrNotFound         ErrorCode = 4004
	ErrConflict         ErrorCode = 4009
	ErrInternalDatabase ErrorCode = 5000
	ErrNotImplemented   ErrorCode = 5001
)

// Enums próprios do SDK para tipagem estrita de DTOs sem dependência de pacotes internos.
type AuthorType uint8

const (
	AuthorRegistration AuthorType = 1
	AuthorAgent        AuthorType = 2
	AuthorTool         AuthorType = 3
	AuthorSystem       AuthorType = 4
)

type MessageRole uint8

const (
	RoleUser      MessageRole = 1
	RoleAssistant MessageRole = 2
	RoleSystem    MessageRole = 3
	RoleTool      MessageRole = 4
)

type MessageContentType uint8

const (
	ContentTypeTextPlain  MessageContentType = 1
	ContentTypeMarkdown   MessageContentType = 2
	ContentTypeJSON       MessageContentType = 3
	ContentTypeToolCall   MessageContentType = 4
	ContentTypeToolResult MessageContentType = 5
)

type MessageStatus uint8

const (
	MessageSent    MessageStatus = 1
	MessageEdited  MessageStatus = 2
	MessageDeleted MessageStatus = 3
)

type ChatStatus uint8

const (
	ChatActive   ChatStatus = 1
	ChatArchived ChatStatus = 2
	ChatDeleted  ChatStatus = 3
)

type ToolStatus uint8

const (
	ToolAvailable  ToolStatus = 1
	ToolDisabled   ToolStatus = 2
	ToolDeprecated ToolStatus = 3
)

// RequestEnvelope representa o envelope padronizado de uma solicitação enviada de um cliente para o Jay Core.
type RequestEnvelope struct {
	ProtocolVersion ProtocolVersion `json:"protocol_version"`
	RequestID       string          `json:"request_id"`
	ClientID        string          `json:"client_id"`
	Type            MessageType     `json:"type"`
	Payload         json.RawMessage `json:"payload"`
}

// ResponseEnvelope representa o envelope padronizado de resposta enviado pelo Jay Core.
type ResponseEnvelope struct {
	ProtocolVersion ProtocolVersion `json:"protocol_version"`
	RequestID       string          `json:"request_id"`
	Type            MessageType     `json:"type"`
	Status          ErrorCode       `json:"status"`
	Error           *ErrorInfo      `json:"error,omitempty"`
	Payload         json.RawMessage `json:"payload,omitempty"`
}

// ErrorInfo contém mensagem e detalhes da falha (sem duplicação do código de erro).
type ErrorInfo struct {
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
}

var (
	ErrInvalidProtocolVersion = errors.New("ipc: versão do protocolo não suportada")
	ErrMissingRequestID       = errors.New("ipc: request_id é obrigatório")
	ErrMissingClientID        = errors.New("ipc: client_id é obrigatório")
	ErrUnknownMessageType     = errors.New("ipc: tipo de mensagem desconhecido")
)

// MarshalPayload converte qualquer struct de payload em json.RawMessage.
// Se payloadStruct for nil ou zero, retorna json.RawMessage("{}") para garantir que o payload nunca seja null.
func MarshalPayload(payloadStruct any) (json.RawMessage, error) {
	if payloadStruct == nil {
		return json.RawMessage("{}"), nil
	}

	bytes, err := json.Marshal(payloadStruct)
	if err != nil {
		return nil, fmt.Errorf("ipc: falha ao serializar payload: %w", err)
	}

	if string(bytes) == "null" {
		return json.RawMessage("{}"), nil
	}

	return json.RawMessage(bytes), nil
}

// UnmarshalPayload desserializa um json.RawMessage na struct de destino fornecida.
func UnmarshalPayload(raw json.RawMessage, targetStruct any) error {
	if len(raw) == 0 || string(raw) == "{}" || string(raw) == "null" {
		return nil
	}
	if err := json.Unmarshal(raw, targetStruct); err != nil {
		return fmt.Errorf("ipc: falha ao desserializar payload: %w", err)
	}
	return nil
}

// NewRequestEnvelope constrói um RequestEnvelope com a versão atual do protocolo (1), gerando um UUID v4 no RequestID.
func NewRequestEnvelope(msgType MessageType, clientID string, payloadStruct any) (*RequestEnvelope, error) {
	cleanClientID := strings.TrimSpace(clientID)
	if cleanClientID == "" {
		return nil, ErrMissingClientID
	}

	payloadRaw, err := MarshalPayload(payloadStruct)
	if err != nil {
		return nil, err
	}

	reqID := generateUUIDv4()

	return &RequestEnvelope{
		ProtocolVersion: ProtocolVersionCurrent,
		RequestID:       reqID,
		ClientID:        cleanClientID,
		Type:            msgType,
		Payload:         payloadRaw,
	}, nil
}

// NewResponseEnvelope constrói um ResponseEnvelope de sucesso (Status = ErrSuccess).
func NewResponseEnvelope(requestID string, msgType MessageType, payloadStruct any) (*ResponseEnvelope, error) {
	cleanReqID := strings.TrimSpace(requestID)
	if cleanReqID == "" {
		return nil, ErrMissingRequestID
	}

	payloadRaw, err := MarshalPayload(payloadStruct)
	if err != nil {
		return nil, err
	}

	return &ResponseEnvelope{
		ProtocolVersion: ProtocolVersionCurrent,
		RequestID:       cleanReqID,
		Type:            msgType,
		Status:          ErrSuccess,
		Payload:         payloadRaw,
	}, nil
}

// NewErrorResponseEnvelope constrói um ResponseEnvelope de erro.
func NewErrorResponseEnvelope(requestID string, msgType MessageType, errCode ErrorCode, message string, details string) *ResponseEnvelope {
	return &ResponseEnvelope{
		ProtocolVersion: ProtocolVersionCurrent,
		RequestID:       strings.TrimSpace(requestID),
		Type:            msgType,
		Status:          errCode,
		Error: &ErrorInfo{
			Message: message,
			Details: details,
		},
		Payload: json.RawMessage("{}"),
	}
}

// ValidateRequestEnvelope realiza validações de formato no envelope de entrada.
func ValidateRequestEnvelope(req *RequestEnvelope) error {
	if req == nil {
		return errors.New("ipc: envelope nulo")
	}
	if req.ProtocolVersion > ProtocolVersionCurrent || req.ProtocolVersion == 0 {
		return ErrInvalidProtocolVersion
	}
	if strings.TrimSpace(req.RequestID) == "" {
		return ErrMissingRequestID
	}
	if strings.TrimSpace(req.ClientID) == "" {
		return ErrMissingClientID
	}
	if req.Type < 100 || req.Type > 599 {
		return ErrUnknownMessageType
	}
	return nil
}

// generateUUIDv4 gera um identificador único de formato UUID v4.
func generateUUIDv4() string {
	var uuid [16]byte
	_, err := rand.Read(uuid[:])
	if err != nil {
		// Fallback pseudo-randômico extremamente seguro se crypto/rand falhar
		return fmt.Sprintf("%x-%x-%x-%x-%x", uuid[0:4], uuid[4:6], uuid[6:8], uuid[8:10], uuid[10:16])
	}
	uuid[6] = (uuid[6] & 0x0f) | 0x40 // RFC 4122 versão 4
	uuid[8] = (uuid[8] & 0x3f) | 0x80 // RFC 4122 variante
	buf := make([]byte, 36)
	hex.Encode(buf[0:8], uuid[0:4])
	buf[8] = '-'
	hex.Encode(buf[9:13], uuid[4:6])
	buf[13] = '-'
	hex.Encode(buf[14:18], uuid[6:8])
	buf[18] = '-'
	hex.Encode(buf[19:23], uuid[8:10])
	buf[23] = '-'
	hex.Encode(buf[24:36], uuid[10:16])
	return string(buf)
}
