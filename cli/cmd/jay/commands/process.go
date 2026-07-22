package commands

import (
	"fmt"

	"github.com/GkIgor/jay-ia/sdk/ipc"
	"github.com/spf13/cobra"
)

func newProcessCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "process [chat_id]",
		Short: "Disparo manual/administrativo de inferência de IA no chat (MsgProcessChat - 350)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			chatID := args[0]
			payload := ipc.ProcessChatRequest{ChatID: chatID}

			fmt.Print("Jay: pensando...")
			resp, err := sendRPC(ipc.MsgProcessChat, globalClientID, payload)
			fmt.Print("\r")

			if err != nil {
				return err
			}

			if resp.Status != ipc.ErrSuccess {
				return fmt.Errorf("erro no processamento da IA: %s", getErrorMessage(resp))
			}

			var procResp ipc.ProcessChatResponse
			if err := ipc.UnmarshalPayload(resp.Payload, &procResp); err != nil {
				return err
			}

			fmt.Println("==========================================================")
			fmt.Printf("RESPOSTA DA IA (Seq: %d):\n", procResp.ProcessedMessage.SequenceNo)
			fmt.Println("==========================================================")
			fmt.Println(procResp.ProcessedMessage.Content)
			fmt.Println("==========================================================")
			return nil
		},
	}
}
