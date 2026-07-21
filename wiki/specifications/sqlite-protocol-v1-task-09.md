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

## 2. Princípios Arquiteturais do SDK IPC

> **O pacote `sdk/ipc` é um pacote folha (zero dependências internas de `core/internal`). Ele define exclusivamente os contratos de dados, enums de protocolo, estruturas de envelope e utilitários de serialização JSON.**

---

## 3. Enums e Constantes do Protocolo v1

```go
type ProtocolVersion uint8
const ProtocolVersionCurrent ProtocolVersion = 1

type MessageType uint16
type ErrorCode uint16

// Enums próprios do SDK para isolamento de dependências
type AuthorType uint8
type MessageRole uint8
type MessageContentType uint8
type MessageStatus uint8
type ChatStatus uint8
type ToolStatus uint8
```

---

## 4. Estrutura de Envelopes do Protocolo e Exemplos JSON

```go
type RequestEnvelope struct {
    ProtocolVersion ProtocolVersion `json:"protocol_version"`
    RequestID       string          `json:"request_id"` // UUID v4
    ClientID        string          `json:"client_id"`
    Type            MessageType     `json:"type"`
    Payload         json.RawMessage `json:"payload"`
}

type ResponseEnvelope struct {
    ProtocolVersion ProtocolVersion `json:"protocol_version"`
    RequestID       string          `json:"request_id"`
    Type            MessageType     `json:"type"`
    Status          ErrorCode       `json:"status"`
    Error           *ErrorInfo      `json:"error,omitempty"`
    Payload         json.RawMessage `json:"payload,omitempty"`
}

type ErrorInfo struct {
    Message string `json:"message"`
    Details string `json:"details,omitempty"`
}
```

---

## 5. Regras de Compatibilidade e Evolução do Protocolo

1. **Garantia de Payload Não-Nulo**: `Payload` é sempre serializado como `{}` caso a struct de entrada seja `nil`. Nunca produz `null`.
2. **Eliminação de Redundância em Erros**: O `ResponseEnvelope` possui `Status ErrorCode`. O objeto interno `ErrorInfo` não duplica o código e contém apenas `message` e `details`.
3. **Versões Incompatíveis**: Requisições com `protocol_version > ProtocolVersionCurrent` são rejeitadas com `ErrInvalidFormat (4000)`.

---

## 6. Critérios de Aceite da Task

- [ ] Pacote `sdk/ipc` compilando sem erros (`go build ./...`).
- [ ] `go vet ./...` e `go test ./...` sem falhas em todo o repositório.
- [ ] Enums de `MessageType` (`uint16`) e `ErrorCode` (`uint16`) com valores numéricos do PRD.
- [ ] Enums de domínio próprios do SDK (`AuthorType`, `MessageRole`, `MessageContentType`, etc.) definidos no pacote.
- [ ] Eliminação da redundância no `ErrorInfo` (somente `message` e `details`).
- [ ] Garantia de que `Payload` nunca é `null` (fallback para `{}`).
- [ ] Helpers `MarshalPayload`, `UnmarshalPayload`, `NewRequestEnvelope`, `NewResponseEnvelope` e `NewErrorResponseEnvelope` implementados.
- [ ] 100% dos testes unitários de serialização JSON aprovados.
