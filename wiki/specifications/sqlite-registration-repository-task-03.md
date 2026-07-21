# Especificação Técnica de Implementação — Task 03: Registration Repository

**Projeto:** Jay Core — Fase 3.5  
**Pacote Go:** `core/internal/storage`  
**Arquivo principal:** `registration_repository.go`  
**Dependências obrigatórias:** Task 01 (StorageEngine), Task 02 (MigrationEngine)  
**Status:** Concluído

---

## 1. Contexto

As Tasks 01 e 02 entregaram a infraestrutura SQLite: o `StorageEngine` gerencia o ciclo de vida da conexão e o `MigrationEngine` garante que o esquema v1 esteja aplicado antes de qualquer operação de domínio.

A Task 03 introduz a primeira unidade de acesso a dados de domínio: o **`RegistrationRepository`**, responsável pelas operações CRUD sobre a entidade `Registration`.

Esta Task também estabelece as **convenções de API e os padrões arquiteturais** que todos os repositories futuros (`ChatRepository`, `MessageRepository`, `ToolRepository`) devem seguir.

### O que é uma Registration?

Uma `Registration` representa uma **identidade lógica** conhecida pelo Core — por exemplo, `"jay_client_cpp"`, `"jay_client_cli"`, `"slack_client"`. Ela **não** representa uma conexão física. Conexões de socket são efêmeras e podem abrir e fechar livremente sem afetar a existência do registro.

O Core não realiza autenticação. Quem fornece um `id` é responsável por ele. O Core apenas persiste, localiza e remove identidades.

---

## 2. Princípio Arquitetural: Repository Purity

> **O Repository apenas persiste e recupera dados. Ele nunca interpreta regras de domínio, permissões ou protocolo.**

Este princípio deve ser respeitado em **todos** os repositories do pacote `storage`, agora e no futuro.

Exemplos de violações **proibidas** dentro de um repository:

```go
// ERRADO: regra de negócio dentro do repository
if reg.Status == RegistrationSuspended {
    return ErrForbidden
}

// ERRADO: lógica de permissão dentro do repository
if requestingClientID != reg.ID {
    return ErrUnauthorized
}

// ERRADO: decisão de protocolo dentro do repository
if len(reg.MetadataJSON) == 0 {
    reg.MetadataJSON = defaultMetadata()
}
```

Se uma dessas verificações for necessária, ela pertence à camada de **Service**, não ao repository. O repository assume que os dados fornecidos já foram validados pela camada superior.

---

## 3. Convenção de API para Repositories

A Task 03 estabelece o padrão de assinatura que todos os repositories do pacote `storage` devem seguir:

| Método        | Assinatura                                              | Obrigatório?                      |
|---------------|---------------------------------------------------------|-----------------------------------|
| Construtor    | `NewXRepository(db *sql.DB) (*XRepository, error)`      | Sempre                            |
| Criação       | `Create(entity X) error`                                | Sempre                            |
| Upsert        | `Upsert(entity X) error`                                | Quando a semântica de protocolo exigir |
| Busca por ID  | `FindByID(id string) (*X, error)`                       | Sempre                            |
| Listagem      | `List() ([]*X, error)`                                  | Sempre                            |
| Remoção       | `Delete(id string) error`                               | Sempre                            |
| Atualização   | `Update(entity X) error`                                | Quando não houver Upsert          |

**Regras:**
- Retornar sempre ponteiros para structs (ver seção 5 para justificativa).
- Retornar slices vazias (não nil) quando não houver resultados em `List`.
- Nunca retornar structs parcialmente preenchidas em nenhum método.
- Aceitar e repassar `context.Context` será avaliado em Task futura, após necessidade comprovada de cancelamento.

---

## 4. Escopo

### O que esta Task faz

- Define o tipo Go `Registration` que mapeia exatamente as colunas da tabela `registrations`.
- Define o enum `RegistrationStatus`.
- Implementa `RegistrationRepository` com os métodos: `Create`, `Upsert`, `FindByID`, `List`, `Delete`.
- Implementa o helper privado do pacote `translateSQLiteError` para centralização do mapeamento de erros do driver.
- Adiciona 4 sentinelas de erro em `errors.go`.

### O que esta Task NÃO faz

