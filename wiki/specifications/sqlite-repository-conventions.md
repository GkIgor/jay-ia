# Convenções e Padrões Arquiteturais da Camada de Repositórios SQLite (`core/internal/storage`)

**Documento:** Especificação Arquitetural de Referência  
**Pacote Go:** `core/internal/storage`  
**Escopo:** Padrões imperativos para todos os repositórios (`RegistrationRepository`, `ChatRepository`, `MessageRepository`, `SharedRuleRepository`, `ToolRepository`)

---

## 1. Princípio de Repository Purity

> **O Repository apenas armazena, atualiza e recupera dados no banco SQLite. Ele NUNCA interpreta regras de negócio, permissões de acesso, comandos de protocolo ou regras de formato de aplicação.**

### Regras:
- Repositórios não validam segredos, permissões de usuário ou regras de autorização (`PermissionEvaluator`).
- Repositórios não disparam IA, executores de ferramentas ou handlers IPC.
- Repositórios não validam formatos de aplicação complexos (ex: JSON Schema, SemVer, Regex). Apenas realizam verificações básicas de presença de argumentos obrigatórios (`strings.TrimSpace(id) != ""`).

---

## 2. Padrão de Assinatura e Retorno por Ponteiro

Todos os repositórios do pacote `storage` seguem estritamente a convenção de assinaturas:

| Operação | Convenção de Retorno | Exemplo |
|---|---|---|
| Consulta por ID | `(*Entity, error)` | `FindByID(id string) (*Chat, error)` |
| Listagem | `([]*Entity, error)` | `ListByOwner(ownerID string) ([]*Chat, error)` |
| Escrita | `error` ou `(int, error)` | `Create(chat Chat) error`, `ReplaceRules(...) (int, error)` |

### Regras de Retorno:
1. **Retorno de Ponteiros (`*Entity`)**: Consultas e listagens retornam ponteiros para structs recém-alocadas. Isso evita cópia de valores à medida que as entidades evoluem e garante consistência em toda a API.
2. **Sem Ausência Silenciosa (`(nil, nil)`)**: Métodos de busca individual por ID nunca retornam `(nil, nil)`. Se o recurso não existir, retornam `(nil, ErrNotFound)`.
3. **Slice Vazia em Listagens**: Quando uma listagem não encontra registros no banco, ela retorna uma slice **inicializada vazia (`[]*Entity{}`)**, e **NUNCA `nil`**.
4. **Ownership das Instâncias**: As structs retornadas pertencem exclusivamente ao chamador e não são reutilizadas internamente pelo repositório.

---

## 3. Mapeamento de Erros em Duas Camadas (Técnico vs. Semântico)

A camada de persistência divide a tradução de erros em duas etapas bem definidas:

```
[ Driver modernc.org/sqlite ]
              │ (mensagem de erro bruta)
              ▼
[ sqlite_errors.go: translateSQLiteError ]
              │ (Erros Técnicos Neutros de Infraestrutura)
              ├── ErrUniqueViolation
              └── ErrForeignKeyViolation
              │
              ▼
[ Repositório Específico: mapXError ]
              │ (Tradução para Erro Semântico do Recurso)
              ├── Em RegistrationRepository -> ErrAlreadyExists / ErrDeleteRestricted
              ├── Em ChatRepository        -> ErrInvalidOwner
              ├── Em MessageRepository     -> ErrInvalidChat
              ├── Em SharedRuleRepository  -> ErrInvalidRegistration
              └── Em ToolRepository        -> ErrOwnershipConflict
```

### Sentinelas Técnicos de Infraestrutura (`sqlite_errors.go`):
- `ErrUniqueViolation`: Mapeado a partir de `"UNIQUE constraint failed"`.
- `ErrForeignKeyViolation`: Mapeado a partir de `"FOREIGN KEY constraint failed"`.

