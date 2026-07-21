# Especificação Técnica de Implementação — Task 07: Tool Repository

**Projeto:** Jay Core — Fase 3.5  
**Pacote Go:** `core/internal/storage`  
**Arquivo principal:** `tool_repository.go`  
**Dependências obrigatórias:** Task 01 (StorageEngine), Task 02 (MigrationEngine), Task 03 (RegistrationRepository), Task 04 (ChatRepository), Task 05 (MessageRepository), Task 06 (SharedRuleRepository)  
**Status:** Aguardando aprovação

---

## 1. Contexto

As Tasks 01 a 06 entregaram a infraestrutura SQLite, o motor de migrações e os repositórios para `Registration`, `Chat`, `Message` e `SharedRule`.

A Task 07 conclui a camada de repositórios SQLite do Jay Core introduzindo o **`ToolRepository`**, responsável pela persistência e catalogação das ferramentas / capacidades funcionais versionadas (`Tool`) registradas por consumidores do ecossistema.

### O que é uma Tool?

Uma `Tool` representa uma capacidade executável oferecida por um consumidor registrado (ex: `"web_search"`, `"system_exec"`, `"code_interpreter"`):
- `id` (String PK): Slug ou UUID único da ferramenta (ex: `"web_search"`).
- `registration_id` (String FK -> `registrations.id ON DELETE CASCADE`): Identidade lógica do consumidor que provê e executa a ferramenta.
- `name` (String): Nome legível da ferramenta.
- `description` (String): Descrição de propósito para ser consumida por Agentes de IA/LLM durante a seleção de ferramentas.
- `version` (String, default `"1.0.0"`): Versão SemVer da ferramenta.
- `schema_json` (String JSON opaca): Schema JSON dos parâmetros aceitos pela ferramenta.
- `status` (Enum `ToolStatus`): `ToolAvailable (1)`, `ToolDisabled (2)`, `ToolDeprecated (3)`.

---

## 2. Princípio de Repository Purity & Desacoplamento

> **O Tool Repository apenas armazena, atualiza e lista as ferramentas cadastradas no banco. Ele NÃO executa ferramentas, NÃO valida o JSON Schema e NÃO realiza chamadas IPC para invocar capacidades.**

---

## 3. Semântica de Registro Idempotente (`Register`)

Conforme especificado no PRD para o comando IPC `MsgRegisterTool (type = 400)`:
- Quando um consumidor reconecta ou reinicia, ele re-registra suas ferramentas disponíveis.
- O método de gravação é **`Register(tool Tool)`**, que possui semântica de **Upsert por `id`**:
  - Se a ferramenta não existir no banco → é criada (`INSERT`).
  - Se a ferramenta já existir no banco → seus dados (`name`, `description`, `version`, `schema_json`, `status` e `updated_at`) são atualizados (`ON CONFLICT(id) DO UPDATE`).
  - O campo `registration_id` da ferramenta **não muda** em atualizações de re-registro.

---

## 4. Estratégia de Remoção e Desativação

Conforme PRD Seção 5.3:
1. **Desativação Lógica**: Suportada via `UpdateStatus(id, status)` para transicionar uma ferramenta para `ToolDisabled (2)` ou `ToolDeprecated (3)`.
2. **Hard Delete**: Suportado via `Delete(id)` (`DELETE FROM tools WHERE id = ?`) ao desregistrar explicitamente a ferramenta via `UnregisterTool`.
3. **Deleção em Cascata SQL**: Se a `Registration` proprietária for excluída, todas as suas `Tool`s associadas são removidas automaticamente pelo SQLite via `FOREIGN KEY ... ON DELETE CASCADE`.

---

## 5. Tipos Go

### 5.1. Enum `ToolStatus`

```go
type ToolStatus int

const (
    ToolAvailable  ToolStatus = 1
    ToolDisabled   ToolStatus = 2
    ToolDeprecated ToolStatus = 3
)
```

### 5.2. Struct `Tool`

```go
type Tool struct {
    ID             string
    RegistrationID string
    Name           string
    Description    string
    Version        string
    SchemaJSON     string
    Status         ToolStatus
    CreatedAt      string
    UpdatedAt      string
}
```

---

## 6. Critérios de Aceite da Task

- [ ] `tool_repository.go` compilando sem erros (`go build ./...`).
- [ ] `go vet ./...` e `go test ./...` sem falhas em todo o repositório.
- [ ] Upsert em `Register` funcionando com fallbacks de versão (`1.0.0`) e status (`Available`).
- [ ] Proteção contra troca de `registration_id` em ferramentas existentes.
- [ ] Deleção em cascata SQL (`ON DELETE CASCADE`) confirmada por teste unitário.
- [ ] `ListAvailable` filtrando por `status = 1` e ordenando por `name ASC`.
