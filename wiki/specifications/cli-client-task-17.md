# Especificação Técnica de Implementação — Task 17: CLI Client Multi-Command Support (com Cobra CLI)

**Projeto:** Jay Core — Fase 3.5  
**Pacote Go:** `cli/cmd/jay`  
**Arquivo principal:** `cli/cmd/jay/main.go`  
**Nova dependência:** `github.com/spf13/cobra`  
**Dependências obrigatórias:** `sdk/ipc` (Task 09), IPC Router (Task 10), Daemon Unix Socket (Task 16)  
**Status:** Aguardando aprovação

---

## 1. Contexto & Justificativa da Dependência `github.com/spf13/cobra`

A Task 16 unificou toda a arquitetura física e lógica do Jay Core no daemon `jayd`, conectando o servidor Unix Domain Socket a todas as 19 rotas RPC.

A **Task 17** atualiza o cliente executável de linha de comando `jay` (`cli/cmd/jay/main.go`), permitindo interagir interativa ou administrativamente com o Daemon através de subcomandos estruturados.

### Por que a dependência `github.com/spf13/cobra` vale a pena:
- **Padrão de Mercado Go**: O Cobra é o padrão da indústria em Go (usado por Kubernetes, Hugo, GitHub CLI `gh`).
- **Subcomandos Aninhados Idiomáticos**: Permite definir facilmente subcomandos como `jay chat create`, `jay msg send`, `jay tool list`.
- **Validação de Argumentos & Menus de Help**: Gera menus `--help` automáticos, validação estrita de quantidade de parâmetros e flags com tratamento POSIX nativo, dispensando a escrita de parsers manuais de `os.Args`.

---

## 2. Árvore de Comandos e Sintaxe da CLI (`jay`)

```
jay
├── register [client_id]                     (MsgRegisterClient - 100)
├── unregister [client_id]                   (MsgUnregisterClient - 101)
├── chat
│   ├── create [title]                       (MsgCreateChat - 200)
│   ├── list                                 (MsgListChats - 204)
│   ├── rename [chat_id] [title]             (MsgRenameChat - 202)
│   └── delete [chat_id]                     (MsgDeleteChat - 201)
├── msg
│   ├── send [chat_id] [content]             (MsgCreateMessage - 300)
│   └── list [chat_id]                       (MsgGetMessages - 303)
├── process [chat_id]                        (MsgProcessChat - 350)
└── tool
    ├── register [id] [name] [desc]          (MsgRegisterTool - 400)
    └── list                                 (MsgListTools - 403)
```

---

## 3. Critérios de Aceite da Task

- [ ] Dependência `github.com/spf13/cobra` adicionada ao `go.mod`.
- [ ] Executável `jay` compilado com sucesso (`go build ./cli/cmd/jay`).
- [ ] Suporte completo à árvore de subcomandos Cobra (`register`, `chat`, `msg`, `tool`, `process`).
- [ ] Envio padronizado de envelopes `RequestEnvelope` (v1) e exibição amigável de `ResponseEnvelope`.
- [ ] `go vet ./...` e `go test ./...` sem falhas em todo o repositório.