- **Não implementa** `SharedRule` nem qualquer lógica de permissões.
- **Não implementa** Service ou Handler sobre o repository.
- **Não implementa** paginação em `List` — cardinalidade de registrations é baixa e controlada.
- **Não implementa** atualização parcial (PATCH) — a operação de atualização de registrations é sempre `Upsert`.
- **Não implementa** Soft Delete — `Registration` usa Hard Delete conforme PRD seção 5.3.
- **Não implementa** filtros em `List` — YAGNI.

---

## 5. Decisões Arquiteturais

### 5.1. Retornar `*Registration` em vez de `Registration`

`FindByID` e `List` retornam ponteiros (`*Registration`, `[]*Registration`), não valores.

**Justificativa consciente:**

Hoje a struct `Registration` é pequena (5 campos simples). Retornar por valor funcionaria sem impacto de performance. No entanto, a convenção de ponteiros foi adotada pelos seguintes motivos:

1. **Evolução sem quebra de API:** A struct pode crescer. Se `Registration` ganhar campos como `CachedTools []Tool` ou `ActiveSessions []VoiceSession`, copiar por valor se tornaria custoso sem nenhuma mudança na assinatura dos callers.
2. **Consistência entre repositories:** Todos os repositories futuros seguem a mesma convenção. Misturar valor e ponteiro dependendo do tamanho atual da struct criaria inconsistência e confusão.
3. **Nil como ausência explícita:** `*Registration` nil é uma representação inequívoca de "não encontrado" — mas **este repositório nunca retorna nil sem erro**. O nil só existe como zero value, nunca como sinal de negócio. Quando não encontrado, retorna `ErrNotFound`.

> **Invariante:** `FindByID` nunca retorna `(nil, nil)`. Ou retorna `(*Registration, nil)` com struct completamente preenchida, ou retorna `(nil, erro)`.

### 5.2. Por que `Upsert` e não apenas `Update`?

O PRD especifica `RegisterClient (type = 100)` com semântica `INSERT ... ON CONFLICT DO UPDATE`. Um cliente que reinicia deve poder se re-registrar sem receber `ErrAlreadyExists`. Logo:

- `Create` → para criação explícita onde duplicidade é um erro.
- `Upsert` → para re-registro idempotente do protocolo.

Ambos coexistem para dar à camada de serviço a escolha do contrato correto.

### 5.3. Quem preenche `updated_at` por operação

Este ponto é explicitado aqui para evitar ambiguidade:

| Operação | `created_at`            | `updated_at`                              |
|----------|-------------------------|-------------------------------------------|
| `Create` | Preenchido pelo banco via `DEFAULT (strftime(...))` | Preenchido pelo banco via `DEFAULT (strftime(...))` |
| `Upsert` (insert) | Preenchido pelo banco via `DEFAULT` | Preenchido pelo banco via `DEFAULT` |
| `Upsert` (update) | **Não alterado** | Fornecido pelo código Go via `time.Now().UTC().Format(time.RFC3339)` |

**Motivo de `Upsert` (update) usar Go em vez de `strftime` do banco:** A cláusula `ON CONFLICT DO UPDATE` do SQLite não re-executa os `DEFAULT` para campos não listados no `SET`. Para garantir que `updated_at` seja atualizado, o valor deve ser fornecido explicitamente na cláusula `SET`.

### 5.4. Hard Delete e Foreign Keys

O PRD (seção 5.3) define que `Registration` usa Hard Delete. As consequências são:

- `shared_rules` associadas: removidas automaticamente via `ON DELETE CASCADE`.
- `chats` associados: o DDL usa `ON DELETE RESTRICT`, portanto o SQLite **bloqueia** a remoção se existirem chats ativos. O repository captura este erro e retorna `ErrDeleteRestricted`.

### 5.5. `metadata_json` como string opaca

O repository não conhece JSON. Ele move bytes. A serialização e desserialização do campo `metadata_json` é responsabilidade exclusiva da camada de serviço que chama o repository. Isso mantém o pacote `storage` sem dependência de formato de dados.

### 5.6. Ordenação de `List` por `created_at ASC`

A ordenação por `created_at ASC` fornece uma **visão cronológica estável** da criação das identidades lógicas. Esta escolha é preferida em detrimento de `id` (que é string opaca sem ordem semântica garantida) ou `updated_at` (que muda a cada re-registro e não reflete a ordem de chegada no sistema).

### 5.7. Helper `translateSQLiteError` centralizado

