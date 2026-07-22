package commands

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	globalClientID string
)

var rootCmd = &cobra.Command{
	Use:   "jay",
	Short: "Jay AI Assistant CLI Client",
	Long:  `Jay CLI é a interface de linha de comando para interação e gerenciamento do assistente Jay via IPC Unix Socket.`,
}

// Execute executa o comando raiz do Cobra e trata erros de linha de comando.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Erro: %v\n", err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&globalClientID, "client-id", "c", "cli_user", "ID do cliente registrado no Jay Core")

	rootCmd.AddCommand(newConfigCmd())
	rootCmd.AddCommand(newRegisterCmd())
	rootCmd.AddCommand(newUnregisterCmd())
	rootCmd.AddCommand(newChatCmd())
	rootCmd.AddCommand(newMsgCmd())
	rootCmd.AddCommand(newProcessCmd())
	rootCmd.AddCommand(newToolCmd())
}
