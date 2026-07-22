package commands

import (
	"fmt"

	"github.com/GkIgor/jay-ia/sdk/ipc"
	"github.com/spf13/cobra"
)

func newMsgCmd() *cobra.Command {
	msgCmd := &cobra.Command{
		Use:   "msg",
		Short: "Gerencia mensagens individuais de conversas",
		Long:  `Subcomandos para enviar mensagens a um chat e consultar o histórico (modelo PULL).`,
	}

	msgCmd.AddCommand(newMsgSendCmd())
	msgCmd.AddCommand(newMsgListCmd())

	return msgCmd
}

func newMsgSendCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "send [chat_id] [content]",
		Short: "Envia uma mensagem para o chat (MsgCreateMessage - 300)",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			chatID := args[0]
			content := args[1]

			payload := ipc.CreateMessageRequest{
				ChatID:       chatID,
				Content:      content,
				TriggerAgent: false,
			}

			resp, err := sendRPC(ipc.MsgCreateMessage, globalClientID, payload)
			if err != nil {
				return err
			}

			if resp.Status != ipc.ErrSuccess {
				return fmt.Errorf("erro ao enviar mensagem: %s", getErrorMessage(resp))
			}

			var createResp ipc.CreateMessageResponse
			if err := ipc.UnmarshalPayload(resp.Payload, &createResp); err != nil {
				return err
			}

			fmt.Printf("✓ Mensagem enviada com sucesso! (ID: %s, Seq: %d)\n", createResp.CreatedMessage.ID, createResp.CreatedMessage.SequenceNo)
			return nil
		},
	}
}

func newMsgListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list [chat_id]",
		Short: "Lista o histórico de mensagens do chat em modelo PULL (MsgGetMessages - 303)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			chatID := args[0]
			payload := ipc.GetMessagesRequest{ChatID: chatID}

			resp, err := sendRPC(ipc.MsgGetMessages, globalClientID, payload)
			if err != nil {
				return err
			}

			if resp.Status != ipc.ErrSuccess {
				return fmt.Errorf("erro ao obter histórico: %s", getErrorMessage(resp))
			}

			var getResp ipc.GetMessagesResponse
			if err := ipc.UnmarshalPayload(resp.Payload, &getResp); err != nil {
				return err
			}

			if len(getResp.Messages) == 0 {
				fmt.Println("Nenhuma mensagem encontrada neste chat.")
				return nil
			}

			fmt.Println("==========================================================")
			fmt.Printf("HISTÓRICO DE MENSAGENS (CHAT: %s)\n", chatID)
			fmt.Println("==========================================================")
			for _, m := range getResp.Messages {
				author := "Usuário"
				if m.AuthorType == ipc.AuthorAgent {
					author = "Jay (IA)"
				}
				fmt.Printf("[%d] %s: %s\n", m.SequenceNo, author, m.Content)
			}
			fmt.Println("==========================================================")
			return nil
		},
	}
}
