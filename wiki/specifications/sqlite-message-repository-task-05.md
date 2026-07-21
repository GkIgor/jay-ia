# Especificação Técnica de Implementação — Task 05: Message Repository

**Projeto:** Jay Core — Fase 3.5  
**Pacote Go:** `core/internal/storage`  
**Arquivo principal:** `message_repository.go`  
**Dependências obrigatórias:** Task 01 (StorageEngine), Task 02 (MigrationEngine), Task 03 (RegistrationRepository), Task 04 (ChatRepository)  
**Status:** Concluído

---

## 1. Contexto

As Tasks 03 e 04 entregaram os repositórios de `Registration` e `Chat`, estabelecendo as convenções de API e o helper compartilhado `sqlite_errors.go`.

A Task 05 introduz o repositório de mensagens: **`MessageRepository`**, responsável pela persistência, recuperação e ordenação estrita do histórico de conversas no Jay Core.

---

## 2. Princípio de Repository Purity & Desacoplamento

Conforme a Regra Arquitetural do PRD (Seção 2.2):
> **O Message Repository é estritamente um serviço de persistência de dados. Ele NÃO aciona IA, NÃO chama agentes e NÃO interpreta comandos.**

---

## 3. Estratégia de Concorrência & Posse do `sequence_no`

### 3.1. Quem é dono do `sequence_no`?

> **Regra Arquitetural:** O `MessageRepository` é o **único componente autorizado** a atribuir `sequence_no` automático. Nenhuma camada superior deve calcular ou inferir esse valor para mensagens novas.

### 3.2. Concorrência via Transação de Escrita

Para garantir que duas requisições concorrentes no mesmo chat não recebam o mesmo `sequence_no`, a inserção com atribuição automática (`sequence_no == 0`) é executada em uma **transação atômica de escrita no SQLite**:

```sql
BEGIN;
SELECT COALESCE(MAX(sequence_no), 0) + 1 FROM messages WHERE chat_id = ?;
INSERT INTO messages (id, chat_id, author_type, author_id, role, content, content_type, status, sequence_no, metadata_json)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?);
COMMIT;
```

### 3.3. Comportamento com `sequence_no` Explícito (`sequence_no > 0`)

Caso a camada chamadora forneça `sequence_no > 0` (ex: importação de histórico ou sincronização externa):
- O repositório respeita o valor informado.
- Antes da inserção, verifica se já existe uma mensagem no mesmo chat com aquele `sequence_no`. Se já existir, a transação sofre Rollback e o repositório retorna `ErrAlreadyExists`.

---

## 4. Seção de Valores Padrão

| Campo | Valor recebido (Zero Value Go) | Valor padrão atribuído pelo Repository |
|---|---|---|
| `Status` | `0` | `MessageSent (1)` |
| `ContentType` | `0` | `ContentTypeTextPlain (1)` |
| `SequenceNo` | `0` | Calculado via `MAX(sequence_no) + 1` em transação |

---

## 5. Máquina de Estados e Ciclo de Vida da Mensagem

```
           +-------------+
           | MessageSent | (Criada)
           +-------------+
                  │
          Update()│ (Altera conteúdo)
                  ▼
          +---------------+
          | MessageEdited | (Permanece Edited em updates futuros)
          +---------------+
             │         │
     Delete()│         │Delete()
             ▼         ▼
          +----------------+
          | MessageDeleted | (Estado interno via Soft Delete)
          +----------------+
```

---

## 6. Tipos Go

### 6.1. Enums

```go
type AuthorType int
const (
    AuthorRegistration AuthorType = 1
    AuthorAgent        AuthorType = 2
    AuthorTool         AuthorType = 3
    AuthorSystem       AuthorType = 4
)

type MessageRole int
const (
    RoleUser      MessageRole = 1
    RoleAssistant MessageRole = 2
    RoleSystem    MessageRole = 3
    RoleTool      MessageRole = 4
)

type MessageContentType int
const (
    ContentTypeTextPlain  MessageContentType = 1
    ContentTypeMarkdown   MessageContentType = 2
    ContentTypeJSON       MessageContentType = 3
    ContentTypeToolCall   MessageContentType = 4
    ContentTypeToolResult MessageContentType = 5
)

type MessageStatus int
const (
    MessageSent    MessageStatus = 1
    MessageEdited  MessageStatus = 2
    MessageDeleted MessageStatus = 3
)
```

### 6.2. Struct `Message`

```go
type Message struct {
    ID           string
    ChatID       string
    AuthorType   AuthorType
    AuthorID     string
    Role         MessageRole
    Content      string
    ContentType  MessageContentType
    Status       MessageStatus
    SequenceNo   int
    MetadataJSON string
    CreatedAt    string
    UpdatedAt    string
}
```

---

## 7. Critérios de Aceite da Task

- [x] `ErrInvalidChat` adicionado a `errors.go`.
- [x] `message_repository.go` compilando sem erros (`go build ./...`).
- [x] `go vet ./...` e `go test ./...` sem falhas em todo o repositório.
- [x] `sequence_no` automático funcionando via transação no `Create`.
- [x] Detecção de duplicação em `sequence_no` explícito retornando `ErrAlreadyExists`.
- [x] Soft Delete idempotente funcionando.
- [x] `ListByChat` respeitando `sinceSequenceNo`, limit cap em 500 e ordenação `sequence_no ASC`.
