# Especificação Técnica de Implementação — Task 17: CLI Client Multi-Command Support

**Projeto:** Jay Core — Fase 3.5  
**Pacote Go:** `cli/cmd/jay`  
**Arquivo principal:** `cli/cmd/jay/main.go`  
**Dependências obrigatórias:** `sdk/ipc` (Task 09), IPC Router (Task 10), Daemon Unix Socket (Task 16)  
**Status:** Aguardando aprovação

---

## 1. Contexto

A Task 16 unificou toda a arquitetura física e lógica do Jay Core no daemon `jayd`, conectando o servidor Unix Domain Socket a todas as 19 rotas RPC.

A **Task 17** atualiza o cliente executável de linha de comando `jay` (`cli/cmd/jay/main.go`), permitindo interagir interativa ou administrativamente com o Daemon através de subcomandos estruturados construídos sobre o `sdk/ipc`.

---

## 2. Comandos e Sintaxe da CLI (`jay`)

| Comando | MessageType | Descrição | Exemplo de Uso |
|---|---|---|---|
| `jay register [client_id]` | `MsgRegisterClient (100)` | Registra um consumidor cliente no Core | `jay register cli_admin` |
| `jay unregister [client_id]` | `MsgUnregisterClient (101)` | Cancela o registro de um cliente | `jay unregister cli_admin` |
| `jay chat create [title]` | `MsgCreateChat (200)` | Cria um novo container de conversa | `jay chat create "Engenharia de Software"` |
| `jay chat list` | `MsgListChats (204)` | Lista os chats pertencentes ao cliente | `jay chat list` |
| `jay chat rename [id] [title]` | `MsgRenameChat (202)` | Altera o título de um chat | `jay chat rename chat-123 "Arquitetura Go"` |
| `jay chat delete [id]` | `MsgDeleteChat (201)` | Executa o Soft Delete em um chat | `jay chat delete chat-123` |
| `jay msg send [chat_id] "text"` | `MsgCreateMessage (300)` | Envia uma mensagem para o chat | `jay msg send chat-123 "Como funciona a Task 17?"` |
| `jay msg list [chat_id]` | `MsgGetMessages (303)` | Lista o histórico de mensagens do chat | `jay msg list chat-123` |
| `jay process [chat_id]` | `MsgProcessChat (350)` | Aciona a inferência da IA no chat | `jay process chat-123` |
| `jay tool register [id] [name] [desc]` | `MsgRegisterTool (400)` | Cadastra uma capacidade no catálogo | `jay tool register calc "Calculadora" "Faz contas"` |
| `jay tool list` | `MsgListTools (403)` | Lista as ferramentas ativas no catálogo | `jay tool list` |

---

## 3. Critérios de Aceite da Task

- [ ] Executável `jay` compilado com sucesso (`go build ./cli/cmd/jay`).
- [ ] Suporte a subcomandos estruturados (`register`, `chat`, `msg`, `tool`, `process`).
- [ ] Envio padronizado de envelopes `RequestEnvelope` (v1) e exibição legível de `ResponseEnvelope`.
- [ ] `go vet ./...` e `go test ./...` sem falhas em todo o repositório.
