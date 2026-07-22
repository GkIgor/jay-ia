package commands

import (
	"fmt"

	"github.com/GkIgor/jay-ia/sdk/ipc"
	"github.com/spf13/cobra"
)

func newRegisterCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "register [client_id]",
		Short: "Registra a aplicação cliente no Jay Core (MsgRegisterClient - 100)",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientID := globalClientID
			if len(args) > 0 {
				clientID = args[0]
			}

			payload := ipc.RegisterClientRequest{ClientID: clientID}
			resp, err := sendRPC(ipc.MsgRegisterClient, clientID, payload)
			if err != nil {
				return err
			}

			if resp.Status != ipc.ErrSuccess {
				return fmt.Errorf("erro no registro: %s", getErrorMessage(resp))
			}

			var regResp ipc.RegisterClientResponse
			if err := ipc.UnmarshalPayload(resp.Payload, &regResp); err != nil {
				return err
			}

			fmt.Printf("✓ Cliente '%s' registrado com sucesso! (ID: %s)\n", regResp.Registration.ID, regResp.Registration.ID)
			return nil
		},
	}
}

func newUnregisterCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "unregister [client_id]",
		Short: "Cancela o registro da aplicação cliente (MsgUnregisterClient - 101)",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientID := globalClientID
			if len(args) > 0 {
				clientID = args[0]
			}

			payload := ipc.UnregisterClientRequest{ClientID: clientID}
			resp, err := sendRPC(ipc.MsgUnregisterClient, clientID, payload)
			if err != nil {
				return err
			}

			if resp.Status != ipc.ErrSuccess {
				return fmt.Errorf("erro ao cancelar registro: %s", getErrorMessage(resp))
			}

			fmt.Printf("✓ Registro do cliente '%s' cancelado com sucesso.\n", clientID)
			return nil
		},
	}
}
