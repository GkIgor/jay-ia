package commands

import (
	"bufio"
	"fmt"
	"os"

	"github.com/GkIgor/jay-ia/sdk/ipc"
	"github.com/spf13/cobra"
)

func newChatReplCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "repl [chat_id]",
		Aliases: []string{"interactive"},
		Short:   "Abre o chat interativo conversacional em tempo real (REPL)",
		Long:    `Inicia um loop conversacional com a IA Jay no terminal. Suporta os comandos /help, /clear, /new, /quit e /exit.`,
		Args:    cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			chatID := ""
			if len(args) > 0 {
				chatID = args[0]
			} else {
				// Cria um novo chat para a sessão REPL
				payload := ipc.CreateChatRequest{Title: "Conversa Interativa CLI"}
				resp, err := sendRPC(ipc.MsgCreateChat, globalClientID, payload)
				if err != nil {
					return fmt.Errorf("falha ao criar chat para o REPL: %w", err)
				}
				if resp.Status != ipc.ErrSuccess {
					return fmt.Errorf("erro ao criar chat REPL: %s", getErrorMessage(resp))
				}
				var createResp ipc.CreateChatResponse
				if err := ipc.UnmarshalPayload(resp.Payload, &createResp); err != nil {
					return err
				}
				chatID = createResp.Chat.ID
			}

			fmt.Println("==========================================================")
			fmt.Println("             🤖 Jay AI Assistant — Chat REPL")
			fmt.Println("==========================================================")
			fmt.Printf(" Chat ID: %s\n", chatID)
			fmt.Println(" Atalhos: /help (ajuda), /clear (limpar tela), /new (novo chat), /quit (sair)")
			fmt.Println("==========================================================")

			scanner := bufio.NewScanner(os.Stdin)

			for {
				fmt.Print("\nVocê > ")
				if !scanner.Scan() {
					break
				}
				input := scanner.Text()
				if input == "" {
					continue
				}

				// Tratamento de comandos internos do REPL
				switch input {
				case "/quit", "/exit":
					fmt.Println("Até logo!")
					return nil
				case "/clear":
					fmt.Print("\033[H\033[2J")
					fmt.Printf("Jay REPL (Chat ID: %s)\n", chatID)
					continue
				case "/help":
					fmt.Println("Comandos do REPL:")
					fmt.Println("  /help  - Exibe esta mensagem de ajuda")
					fmt.Println("  /clear - Limpa a tela do terminal")
					fmt.Println("  /new   - Inicia uma nova conversa")
					fmt.Println("  /quit  - Encerra o REPL")
					continue
				case "/new":
					payload := ipc.CreateChatRequest{Title: "Nova Conversa REPL"}
					resp, err := sendRPC(ipc.MsgCreateChat, globalClientID, payload)
					if err == nil && resp.Status == ipc.ErrSuccess {
						var createResp ipc.CreateChatResponse
						if err := ipc.UnmarshalPayload(resp.Payload, &createResp); err == nil {
							chatID = createResp.Chat.ID
							fmt.Printf("✓ Nova conversa iniciada! (ID: %s)\n", chatID)
						}
					}
					continue
				}

				// 1. Registra a mensagem do usuário no chat
				msgPayload := ipc.CreateMessageRequest{
					ChatID:       chatID,
					Content:      input,
					TriggerAgent: true,
				}
				msgResp, err := sendRPC(ipc.MsgCreateMessage, globalClientID, msgPayload)
				if err != nil {
					fmt.Printf("❌ Erro ao enviar mensagem: %v\n", err)
					continue
				}
				if msgResp.Status != ipc.ErrSuccess {
					fmt.Printf("❌ Erro ao registrar mensagem: %s\n", getErrorMessage(msgResp))
					continue
				}

				// 2. Feedback Visual e acionamento da inferência de IA
				fmt.Print("Jay  > pensando...")

				procPayload := ipc.ProcessChatRequest{ChatID: chatID}
				procResp, err := sendRPC(ipc.MsgProcessChat, globalClientID, procPayload)

				// Limpa o indicador "pensando..."
				fmt.Print("\rJay  > ")

				if err != nil {
					fmt.Printf("❌ Erro no processamento da IA: %v\n", err)
					continue
				}
				if procResp.Status != ipc.ErrSuccess {
					fmt.Printf("❌ Erro da IA: %s\n", getErrorMessage(procResp))
					continue
				}

				var processOutput ipc.ProcessChatResponse
				if err := ipc.UnmarshalPayload(procResp.Payload, &processOutput); err != nil {
					fmt.Printf("❌ Falha ao decodificar resposta: %v\n", err)
					continue
				}

				fmt.Println(processOutput.ProcessedMessage.Content)
			}

			return nil
		},
	}
}
