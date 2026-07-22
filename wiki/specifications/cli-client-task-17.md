# Especificação Técnica de Implementação — Task 17: CLI Client Multi-Command Support (com Cobra CLI & Comandos Modularizados)

**Projeto:** Jay Core — Fase 3.5  
**Pacote Go:** `cli/cmd/jay` e subpacote `cli/cmd/jay/commands`  
**Arquivo principal:** `cli/cmd/jay/main.go`  
**Nova dependência:** `github.com/spf13/cobra`  
**Dependências obrigatórias:** `sdk/ipc` (Task 09), IPC Router (Task 10), Daemon Unix Socket (Task 16)  
**Status:** Aguardando aprovação

---

## 1. Contexto & Arquitetura da CLI Modularizada

A Task 16 unificou toda a arquitetura física e lógica do Jay Core no daemon `jayd`, conectando o servidor Unix Domain Socket a todas as 19 rotas RPC.

A **Task 17** atualiza o cliente de linha de comando `jay`, estruturando os comandos Cobra em **múltiplos arquivos separados por domínio** no pacote `cli/cmd/jay/commands`, garantindo escalabilidade e facilidade de manutenção para novas funcionalidades.

---

## 2. Estrutura Modular de Arquivos (`cli/cmd/jay`)

```
cli/
└── cmd/
    └── jay/
        ├── main.go                       [Ponto de entrada: chama commands.Execute()]
        └── commands/
            ├── root.go                   [Comando raiz do Cobra]
            ├── client.go                 [ipc Client wrapper para dial e IPC]
            ├── config.go                 [jay config / jay init — Setup em ~/.jay/.env]
            ├── register.go               [jay register & jay unregister]
            ├── chat.go                   [jay chat (create, list, rename, delete)]
            ├── chat_repl.go              [jay chat repl — REPL dedicado com /clear, /new, etc.]
            ├── message.go                [jay msg (send, list)]
            ├── tool.go                   [jay tool (register, list)]
            ├── process.go                [jay process — Trigger administrativo da IA]
            └── commands_test.go          [Testes unitários de parsing Cobra e envelopes IPC]
```

---

## 3. Critérios de Aceite da Task

- [ ] Dependência `github.com/spf13/cobra` adicionada ao `go.mod`.
- [ ] Estrutura modular em `cli/cmd/jay/commands/` com arquivos separados por domínio.
- [ ] Executável `jay` compilado com sucesso (`go build ./cli/cmd/jay`).
- [ ] Modo Interativo de Configuração `jay config` funcional escrevendo em `~/.jay/.env`.
- [ ] Feature dedicada de Chat Interativo `jay chat repl` funcional com feedback visual e comandos `/help`, `/clear`, `/new`, `/quit`.
- [ ] Suporte completo à árvore de subcomandos Cobra (`register`, `chat`, `msg`, `tool`, `process`).
- [ ] `go vet ./...` e `go test ./...` sem falhas em todo o repositório.
