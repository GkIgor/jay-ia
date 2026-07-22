package commands

import (
	"fmt"

	"github.com/GkIgor/jay-ia/sdk/ipc"
	"github.com/spf13/cobra"
)

func newToolCmd() *cobra.Command {
	toolCmd := &cobra.Command{
		Use:   "tool",
		Short: "Gerencia o catálogo de ferramentas e capacidades",
		Long:  `Subcomandos para cadastrar ferramentas nativas/externas e listar a matriz de capacidades do catálogo.`,
	}

	toolCmd.AddCommand(newToolRegisterCmd())
	toolCmd.AddCommand(newToolListCmd())

	return toolCmd
}

func newToolRegisterCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "register [id] [name] [description]",
		Short: "Registra ou atualiza uma ferramenta no catálogo (MsgRegisterTool - 400)",
		Args:  cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			id := args[0]
			name := args[1]
			desc := args[2]

			payload := ipc.RegisterToolRequest{
				ID:          id,
				Name:        name,
				Description: desc,
				Version:     "1.0.0",
				SchemaJSON:  "{}",
			}

			resp, err := sendRPC(ipc.MsgRegisterTool, globalClientID, payload)
			if err != nil {
				return err
			}

			if resp.Status != ipc.ErrSuccess {
				return fmt.Errorf("erro ao registrar ferramenta: %s", getErrorMessage(resp))
			}

			fmt.Printf("✓ Ferramenta '%s' (%s) registrada no catálogo com sucesso.\n", name, id)
			return nil
		},
	}
}

func newToolListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "Lista todas as ferramentas cadastradas no catálogo (MsgListTools - 403)",
		RunE: func(cmd *cobra.Command, args []string) error {
			payload := ipc.ListToolsRequest{}

			resp, err := sendRPC(ipc.MsgListTools, globalClientID, payload)
			if err != nil {
				return err
			}

			if resp.Status != ipc.ErrSuccess {
				return fmt.Errorf("erro ao listar ferramentas: %s", getErrorMessage(resp))
			}

			var listResp ipc.ListToolsResponse
			if err := ipc.UnmarshalPayload(resp.Payload, &listResp); err != nil {
				return err
			}

			if len(listResp.Tools) == 0 {
				fmt.Println("Nenhuma ferramenta cadastrada no catálogo.")
				return nil
			}

			fmt.Println("==========================================================")
			fmt.Printf("%-20s | %-20s | %-30s\n", "ID DA FERRAMENTA", "NOME", "DESCRIÇÃO")
			fmt.Println("==========================================================")
			for _, t := range listResp.Tools {
				fmt.Printf("%-20s | %-20s | %-30s\n", t.ID, t.Name, t.Description)
			}
			fmt.Println("==========================================================")
			return nil
		},
	}
}
