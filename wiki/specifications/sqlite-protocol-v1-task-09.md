# Especificação Técnica de Implementação — Task 09: IPC Protocol Frame Mappings & Serialization

**Projeto:** Jay Core / SDK — Fase 3.5  
**Pacote Go:** `sdk/ipc`  
**Arquivos principais:** `sdk/ipc/protocol.go`, `sdk/ipc/messages.go`  
**Dependências obrigatórias:** Nenhuma (pacote folha do SDK)  
**Status:** Aguardando aprovação

---

## 1. Contexto

A Task 08 entregou o `Permission Evaluator Engine` em `core/internal/permission`, concluindo a lógica pura de autorização.

A Task 09 introduz a especificação de serialização e estruturação de mensagens do **Protocolo IPC v1** (`sdk/ipc`). O pacote `sdk/ipc` define as estruturas de dados compartilhadas entre clientes (CLI, C++ Frontend, extensões) e o Jay Core Daemon via Unix Domain Sockets (JSON-over-Socket).

---

## 2. Princípio Arquitetural do SDK IPC

> **O pacote `sdk/ipc` é um pacote folha (zero dependências internas de `core/internal`). Ele define exclusivamente os contratos de dados, enums de protocolo, estruturas de envelope e utilitários de serialização JSON.**

---

## 3. Enums e Constantes do Protocolo v1

### 3.1. Versão do Protocolo
```go
const ProtocolVersionCurrent = 1
```

### 3.2. Tipo de Mensagem (`MessageType`)
```go
type MessageType int

const (
    MsgRegisterClient     MessageType = 100
    MsgUnregisterClient   MessageType = 101
    MsgUpdateRegistration MessageType = 102
    MsgGetRegistration    MessageType = 103
    MsgListRegistrations  MessageType = 104
    MsgUpdateSharedRules  MessageType = 105

    MsgCreateChat MessageType = 200
    MsgDeleteChat MessageType = 201
    MsgRenameChat MessageType = 202
    MsgGetChat    MessageType = 203
    MsgListChats  MessageType = 204

    MsgCreateMessage MessageType = 300
    MsgUpdateMessage MessageType = 301
    MsgDeleteMessage MessageType = 302
    MsgGetMessages   MessageType = 303

    MsgProcessChat MessageType = 350

    MsgRegisterTool   MessageType = 400
    MsgUnregisterTool MessageType = 401
    MsgGetTool        MessageType = 402
    MsgListTools      MessageType = 403

    MsgCreateVoiceSession MessageType = 500
    MsgGetVoiceSession    MessageType = 501
    MsgCloseVoiceSession  MessageType = 502
)
```

### 3.3. Códigos de Erro Padronizados (`ErrorCode`)
```go
type ErrorCode int

const (
    ErrSuccess          ErrorCode = 0
    ErrInvalidFormat    ErrorCode = 4000
    ErrUnauthorized     ErrorCode = 4001
    ErrForbidden        ErrorCode = 4003
    ErrNotFound         ErrorCode = 4004
    ErrConflict         ErrorCode = 4009
    ErrInternalDatabase ErrorCode = 5000
    ErrNotImplemented   ErrorCode = 5001
)
```

---

## 4. Estrutura de Envelopes do Protocolo

```go
type RequestEnvelope struct {
    ProtocolVersion int             `json:"protocol_version"`
    RequestID       string          `json:"request_id"`
    ClientID        string          `json:"client_id"`
    Type            MessageType     `json:"type"`
    Payload         json.RawMessage `json:"payload"`
}

type ResponseEnvelope struct {
    ProtocolVersion int             `json:"protocol_version"`
    RequestID       string          `json:"request_id"`
    Type            MessageType     `json:"type"`
    Status          ErrorCode       `json:"status"`
    Error           *ErrorInfo      `json:"error,omitempty"`
    Payload         json.RawMessage `json:"payload,omitempty"`
}

type ErrorInfo struct {
    Code    ErrorCode `json:"code"`
    Message string    `json:"message"`
    Details string    `json:"details,omitempty"`
}
```

---

## 5. Critérios de Aceite da Task

- [ ] Pacote `sdk/ipc` compilando sem erros (`go build ./...`).
- [ ] `go vet ./...` e `go test ./...` sem falhas em todo o repositório.
- [ ] Todos os enums de `MessageType` e `ErrorCode` definidos com os códigos numéricos exatos do PRD.
- [ ] Construtores helper `NewRequestEnvelope`, `NewResponseEnvelope` e `NewErrorResponseEnvelope` disponíveis.
- [ ] `json.RawMessage` utilizado nos envelopes para desacoplamento seguro de payload.
- [ ] 100% dos testes unitários de serialização JSON aprovados.
