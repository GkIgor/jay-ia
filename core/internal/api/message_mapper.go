package api

import (
	"github.com/GkIgor/jay-ia/core/internal/storage"
	"github.com/GkIgor/jay-ia/sdk/ipc"
)

// toMessageDTO converte uma entidade storage.Message em DTO do SDK IPC.
// Para mensagens com Soft Delete (status = 3), o conteúdo é ocultado por segurança (content = "").
func toMessageDTO(msg *storage.Message) ipc.MessageDTO {
	if msg == nil {
		return ipc.MessageDTO{}
	}

	content := msg.Content
	if msg.Status == storage.MessageDeleted {
		content = "" // Scrubbing de segurança em mensagens deletadas
	}

	return ipc.MessageDTO{
		ID:           msg.ID,
		ChatID:       msg.ChatID,
		AuthorType:   ipc.AuthorType(msg.AuthorType),
		AuthorID:     msg.AuthorID,
		Role:         ipc.MessageRole(msg.Role),
		Content:      content,
		ContentType:  ipc.MessageContentType(msg.ContentType),
		Status:       ipc.MessageStatus(msg.Status),
		SequenceNo:   msg.SequenceNo,
		MetadataJSON: msg.MetadataJSON,
		CreatedAt:    msg.CreatedAt,
		UpdatedAt:    msg.UpdatedAt,
	}
}

// toMessageDTOs converte uma lista de entidades storage.Message em lista de DTOs IPC.
func toMessageDTOs(msgs []*storage.Message) []ipc.MessageDTO {
	dtos := make([]ipc.MessageDTO, 0, len(msgs))
	for _, msg := range msgs {
		dtos = append(dtos, toMessageDTO(msg))
	}
	return dtos
}
