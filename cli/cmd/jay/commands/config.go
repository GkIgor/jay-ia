package commands

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/GkIgor/jay-ia/sdk/ipc"
	"github.com/spf13/cobra"
)

func newConfigCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "config",
		Aliases: []string{"init"},
		Short:   "Assistente interativo de configuração do ambiente (~/.jay/.env)",
		Long:    `Executa um guia interativo para configurar o provedor de LLM, chave de API e validar a conexão com o daemon Jay.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			reader := bufio.NewReader(os.Stdin)

			fmt.Println("=====================================================")
			fmt.Println("       🤖 Assistente de Configuração — Jay AI")
			fmt.Println("=====================================================")

			fmt.Print("\n1. Selecione o Provedor de LLM (mock / openrouter / gemini) [padrão: mock]: ")
			providerInput, _ := reader.ReadString('\n')
			provider := strings.TrimSpace(providerInput)
			if provider == "" {
				provider = "mock"
			}

			apiKey := ""
			if provider == "openrouter" || provider == "gemini" {
				fmt.Printf("2. Digite a Chave de API para %s: ", provider)
				keyInput, _ := reader.ReadString('\n')
				apiKey = strings.TrimSpace(keyInput)
			}

			fmt.Print("3. Digite o ID do Cliente de Registro [padrão: cli_user]: ")
			clientInput, _ := reader.ReadString('\n')
			clientID := strings.TrimSpace(clientInput)
			if clientID == "" {
				clientID = "cli_user"
			}

			// Escreve em ~/.jay/.env ou XDG_CONFIG_HOME
			homeDir, err := os.UserHomeDir()
			if err != nil {
				homeDir = "."
			}
			configDir := filepath.Join(homeDir, ".jay")
			_ = os.MkdirAll(configDir, 0700)
			envPath := filepath.Join(configDir, ".env")

			envContent := fmt.Sprintf("LLM_PROVIDER=%s\n", provider)
			if provider == "openrouter" {
				envContent += fmt.Sprintf("OPENROUTER_API_KEY=%s\n", apiKey)
			} else if provider == "gemini" {
				envContent += fmt.Sprintf("GEMINI_API_KEY=%s\n", apiKey)
			}
			envContent += fmt.Sprintf("JAY_CLIENT_ID=%s\n", clientID)

			if err := os.WriteFile(envPath, []byte(envContent), 0600); err != nil {
				return fmt.Errorf("falha ao salvar arquivo de configuração %s: %w", envPath, err)
			}

			fmt.Printf("\n✓ Configurações salvas com sucesso em %s!\n", envPath)
			fmt.Println("Validando conexão IPC com o daemon 'jayd'...")

			resp, err := sendRPC(ipc.MsgRegisterClient, clientID, ipc.RegisterClientRequest{ClientID: clientID})
			if err != nil {
				fmt.Printf("⚠️  Aviso: Não foi possível conectar ao daemon 'jayd': %v\n", err)
				fmt.Println("Dica: Certifique-se de que o daemon 'jayd' esteja rodando antes de enviar comandos.")
			} else if resp.Status == ipc.ErrSuccess {
				fmt.Println("✓ Conexão IPC estabelecida e cliente registrado com sucesso no Jay Core!")
			} else {
				fmt.Printf("⚠️  Aviso ao registrar cliente: %s\n", getErrorMessage(resp))
			}

			return nil
		},
	}
	return cmd
}
