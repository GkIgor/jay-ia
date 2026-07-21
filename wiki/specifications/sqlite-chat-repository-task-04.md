# Especificação Técnica de Implementação — Task 04: Chat Repository

**Projeto:** Jay Core — Fase 3.5  
**Pacote Go:** `core/internal/storage`  
**Arquivo principal:** `chat_repository.go`  
**Dependências obrigatórias:** Task 01 (StorageEngine), Task 02 (MigrationEngine), Task 03 (RegistrationRepository)  
**Status:** Concluído

---

## 1. Contexto

A Task 03 estabeleceu o `RegistrationRepository` e definiu as convenções de API e princípios arquiteturais para a camada de acesso a dados (`Repository Purity`, ponteiros em consultas, centralização de erros do SQLite, etc.).

A Task 04 introduz a segunda entidade do modelo de domínio: o **`ChatRepository`**, responsável pelas operações de persistência sobre os containers de conversa (`Chat`).

---

## 2. Separação Arquitetural: Erros Técnicos vs. Erros Semânticos

```
[ Driver modernc.org/sqlite ]
              │ (mensagem de erro bruta)
              ▼
[ sqlite_errors.go: translateSQLiteError ]
              │ (retorna Erros Técnicos Genéricos da Infraestrutura)
              ├── ErrUniqueViolation
              └── ErrForeignKeyViolation
              │
              ▼
[ Repository Específico (RegistrationRepo / ChatRepo) ]
              │ (mapeia Erro Técnico em Erro Semântico do Domínio/Recurso)
              ├── Em RegistrationRepo.Delete -> ErrDeleteRestricted
              └── Em ChatRepo.Create        -> ErrInvalidOwner
```

### Regras:
1. `sqlite_errors.go` **não conhece** regras de domínio ou entidades (`Registration`, `Chat`, `Owner`). Ele produz apenas sentinelas neutros de infraestrutura banco (`ErrUniqueViolation`, `ErrForeignKeyViolation`).
2. O **Repository específico** avalia o contexto do método chamado e traduz o erro neutro da infraestrutura no erro semântico apropriado para a camada de serviço.

---

## 3. Máquina de Estados e Ciclo de Vida do Chat

O campo `status` da entidade `Chat` segue a máquina de estados abaixo:

```
           +--------------+
           | ChatActive   |
           +--------------+
              │        ▲
      Update()│        │Update()
              ▼        │
           +--------------+
           | ChatArchived |
           +--------------+
              │        │
      Delete()│        │Delete()
              ▼        ▼
           +--------------+
           | ChatDeleted  | (Estado final/interno)
           +--------------+
```

### Regras de Transição de Estado:
- **`ChatActive` ↔ `ChatArchived`**: Permitido bidirecionalmente exclusivamente via `Update(chat)`.
- **`ChatActive` / `ChatArchived` → `ChatDeleted`**: Permitido exclusivamente via `Delete(id)`.
- **Tentativa de atribuir `ChatDeleted` via `Update(chat)`**: **Proibida**. Retorna `ErrInvalidArgument`.
- **Tentativa de ressuscitar chat `ChatDeleted` → `ChatActive`/`ChatArchived`**: **Proibida**. Retorna `ErrNotFound`.

---

## 4. Escopo

### O que esta Task faz

- Extrai e isola a tradução técnica de erros do SQLite em `sqlite_errors.go`, com os sentinelas neutros `ErrUniqueViolation` e `ErrForeignKeyViolation`.
- Define o enum `ChatStatus` (`ChatActive`, `ChatArchived`, `ChatDeleted`).
- Define o enum `ChatFilter` (`ChatFilterActiveOnly`, `ChatFilterIncludeArchived`) para consultas legíveis.
- Define a struct Go `Chat`.
- Implementa `ChatRepository` com a API padronizada:
  - `Create(chat Chat) error`
  - `FindByID(id string) (*Chat, error)`
  - `ListByOwner(ownerRegistrationID string, filter ChatFilter) ([]*Chat, error)`
  - `Update(chat Chat) error`
  - `Delete(id string) error` (Soft Delete Idempotente)

---

## 5. Decisões Arquiteturais Detalhadas

### 5.1. Imutabilidade de Campos em `Update`

O método `Update(chat Chat)` atualiza **apenas**:
- `title`
- `status` (`ChatActive` ou `ChatArchived`)
- `metadata_json`
- `updated_at` (preenchido com `time.Now().UTC().Format(time.RFC3339)`)

Os campos `id`, `owner_registration_id` e `created_at` são **estritamente imutáveis**. Cualquier valor passado nesses campos no parâmetro `chat` é ignorado na cláusula `SET` do SQL.

### 5.2. Idempotência do `Delete(id)`

O método `Delete(id string)` realiza Soft Delete (`status = ChatDeleted`).

Comportamento:
- Se o chat existe e está `ChatActive` ou `ChatArchived`: atualiza `status = ChatDeleted`, renova `updated_at` e retorna `nil`.
- Se o chat existe mas **já está** `ChatDeleted`: a operação é **idempotente** e retorna `nil`.
- Se o `id` **não existe** no banco de dados: retorna `ErrNotFound`.

### 5.3. Hiding de Registros Deletados em `FindByID`

`FindByID(id)` consulta `WHERE id = ? AND status != 3`. Chats com Soft Delete são tratados como **inexistentes** para a aplicação e retornam `ErrNotFound`.

---

## 6. Tipos Go

### 6.1. Enums

```go
type ChatStatus int

const (
    ChatActive   ChatStatus = 1
    ChatArchived ChatStatus = 2
    ChatDeleted  ChatStatus = 3
)

type ChatFilter int

const (
    ChatFilterActiveOnly      ChatFilter = 0
    ChatFilterIncludeArchived ChatFilter = 1
)
```

### 6.2. Struct `Chat`

```go
type Chat struct {
    ID                  string
    OwnerRegistrationID string
    Title               string
    Status              ChatStatus
    MetadataJSON        string
    CreatedAt           string
    UpdatedAt           string
}
```

---

## 7. Critérios de Aceite da Task

- [x] Arquivo `sqlite_errors.go` criado contendo `translateSQLiteError`.
- [x] `registration_repository.go` refatorado para utilizar `sqlite_errors.go`.
- [x] `chat_repository.go` compilando sem erros (`go build ./...`).
- [x] `go vet ./...` e `go test ./...` sem falhas.
- [x] Soft Delete idempotente funcionando.
- [x] Proibição de transição para `ChatDeleted` via `Update` ou `Create`.
- [x] Mapeamento semântico de `ErrForeignKeyViolation` para `ErrInvalidOwner`.
- [x] `ListByOwner` utilizando `ChatFilter` e ordenando por `updated_at DESC, created_at DESC`.
