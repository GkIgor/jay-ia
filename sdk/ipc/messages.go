package ipc

// DTOs de Domínio (Transferência de Recursos)

type RegistrationDTO struct {
	ID           string `json:"id"`
	Status       int    `json:"status"`
	MetadataJSON string `json:"metadata_json,omitempty"`
	CreatedAt    string `json:"created_at,omitempty"`
	UpdatedAt    string `json:"updated_at,omitempty"`
}

type ChatDTO struct {
	ID                  string     `json:"id"`
	OwnerRegistrationID string     `json:"owner_registration_id"`
	Title               string     `json:"title"`
	Status              ChatStatus `json:"status"`
	IsOwner             bool       `json:"is_owner,omitempty"`
	MetadataJSON        string     `json:"metadata_json,omitempty"`
	CreatedAt           string     `json:"created_at,omitempty"`
	UpdatedAt           string     `json:"updated_at,omitempty"`
}

type MessageDTO struct {
	ID           string             `json:"id"`
	ChatID       string             `json:"chat_id"`
	AuthorType   AuthorType         `json:"author_type"`
	AuthorID     string             `json:"author_id"`
	Role         MessageRole        `json:"role"`
	Content      string             `json:"content"`
	ContentType  MessageContentType `json:"content_type"`
	Status       MessageStatus      `json:"status"`
	SequenceNo   int                `json:"sequence_no"`
	MetadataJSON string             `json:"metadata_json,omitempty"`
	CreatedAt    string             `json:"created_at,omitempty"`
	UpdatedAt    string             `json:"updated_at,omitempty"`
}

type ToolDTO struct {
	ID             string     `json:"id"`
	RegistrationID string     `json:"registration_id"`
	Name           string     `json:"name"`
	Description    string     `json:"description"`
	Version        string     `json:"version"`
	SchemaJSON     string     `json:"schema_json,omitempty"`
	Status         ToolStatus `json:"status"`
	CreatedAt      string     `json:"created_at,omitempty"`
	UpdatedAt      string     `json:"updated_at,omitempty"`
}

// Payloads Específicos por Comando

// --- Registros ---

type RegisterClientRequest struct {
	ClientID string `json:"client_id"`
	Metadata string `json:"metadata,omitempty"`
}

type RegisterClientResponse struct {
	Registration RegistrationDTO `json:"registration"`
}

type UnregisterClientRequest struct {
	ClientID string `json:"client_id"`
}

type UnregisterClientResponse struct {
	Success bool `json:"success"`
}

type UpdateRegistrationRequest struct {
	ClientID string `json:"client_id"`
	Status   int    `json:"status,omitempty"`
	Metadata string `json:"metadata,omitempty"`
}

type UpdateRegistrationResponse struct {
	Registration RegistrationDTO `json:"registration"`
}

type GetRegistrationRequest struct {
	RegistrationID string `json:"registration_id"`
}

type GetRegistrationResponse struct {
	Registration RegistrationDTO `json:"registration"`
}

type ListRegistrationsRequest struct {
	Limit  int `json:"limit,omitempty"`
	Offset int `json:"offset,omitempty"`
}

type ListRegistrationsResponse struct {
	Registrations []RegistrationDTO `json:"registrations"`
	Total         int               `json:"total"`
}

type SharedRulePayload struct {
	TargetScope    int    `json:"target_scope"`
	Pattern        string `json:"pattern"`
	MatchType      int    `json:"match_type"`
	AllowedActions int    `json:"allowed_actions"`
}

type UpdateSharedRulesRequest struct {
	Rules []SharedRulePayload `json:"rules"`
}

type UpdateSharedRulesResponse struct {
	AppliedRulesCount int `json:"applied_rules_count"`
}

// --- Chats ---

type CreateChatRequest struct {
	Title    string `json:"title"`
	Metadata string `json:"metadata,omitempty"`
}

type CreateChatResponse struct {
	Chat ChatDTO `json:"chat"`
}

type ListChatsRequest struct {
	IncludeShared bool `json:"include_shared,omitempty"`
	Limit         int  `json:"limit,omitempty"`
}

type ListChatsResponse struct {
	Chats []ChatDTO `json:"chats"`
	Total int       `json:"total"`
}

// --- Mensagens & Processamento ---

type CreateMessageRequest struct {
	ChatID       string             `json:"chat_id"`
	AuthorType   AuthorType         `json:"author_type"`
	AuthorID     string             `json:"author_id"`
	Role         MessageRole        `json:"role"`
	Content      string             `json:"content"`
	ContentType  MessageContentType `json:"content_type,omitempty"`
	TriggerAgent bool               `json:"trigger_agent,omitempty"`
	Metadata     string             `json:"metadata,omitempty"`
}

type CreateMessageResponse struct {
	CreatedMessage   MessageDTO  `json:"created_message"`
	ProcessedMessage *MessageDTO `json:"processed_message,omitempty"`
}

type ProcessChatRequest struct {
	ChatID string `json:"chat_id"`
}

type ProcessChatResponse struct {
	ProcessedMessage MessageDTO `json:"processed_message"`
}

type GetMessagesRequest struct {
	ChatID          string `json:"chat_id"`
	SinceSequenceNo int    `json:"since_sequence_no,omitempty"`
	Limit           int    `json:"limit,omitempty"`
}

type GetMessagesResponse struct {
	ChatID   string       `json:"chat_id"`
	Messages []MessageDTO `json:"messages"`
	HasMore  bool         `json:"has_more"`
}

// --- Ferramentas ---

type RegisterToolRequest struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Version     string `json:"version,omitempty"`
	SchemaJSON  string `json:"schema_json,omitempty"`
}

type RegisterToolResponse struct {
	Tool ToolDTO `json:"tool"`
}

type ListToolsResponse struct {
	Tools []ToolDTO `json:"tools"`
}

// Estruturas genéricas legadas mantidas para retrocompatibilidade simples

type Message struct {
	Type    string `json:"type"`
	Payload any    `json:"payload"`
}

type Command struct {
	ID     string `json:"id"`
	Action string `json:"action"`
	Data   any    `json:"data,omitempty"`
}

type Response struct {
	RefID  string `json:"ref_id"`
	Status string `json:"status"`
	Data   any    `json:"data,omitempty"`
}

type Event struct {
	Topic string `json:"topic"`
	Data  any    `json:"data,omitempty"`
}

type Error struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}