O driver `modernc.org/sqlite` não expõe tipos de erro estruturados. A única forma de distinguir `UNIQUE constraint failed` de `FOREIGN KEY constraint failed` é via `strings.Contains` na mensagem de erro.

Para evitar que esse padrão seja duplicado em cada repository futuro, esta Task cria um helper **privado do pacote** (não exportado):

```go
func translateSQLiteError(err error) error
```

Toda tradução de erro SQLite em **qualquer repository** do pacote passa por esta função. Se o driver mudar as mensagens de erro em versão futura, altera-se um único lugar.

---

## 6. Estrutura de Arquivos

```
core/internal/storage/
├── engine.go                         [EXISTENTE — sem alteração]
├── engine_test.go                    [EXISTENTE — sem alteração]
├── errors.go                         [MODIFICAR — adicionar 4 sentinelas]
├── migrations.go                     [EXISTENTE — sem alteração]
├── migrations_test.go                [EXISTENTE — sem alteração]
├── migrations_v1.go                  [EXISTENTE — sem alteração]
├── registration_repository.go        [NOVO — repository + translateSQLiteError]
└── registration_repository_test.go   [NOVO — testes]
```

---

## 7. Tipos Go

### 7.1. Enum `RegistrationStatus`

```go
type RegistrationStatus int

const (
    RegistrationActive    RegistrationStatus = 1
    RegistrationInactive  RegistrationStatus = 2
    RegistrationSuspended RegistrationStatus = 3
)
```

Os valores inteiros correspondem diretamente às colunas do banco. Nenhuma conversão de string ocorre no driver.

### 7.2. Struct `Registration`

```go
// Registration representa uma identidade lógica conhecida pelo Jay Core.
// Persiste na tabela `registrations`.
//
// Invariantes:
//   - ID nunca é vazio em uma instância retornada pelo repository.
//   - MetadataJSON é uma string opaca; o repository não faz parse deste campo.
//   - CreatedAt e UpdatedAt estão no formato ISO-8601 UTC (RFC3339).
type Registration struct {
    ID           string
    MetadataJSON string
    Status       RegistrationStatus
    CreatedAt    string
    UpdatedAt    string
}
```

### 7.3. `RegistrationRepository`

```go
type RegistrationRepository struct {
    db *sql.DB
}

func NewRegistrationRepository(db *sql.DB) (*RegistrationRepository, error)
```

Retorna `ErrNilDatabase` se `db` for nil.

---

## 8. Contratos dos Métodos

### 8.1. `Create(reg Registration) error`

Insere um novo `Registration`. `created_at` e `updated_at` são preenchidos pelo banco via `DEFAULT`.

```sql
INSERT INTO registrations (id, metadata_json, status)
VALUES (?, ?, ?);
```

| Condição                    | Retorno                |
|-----------------------------|------------------------|
| `id` vazio                  | `ErrInvalidArgument`   |
| `id` já existente           | `ErrAlreadyExists`     |
| Erro inesperado do banco    | erro wrappado          |
| Sucesso                     | `nil`                  |

---

### 8.2. `Upsert(reg Registration) error`

Insere ou atualiza. Em caso de conflito de `id`, atualiza `metadata_json`, `status` e `updated_at`. O campo `created_at` **nunca é alterado pelo Upsert**.

`updated_at` em caso de update é fornecido pelo código Go como `time.Now().UTC().Format(time.RFC3339)`.

```sql
INSERT INTO registrations (id, metadata_json, status, updated_at)
VALUES (?, ?, ?, ?)
ON CONFLICT(id) DO UPDATE SET
    metadata_json = excluded.metadata_json,
    status        = excluded.status,
    updated_at    = excluded.updated_at;
```

| Condição                 | Retorno              |
|--------------------------|----------------------|
| `id` vazio               | `ErrInvalidArgument` |
| Erro inesperado do banco | erro wrappado        |
| Sucesso (insert ou update) | `nil`              |

---

### 8.3. `FindByID(id string) (*Registration, error)`

Busca pelo `id`. Retorna ponteiro para `Registration` completamente preenchida, ou erro.

```sql
SELECT id, metadata_json, status, created_at, updated_at
FROM registrations
WHERE id = ?;
```

| Condição                 | Retorno                       |
|--------------------------|-------------------------------|
| `id` vazio               | `nil, ErrInvalidArgument`     |
| Registro não encontrado  | `nil, ErrNotFound`            |
| Erro inesperado do banco | `nil, erro wrappado`          |
| Sucesso                  | `*Registration, nil`          |

