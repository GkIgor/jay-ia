package api

import (
	"github.com/GkIgor/jay-ia/core/internal/storage"
	"github.com/GkIgor/jay-ia/sdk/ipc"
)

// toRegistrationDTO converte uma entidade storage.Registration em DTO do SDK IPC.
func toRegistrationDTO(reg *storage.Registration) ipc.RegistrationDTO {
	if reg == nil {
		return ipc.RegistrationDTO{}
	}
	return ipc.RegistrationDTO{
		ID:           reg.ID,
		Status:       int(reg.Status),
		MetadataJSON: reg.MetadataJSON,
		CreatedAt:    reg.CreatedAt,
		UpdatedAt:    reg.UpdatedAt,
	}
}

// toRegistrationDTOs converte uma lista de entidades storage.Registration em lista de DTOs IPC.
func toRegistrationDTOs(regs []*storage.Registration) []ipc.RegistrationDTO {
	dtos := make([]ipc.RegistrationDTO, 0, len(regs))
	for _, reg := range regs {
		dtos = append(dtos, toRegistrationDTO(reg))
	}
	return dtos
}