### Sentinelas Semânticos do Domínio (`errors.go`):
- `ErrNotFound`: Registro não encontrado ou soft-deleted.
- `ErrAlreadyExists`: Chave primária ou identificador único duplicado.
- `ErrInvalidArgument`: Argumento obrigatório vazio ou inválido.
- `ErrDeleteRestricted`: Remoção bloqueada por restrição `ON DELETE RESTRICT`.
- `ErrInvalidOwner`: O `owner_registration_id` não existe na tabela `registrations`.
- `ErrInvalidChat`: O `chat_id` não existe na tabela `chats`.
- `ErrInvalidRegistration`: O `registration_id` não existe na tabela `registrations`.
- `ErrOwnershipConflict`: Tentativa de registrar recurso com ID já pertencente a outra identidade.

---

## 4. Matriz de Estratégias de Exclusão (Soft Delete vs. Hard Delete)

| Entidade | Estratégia | Comportamento no SQL | Idempotência em `Delete(id)` |
|---|---|---|---|
| **`Registration`** | **Hard Delete** | `DELETE FROM registrations WHERE id = ?` | Retorna `nil` se já não existir |
| **`Chat`** | **Soft Delete** | `UPDATE chats SET status = 3, updated_at = ? WHERE id = ?` | Retorna `nil` se já estiver `status = 3` ou se não existir |
| **`Message`** | **Soft Delete** | `UPDATE messages SET status = 3, updated_at = ? WHERE id = ?` | Retorna `nil` se já estiver `status = 3` ou se não existir |
| **`SharedRule`** | **Hard Delete** | `DELETE FROM shared_rules WHERE registration_id = ?` | Retorna `nil` se já não existirem regras |
| **`Tool`** | **Hard Delete** | `DELETE FROM tools WHERE id = ?` | Retorna `nil` se já não existir |

### Regra Universal de Idempotência do `Delete(id)`:
> **Todos os métodos `Delete(id)` do pacote `storage` são 100% idempotentes.**  
> Se o recurso foi excluído anteriormente ou já não existe no banco, a operação retorna `nil`. Se o `id` fornecido for vazio (`""`), retorna `ErrInvalidArgument`.

---

## 5. Imutabilidade de Campos

Camadas superiores não podem alterar campos estruturais de identidade após a criação. Repositórios ignoram alterações nestes campos durante chamadas de `Update`:

| Entidade | Campos Imutáveis |
|---|---|
| **`Registration`** | `id`, `created_at` |
| **`Chat`** | `id`, `owner_registration_id`, `created_at` |
| **`Message`** | `id`, `chat_id`, `author_type`, `author_id`, `role`, `sequence_no`, `created_at` |
| **`SharedRule`** | `id`, `created_at` (`registration_id` do método substitui qualquer valor interno) |
| **`Tool`** | `id`, `registration_id`, `created_at` |

---

## 6. Aplicação de Valores Padrão (Fallbacks)

Quando chamadas de criação/registro recebem o valor zero Go de seus tipos, o repositório atribui automaticamente os seguintes defaults:

- **`Chat`**: `status == 0` → `ChatActive (1)`.
- **`Message`**: `status == 0` → `MessageSent (1)`; `content_type == 0` → `ContentTypeTextPlain (1)`.
- **`Tool`**: `status == 0` → `ToolAvailable (1)`; `version == ""` → `"1.0.0"`; `schema_json == ""` → `"{}"`.

---

## 7. Determinismo de Ordenação

Consultas de listagem aplicam critérios de ordenação determinísticos:

- `RegistrationRepository.List()` → `ORDER BY created_at ASC`
- `ChatRepository.ListByOwner()` → `ORDER BY updated_at DESC, created_at DESC`
- `MessageRepository.ListByChat()` → `ORDER BY sequence_no ASC`
- `SharedRuleRepository.ListByRegistration()` → `ORDER BY id ASC`
- `ToolRepository.ListAvailable()` → `ORDER BY name ASC`