> **Invariante:** nunca retorna `(nil, nil)` nem struct parcialmente preenchida.

---

### 8.4. `List() ([]*Registration, error)`

Retorna todos os registros ordenados por `created_at ASC`. A ordenação cronológica garante estabilidade da visão da lista independente de re-registros.

```sql
SELECT id, metadata_json, status, created_at, updated_at
FROM registrations
ORDER BY created_at ASC;
```

| Condição                  | Retorno                           |
|---------------------------|-----------------------------------|
| Tabela vazia              | `[]*Registration{}` (vazia, não nil), `nil` |
| Erro inesperado do banco  | `nil, erro wrappado`              |
| Sucesso                   | `[]*Registration, nil`            |

---

### 8.5. `Delete(id string) error`

Remove fisicamente (Hard Delete). `shared_rules` associadas são removidas em cascata pelo banco. `chats` ativos bloqueiam a remoção.

A verificação de `ErrNotFound` é feita via `Result.RowsAffected() == 0`.  
A detecção de FK RESTRICT é feita via `translateSQLiteError`.

```sql
DELETE FROM registrations WHERE id = ?;
```

| Condição                                     | Retorno                  |
|----------------------------------------------|--------------------------|
| `id` vazio                                   | `ErrInvalidArgument`     |
| `id` não encontrado                          | `ErrNotFound`            |
| `Registration` referenciada por `Chat` ativo | `ErrDeleteRestricted`    |
| Erro inesperado do banco                     | erro wrappado            |
| Sucesso                                      | `nil`                    |

---

## 9. Helper `translateSQLiteError`

```go
// translateSQLiteError mapeia erros do driver modernc.org/sqlite para sentinelas do pacote.
// Deve ser usado por todos os repositories do pacote storage.
// Se o erro não for reconhecido, retorna o erro original sem alteração.
func translateSQLiteError(err error) error {
    if err == nil {
        return nil
    }
    msg := err.Error()
    switch {
    case strings.Contains(msg, "UNIQUE constraint failed"):
        return ErrAlreadyExists
    case strings.Contains(msg, "FOREIGN KEY constraint failed"):
        return ErrDeleteRestricted
    default:
        return err
    }
}
```

---

## 10. Novos Erros em `errors.go`

```go
// ErrNotFound é retornado quando o registro solicitado não existe no banco.
ErrNotFound = errors.New("storage: registro não encontrado")

// ErrAlreadyExists é retornado quando se tenta inserir um registro com ID já existente.
ErrAlreadyExists = errors.New("storage: registro já existe")

// ErrInvalidArgument é retornado quando um argumento obrigatório (ex: id) é inválido ou vazio.
ErrInvalidArgument = errors.New("storage: argumento inválido")

// ErrDeleteRestricted é retornado quando a remoção é bloqueada por dependência via FK RESTRICT.
ErrDeleteRestricted = errors.New("storage: remoção restrita por dependência existente")
```

---

## 11. Garantias do Repository

| Após...              | Garantia                                                                   |
|----------------------|----------------------------------------------------------------------------|
| `Create(reg)`        | `FindByID(reg.ID)` retorna o registro com os campos fornecidos.           |
| `Upsert(reg)`        | `FindByID(reg.ID)` retorna o registro; `created_at` é o da criação original se o ID já existia. |
| `Delete(id)`         | `FindByID(id)` retorna `ErrNotFound`.                                     |
| `List()`             | Retorna slice não-nil; vazia se não houver registros.                     |
| `FindByID(id)`       | Nunca retorna `(nil, nil)`; nunca retorna struct parcialmente preenchida. |

---

## 12. Critérios de Aceite da Task

- [x] `go build ./...` passa sem erros.
- [x] `go vet ./...` não reporta advertências.
- [x] `go test -v ./core/internal/storage/...` passa com 100% de sucesso.
- [x] Os 4 sentinelas estão em `errors.go`.
- [x] `translateSQLiteError` existe como função privada do pacote.
- [x] `metadata_json` é armazenado e retornado como string opaca — sem parse Go no repository.
- [x] `created_at` nunca é alterado pelo `Upsert`.
- [x] Hard Delete propaga cascade em `shared_rules` corretamente.
- [x] Hard Delete falha com `ErrDeleteRestricted` quando `Chat` ativo existe.
