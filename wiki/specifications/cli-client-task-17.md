# Especificação Técnica de Implementação — Task 17: CLI Client Multi-Command Support (com Cobra CLI, Modo Config & Chat REPL)

**Projeto:** Jay Core — Fase 3.5  
**Pacote Go:** `cli/cmd/jay`  
**Arquivo principal:** `cli/cmd/jay/main.go`  
**Nova dependência:** `github.com/spf13/cobra`  
**Dependências obrigatórias:** `sdk/ipc` (Task 09), IPC Router (Task 10), Daemon Unix Socket (Task 16)  
**Status:** Aguardando aprovação

---

## 1. Contexto & Separação de Responsabilidades da CLI

A Task 16 unificou toda a arquitetura física e lógica do Jay Core no daemon `jayd`, conectando o servidor Unix Domain Socket a todas as 19 rotas RPC.

A **Task 17** atualiza o cliente executável de linha de comando `jay` (`cli/cmd/jay/main.go`), estruturado com o `github.com/spf13/cobra` com separação explícita entre:

1. **Modo Interativo de Configuração (`jay config` / `jay init`)**:
   - Voltado para setup, primeiro uso, ajustes de ambiente (ex: chave de API LLM, provedor, caminho do banco SQLite e registro de regras de acesso).
2. **Feature Dedicada de Chat Interativo (`jay chat repl` / `jay chat interactive`)**:
   - Subcomando específico sob o grupo `chat` para abrir um loop REPL conversacional em tempo real com a IA.

---

## 2. Árvore Completa de Comandos da CLI (`jay`)

```
jay
├── config / init                            (Modo Interativo para Ajustes e Setup do Ambiente)
├── register [client_id]                     (MsgRegisterClient - 100)
├── unregister [client_id]                   (MsgUnregisterClient - 101)
├── chat
│   ├── repl / interactive [chat_id]         (Loop REPL Conversacional Dedicado com a IA)
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
- [ ] Modo Interativo de Configuração `jay config` / `jay init` funcional para onboarding.
- [ ] Feature dedicada de Chat Interativo `jay chat repl` funcional.
- [ ] Suporte completo à árvore de subcomandos Cobra (`register`, `chat`, `msg`, `tool`, `process`).
- [ ] Envio padronizado de envelopes `RequestEnvelope` (v1) e exibição amigável de `ResponseEnvelope`.
- [ ] `go vet ./...` e `go test ./...` sem falhas em todo o repositório.
