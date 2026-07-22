package api

import (
	"github.com/GkIgor/jay-ia/core/internal/storage"
	"github.com/GkIgor/jay-ia/sdk/ipc"
)

// toChatDTO converte uma entidade storage.Chat em DTO do SDK IPC.
func toChatDTO(chat *storage.Chat, requesterID string) ipc.ChatDTO {
	if chat == nil {
		return ipc.ChatDTO{}
	}
	return ipc.ChatDTO{
		ID:                  chat.ID,
		OwnerRegistrationID: chat.OwnerRegistrationID,
		Title:               chat.Title,
		Status:              ipc.ChatStatus(chat.Status),
		IsOwner:             chat.OwnerRegistrationID == requesterID,
		MetadataJSON:        chat.MetadataJSON,
		CreatedAt:           chat.CreatedAt,
		UpdatedAt:           chat.UpdatedAt,
	}
}

// toChatDTOs converte uma lista de entidades storage.Chat em lista de DTOs IPC.
func toChatDTOs(chats []*storage.Chat, requesterID string) []ipc.ChatDTO {
	dtos := make([]ipc.ChatDTO, 0, len(chats))
	for _, chat := range chats {
		dtos = append(dtos, toChatDTO(chat, requesterID))
	}
	return dtos
}
