package api

import (
	"github.com/GkIgor/jay-ia/core/internal/storage"
	"github.com/GkIgor/jay-ia/sdk/ipc"
)

// toToolDTO converte uma entidade storage.Tool em DTO do SDK IPC.
func toToolDTO(tool *storage.Tool) ipc.ToolDTO {
	if tool == nil {
		return ipc.ToolDTO{}
	}
	return ipc.ToolDTO{
		ID:             tool.ID,
		RegistrationID: tool.RegistrationID,
		Name:           tool.Name,
		Description:    tool.Description,
		Version:        tool.Version,
		SchemaJSON:     tool.SchemaJSON,
		Status:         ipc.ToolStatus(tool.Status),
		CreatedAt:      tool.CreatedAt,
		UpdatedAt:      tool.UpdatedAt,
	}
}

// toToolDTOs converte uma lista de entidades storage.Tool em lista de DTOs IPC.
func toToolDTOs(tools []*storage.Tool) []ipc.ToolDTO {
	dtos := make([]ipc.ToolDTO, 0, len(tools))
	for _, t := range tools {
		dtos = append(dtos, toToolDTO(t))
	}
	return dtos
}
