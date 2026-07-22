package api

import (
	"github.com/GkIgor/jay-ia/core/internal/storage"
	"github.com/GkIgor/jay-ia/sdk/ipc"
)

// toProcessChatResponse constrói a resposta RPC ipc.ProcessChatResponse a partir da mensagem gerada pela IA.
func toProcessChatResponse(msg *storage.Message) ipc.ProcessChatResponse {
	if msg == nil {
		return ipc.ProcessChatResponse{}
	}
	return ipc.ProcessChatResponse{
		ProcessedMessage: toMessageDTO(msg),
	}
}
