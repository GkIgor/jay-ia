package commands

import (
	"testing"
)

func TestCobra_RootCommand(t *testing.T) {
	if rootCmd == nil {
		t.Fatalf("esperava rootCmd instanciado no pacote commands")
	}

	expectedSubcommands := map[string]bool{
		"config":     false,
		"register":   false,
		"unregister": false,
		"chat":       false,
		"msg":        false,
		"process":    false,
		"tool":       false,
	}

	for _, cmd := range rootCmd.Commands() {
		if _, ok := expectedSubcommands[cmd.Name()]; ok {
			expectedSubcommands[cmd.Name()] = true
		}
	}

	for name, found := range expectedSubcommands {
		if !found {
			t.Errorf("subcomando '%s' não encontrado no rootCmd", name)
		}
	}
}

func TestCobra_ChatSubcommands(t *testing.T) {
	chatCmd := newChatCmd()

	expectedSubcommands := map[string]bool{
		"create": false,
		"list":   false,
		"rename": false,
		"delete": false,
		"repl":   false,
	}

	for _, cmd := range chatCmd.Commands() {
		if _, ok := expectedSubcommands[cmd.Name()]; ok {
			expectedSubcommands[cmd.Name()] = true
		}
	}

	for name, found := range expectedSubcommands {
		if !found {
			t.Errorf("subcomando 'chat %s' não encontrado", name)
		}
	}
}
