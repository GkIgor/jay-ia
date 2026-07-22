package commands

import (
	"fmt"

	"github.com/GkIgor/jay-ia/sdk/ipc"
	"github.com/spf13/cobra"
)

func newChatCmd() *cobra.Command {
	chatCmd := &cobra.Command{
		Use:   "chat",
		Short: "Gerencia sessões de conversa e abre o chat interativo",
		Long:  `Subcomandos para criar, listar, renomear, deletar conversas e abrir o terminal interativo REPL.`,
	}

	chatCmd.AddCommand(newChatCreateCmd())
	chatCmd.AddCommand(newChatListCmd())
	chatCmd.AddCommand(newChatRenameCmd())
	chatCmd.AddCommand(newChatDeleteCmd())
	chatCmd.AddCommand(newChatReplCmd())

	return chatCmd
}

func newChatCreateCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "create [title]",
		Short: "Cria uma nova conversa (MsgCreateChat - 200)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			title := args[0]
			payload := ipc.CreateChatRequest{Title: title}

			resp, err := sendRPC(ipc.MsgCreateChat, globalClientID, payload)
			if err != nil {
				return err
			}

			if resp.Status != ipc.ErrSuccess {
				return fmt.Errorf("erro ao criar chat: %s", getErrorMessage(resp))
			}

			var createResp ipc.CreateChatResponse
			if err := ipc.UnmarshalPayload(resp.Payload, &createResp); err != nil {
				return err
			}

			fmt.Printf("✓ Chat '%s' criado com sucesso!\nID: %s\n", createResp.Chat.Title, createResp.Chat.ID)
			return nil
		},
	}
}

func newChatListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "Lista todas as conversas do cliente (MsgListChats - 204)",
		RunE: func(cmd *cobra.Command, args []string) error {
			payload := ipc.ListChatsRequest{IncludeShared: true}

			resp, err := sendRPC(ipc.MsgListChats, globalClientID, payload)
			if err != nil {
				return err
			}

			if resp.Status != ipc.ErrSuccess {
				return fmt.Errorf("erro ao listar chats: %s", getErrorMessage(resp))
			}

			var listResp ipc.ListChatsResponse
			if err := ipc.UnmarshalPayload(resp.Payload, &listResp); err != nil {
				return err
			}

			if len(listResp.Chats) == 0 {
				fmt.Println("Nenhum chat encontrado.")
				return nil
			}

			fmt.Println("==========================================================")
			fmt.Printf("%-38s | %-25s | %-10s\n", "ID DO CHAT", "TÍTULO", "STATUS")
			fmt.Println("==========================================================")
			for _, c := range listResp.Chats {
				statusStr := "Ativo"
				if c.Status == ipc.ChatArchived {
					statusStr = "Arquivado"
				}
				fmt.Printf("%-38s | %-25s | %-10s\n", c.ID, c.Title, statusStr)
			}
			fmt.Println("==========================================================")
			return nil
		},
	}
}

func newChatRenameCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "rename [chat_id] [new_title]",
		Short: "Altera o título de uma conversa (MsgRenameChat - 202)",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			chatID := args[0]
			newTitle := args[1]

			payload := ipc.RenameChatRequest{ChatID: chatID, NewTitle: newTitle}
			resp, err := sendRPC(ipc.MsgRenameChat, globalClientID, payload)
			if err != nil {
				return err
			}

			if resp.Status != ipc.ErrSuccess {
				return fmt.Errorf("erro ao renomear chat: %s", getErrorMessage(resp))
			}

			fmt.Printf("✓ Chat '%s' renomeado para '%s' com sucesso.\n", chatID, newTitle)
			return nil
		},
	}
}

func newChatDeleteCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "delete [chat_id]",
		Short: "Deleta logicamente uma conversa (MsgDeleteChat - 201)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			chatID := args[0]

			payload := ipc.DeleteChatRequest{ChatID: chatID}
			resp, err := sendRPC(ipc.MsgDeleteChat, globalClientID, payload)
			if err != nil {
				return err
			}

			if resp.Status != ipc.ErrSuccess {
				return fmt.Errorf("erro ao deletar chat: %s", getErrorMessage(resp))
			}

			fmt.Printf("✓ Chat '%s' removido com sucesso.\n", chatID)
			return nil
		},
	}
}
