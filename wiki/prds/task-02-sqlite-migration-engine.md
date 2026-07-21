# Especificação Técnica de Implementação: Task 02 — Migration Engine & DDL Base

**Documento:** Technical Implementation Specification  
**Alvo:** `jay-ia/core/internal/storage`  
**Escopo Exclusivo:** Task 02 (Migration Engine, PRAGMA user_version, DDLs da v1 e Testes de Migração)  
**Status:** Especificação Técnica Oficial Definitiva  

---

## 1. Objetivo

### 1.1. Propósito da Task
O propósito exclusivo da **Task 02** é construir o motor de migrações de esquema de banco de dados (`MigrationEngine`) e registrar o conjunto inicial de DDLs (versão `v1`) para a infraestrutura SQLite do **Jay Core**.

### 1.2. Motivação da Existência
Na Task 01, estabeleceu-se a infraestrutura de conexão SQLite (`StorageEngine`). No entanto, um banco de dados relacional recém-criado encontra-se zerado, sem tabelas, índices ou restrições de integridade. Para que as Tasks de Repositórios (Tasks 03 a 07) possam operar, o Core necessita de um componente automatizado, determinístico e idempotente responsável por:
1. Detectar o estado e a versão atual do esquema no banco SQLite.
2. Criar incrementalmente todas as tabelas e índices da versão `v1` sob transações atômicas (ACID).
3. Atualizar o marcador de versão do SQLite (`PRAGMA user_version`).

### 1.3. Motivação do Isolamento em relação às Outras Tasks
Esta Task foi deliberadamente isolada para respeitar o **Princípio da Responsabilidade Única (SRP)**:
- A Task 01 é responsável por **ABRIR** e **CONFIGURAR** a infraestrutura física de I/O da conexão.
- A Task 02 é responsável por **CRIAR** e **EVOLUIR** a estrutura das tabelas e índices.
- As Tasks 03 a 07 serão responsáveis por **OPERAR** os dados (inserir, listar, atualizar) através de repositórios.

Isolar o `MigrationEngine` garante que o mecanismo de migração de esquema possa ser testado de forma independente antes de qualquer operação de escrita de dados de domínio.

---

## 2. Contexto Arquitetural

### 2.1. Posicionamento no Backlog de Fases
A Task 02 é o segundo passo da reconstrução da camada de persistência do Jay Core:

```
[Task 01: StorageEngine (Infraestrutura SQLite)]
                         │
                         ▼
[Task 02: Migration Engine & DDL Base (Tabelas e Índices v1)]
                         │
                         ▼
[Tasks 03-07: Repositórios SQLite (Registration, Chat, Message, Rule, Tool)]
                         │
                         ▼
[Task 08: Permission Evaluator Engine]
                         │
                         ▼
[Tasks 09-15: IPC Protocol v1, Message Service & Chat Processing]
                         │
                         ▼
[Tasks 16-17: Daemon & CLI Integration]
```

### 2.2. Dependências
- **Depende de**: Task 01 (`StorageEngine`), da qual consome a referência de conexão `*sql.DB` via `engine.DB()`.
- **Prepara o terreno para**:
  - **Task 03 (Registration Repository)**: Depende da tabela `registrations`.
  - **Task 04 (Chat Repository)**: Depende da tabela `chats`.
  - **Task 05 (Message Repository)**: Depende da tabela `messages`.
  - **Task 06 (SharedRule Repository)**: Depende da tabela `shared_rules`.
  - **Task 07 (Tool Repository)**: Depende da tabela `tools`.

---

## 3. Responsabilidades

### 3.1. O que Pertence à Task 02
- Implementar a struct `MigrationEngine` que aceita a conexão `*sql.DB`.
- Consultar a versão atual do esquema SQLite executando `PRAGMA user_version`.
- Executar os scripts DDL da versão `v1` caso a versão atual seja `0`.
- Executar as DDLs dentro de uma **transação atômica** (`*sql.Tx`).
- Atualizar a versão do banco para `PRAGMA user_version = 1` ao finalizar a migration com sucesso.
- Declarar as DDLs completas das 6 tabelas (`registrations`, `shared_rules`, `chats`, `messages`, `tools`, `voice_sessions`) e seus índices associados.
- Garantir comportamento **idempotente** (se a versão atual já for `1`, não re-executa as DDLs).

### 3.2. O que NÃO Pertence à Task 02
- **NENHUMA** abertura de arquivo de banco de dados ou aplicação de pragmas de I/O (`WAL`, `busy_timeout`). Isso é responsabilidade exclusiva da Task 01 (`StorageEngine`).
- **NENHUMA** inserção de dados iniciais (*seed data*) ou registros padrão.
- **NENHUMA** implementação de métodos CRUD de repositórios.
- **NENHUMA** lógica de regras de negócio, serviços, protocolo IPC ou handlers.

---

## 4. Princípios Arquiteturais Fundamentais

### 4.1. Princípio de Transacionalidade Atômica (*All-or-Nothing*)
> **Todas as instruções DDL de uma versão de migração devem ser executadas dentro de uma única transação SQLite (`BEGIN TRANSACTION ... COMMIT`). Se qualquer instrução DDL ou alteração de pragma falhar, a transação deve sofrer Rollback imediato, garantindo que o banco de dados nunca permaneça em um estado parcialmente migrado.**

### 4.2. Princípio da Imutabilidade de Migrações Concluídas
> **Uma versão de migração executada (ex: `v1`) torna-se imutável. Alterações futuras no banco de dados devem obrigatoriamente ser feitas através de novas migrations incrementais (`v2`, `v3`, etc.).**

---

## 5. Orquestração do Boot e Papel do Migration Engine

O `MigrationEngine` é um componente **passivo** acionado durante a inicialização do sistema.

### 5.1. Fluxo de Execução no Boot do Daemon (`jayd`)
```text
               main() [Daemon jayd]
                 │
                 ├── 1. Instancia e Abre a Infraestrutura (Task 01)
                 │      engine := storage.NewStorageEngine(cfg)
                 │      if err := engine.Open(); err != nil { exit(1) }
                 │
                 ├── 2. Executa o Migration Engine (Task 02)
                 │      migrator := storage.NewMigrationEngine(engine.DB())
                 │      if err := migrator.Run(); err != nil { exit(1) }
                 │
                 ├── 3. Instancia os Repositórios (Tasks 03-07)
                 │      chatRepo := storage.NewChatRepository(engine.DB())
                 │
                 └── 4. Inicia Servidores IPC e Atendimento
```

---

## 6. Arquitetura das Camadas

```text
+-------------------------------------------------------------+
|                     Daemon Main (jayd)                      |
+-------------------------------------------------------------+
                              │
                              │ Chama migrator.Run()
                              ▼
+-------------------------------------------------------------+
|              MigrationEngine (core/internal/storage)        |
|  - Lê PRAGMA user_version                                   |
|  - Executa Transação DDL (migrations_v1.go)                 |
|  - Atualiza PRAGMA user_version = 1                         |
+-------------------------------------------------------------+
                              │
                              │ Usa a referência *sql.DB da Task 01
                              ▼
+-------------------------------------------------------------+
|               StorageEngine (Task 01)                       |
+-------------------------------------------------------------+
```

---

## 7. Organização dos Diretórios e Arquivos

A Task 02 adicionará exatamente **2 arquivos de código e 1 arquivo de teste** no pacote `core/internal/storage/`:

```text
core/internal/storage/
├── engine.go        # (Task 01) Infraestrutura SQLite
├── errors.go        # (Task 01 & 02) Erros sentinela
├── engine_test.go   # (Task 01) Testes da infraestrutura
├── migrations.go    # (Task 02) Struct MigrationEngine, Run e controle de versão
├── migrations_v1.go # (Task 02) Constantes/slices com as instruções DDL da v1
└── migrations_test.go # (Task 02) Testes unitários de migração e DDLs
```

### 7.1. Detalhamento dos Novos Arquivos

#### `migrations.go`
- **Responsabilidade**: Conter o construtor `NewMigrationEngine(db *sql.DB)` e os métodos privados para leitura de `PRAGMA user_version` e aplicação atômica de versões.
- **O que NUNCA deve existir nele**: Declarações extensas de DDLs em string ou métodos de repositório.

#### `migrations_v1.go`
- **Responsabilidade**: Isolar as declarações SQL DDL puras da versão `v1` (tabelas e índices) em constantes/slices limpos de strings Go.
- **O que NUNCA deve existir nele**: Lógica de fluxo Go ou consultas de repositório.

#### `migrations_test.go`
- **Responsabilidade**: Validar em banco de dados `:memory:` e arquivo temporário se as migrações sobem um banco zerado, criam as 6 tabelas, definem os índices e atualizam o `user_version` para `1` de forma idempotente.

---

## 8. Diagrama de Estados do Migration Engine

```text
 [Uninitialized] ── NewMigrationEngine(db) ──> [Created]
                                                 │
                                              Run()
                                                 │
                                                 ▼
                                      [Reading UserVersion]
                                                 │
                           ┌─────────────────────┴─────────────────────┐
             user_version == 0                                  user_version == 1
                           │                                           │
                           ▼                                           ▼
               [Applying Migration V1]                          [UpToDate]
             (BEGIN TX -> DDLs -> COMMIT)                              │
                           │                                           │
             ┌─────────────┴─────────────┐                             │
          Sucesso                      Falha                           │
             │                           │                             │
             ▼                           ▼                             │
    user_version = 1            ROLLBACK TRANSACTION                   │
             │                           │                             │
             ▼                           ▼                             ▼
        [UpToDate]                   [Failed]                     Retorna nil
       Retorna nil                 Retorna error
```

---

## 9. Invariantes Arquiteturais da Task 02

1. **Invariante de Versão de Banco**: Após a execução com sucesso de `migrator.Run()`, a consulta `PRAGMA user_version;` retornará obrigatoriamente um valor maior ou igual a `1`.
2. **Invariante de Presença de Esquema**: Após a conclusão da migration `v1`, todas as 6 tabelas (`registrations`, `shared_rules`, `chats`, `messages`, `tools`, `voice_sessions`) e seus 5 índices existirão no banco de dados.
3. **Invariante de Idempotência**: Executar `migrator.Run()` múltiplas vezes consecutivas em um banco já migrado é umaoperação inofensiva (*no-op*) que retorna `nil` sem executar comandos DDL novamente.
4. **Invariante de Isolamento I/O**: O `MigrationEngine` não abre nem fecha conexões de arquivo de banco. Ele depende estritamente do `*sql.DB` fornecido pela Task 01.

---

## 10. Discussão Arquitetural: DDLs Embedadas em Go vs Arquivos `.sql` Externos

### 10.1. A Questão
> *Por que declarar os scripts DDL em constantes/slices Go em `migrations_v1.go` em vez de ler arquivos `.sql` externos do disco com `os.ReadFile`?*

### 10.2. Racional da Escolha (DDLs Embedadas em Go)
1. **Binário Único Autocontido**: O Jay Core é compilado como um único arquivo binário Go executável. Ter os scripts DDL embutidos diretamente no binário evita falhas de inicialização em tempo de execução causadas por arquivos `.sql` ausentes ou com caminhos relativos incorretos.
2. **Desempenho e Confiabilidade**: Elimina I/O desnecessário de leitura de arquivos no sistema de arquivos durante a inicialização.
3. **Segurança contra Tampering**: Evita que usuários ou processos externos alterem acidentalmente arquivos `.sql` no disco e corrompam o esquema de migração do Core.

---

## 11. Especificação Detalhada do Esquema da Versão 1 (`v1`)

As instruções SQL a seguir serão declaradas no arquivo `migrations_v1.go`:

### 11.1. Tabela `registrations` (Identidades Lógicas)
```sql
CREATE TABLE IF NOT EXISTS registrations (
    id TEXT PRIMARY KEY NOT NULL,
    metadata_json TEXT NOT NULL DEFAULT '{}',
    status INTEGER NOT NULL DEFAULT 1,
    created_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ', 'now')),
    updated_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ', 'now'))
);
```

### 11.2. Tabela `shared_rules` (Regra de Compartilhamento)
```sql
CREATE TABLE IF NOT EXISTS shared_rules (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    registration_id TEXT NOT NULL,
    target_scope INTEGER NOT NULL DEFAULT 0,
    pattern TEXT NOT NULL,
    match_type INTEGER NOT NULL DEFAULT 1,
    allowed_actions INTEGER NOT NULL DEFAULT 1,
    created_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ', 'now')),
    FOREIGN KEY (registration_id) REFERENCES registrations(id) ON DELETE CASCADE
);
```

### 11.3. Tabela `chats` (Conversas)
```sql
CREATE TABLE IF NOT EXISTS chats (
    id TEXT PRIMARY KEY NOT NULL,
    owner_registration_id TEXT NOT NULL,
    title TEXT NOT NULL,
    status INTEGER NOT NULL DEFAULT 1,
    metadata_json TEXT NOT NULL DEFAULT '{}',
    created_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ', 'now')),
    updated_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ', 'now')),
    FOREIGN KEY (owner_registration_id) REFERENCES registrations(id) ON DELETE RESTRICT
);
```

### 11.4. Tabela `messages` (Mensagens com Autoria Composta)
```sql
CREATE TABLE IF NOT EXISTS messages (
    id TEXT PRIMARY KEY NOT NULL,
    chat_id TEXT NOT NULL,
    author_type INTEGER NOT NULL DEFAULT 1,
    author_id TEXT NOT NULL,
    role INTEGER NOT NULL,
    content TEXT NOT NULL,
    content_type INTEGER NOT NULL DEFAULT 1,
    status INTEGER NOT NULL DEFAULT 1,
    sequence_no INTEGER NOT NULL,
    metadata_json TEXT NOT NULL DEFAULT '{}',
    created_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ', 'now')),
    updated_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ', 'now')),
    FOREIGN KEY (chat_id) REFERENCES chats(id) ON DELETE CASCADE
);
```

### 11.5. Tabela `tools` (Ferramentas Versionadas)
```sql
CREATE TABLE IF NOT EXISTS tools (
    id TEXT PRIMARY KEY NOT NULL,
    registration_id TEXT NOT NULL,
    name TEXT NOT NULL,
    description TEXT NOT NULL,
    version TEXT NOT NULL DEFAULT '1.0.0',
    schema_json TEXT NOT NULL DEFAULT '{}',
    status INTEGER NOT NULL DEFAULT 1,
    created_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ', 'now')),
    updated_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ', 'now')),
    FOREIGN KEY (registration_id) REFERENCES registrations(id) ON DELETE CASCADE
);
```

### 11.6. Tabela `voice_sessions` (Sessões de Voz)
```sql
CREATE TABLE IF NOT EXISTS voice_sessions (
    id TEXT PRIMARY KEY NOT NULL,
    chat_id TEXT NOT NULL,
    status INTEGER NOT NULL DEFAULT 1,
    audio_codec INTEGER NOT NULL DEFAULT 1,
    sample_rate INTEGER NOT NULL DEFAULT 16000,
    channels INTEGER NOT NULL DEFAULT 1,
    metadata_json TEXT NOT NULL DEFAULT '{}',
    created_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ', 'now')),
    updated_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ', 'now')),
    FOREIGN KEY (chat_id) REFERENCES chats(id) ON DELETE CASCADE
);
```

### 11.7. Índices de Desempenho
```sql
CREATE INDEX IF NOT EXISTS idx_shared_rules_reg ON shared_rules(registration_id);
CREATE INDEX IF NOT EXISTS idx_chats_owner ON chats(owner_registration_id);
CREATE INDEX IF NOT EXISTS idx_messages_chat_seq ON messages(chat_id, sequence_no ASC);
CREATE INDEX IF NOT EXISTS idx_tools_reg ON tools(registration_id);
CREATE INDEX IF NOT EXISTS idx_voice_sessions_chat ON voice_sessions(chat_id);
```

---

## 12. Tratamento de Erros e Transacionalidade

### 12.1. Erros Sentinela (`errors.go`)
- `ErrNilDatabase`: Disparado se `NewMigrationEngine` receber um `*sql.DB` nulo.
- `ErrMigrationFailed`: Disparado se a transação DDL falhar e sofrer Rollback.

### 12.2. Garantia de Rollback
As alterações DDL da versão `v1` serão executadas dentro de um bloco transacional Go:
```text
tx, err := db.Begin()
if err != nil { return err }

para cada ddl em ddls_v1:
   _, err := tx.Exec(ddl)
   se err != nil:
      tx.Rollback()
      retornar ErrMigrationFailed empacotado

// Atualiza a versão dentro da transação ou logo após o commit
tx.Exec("PRAGMA user_version = 1;")
tx.Commit()
```

---

## 13. Especificação dos Contratos das Funções Públicas

### 13.1. `NewMigrationEngine(db *sql.DB) (*MigrationEngine, error)`
- Construtor do motor de migração. Valida se `db != nil`.

### 13.2. `migrator.Run() error`
- Executa a verificação do `PRAGMA user_version` e aplica as migrações pendentes em ordem cronológica de versão. Retorna `nil` se o banco estiver atualizado ou após aplicar a migração com sucesso.

### 13.3. `migrator.CurrentVersion() (int, error)`
- Executa a consulta `PRAGMA user_version;` e retorna o número inteiro da versão atual do esquema no banco SQLite.

---

## 14. Critérios de Aceite Objetivos (Checklist de Conclusão)

- [ ] O pacote `core/internal/storage` inclui os novos arquivos `migrations.go`, `migrations_v1.go` e `migrations_test.go`.
- [ ] O construtor `NewMigrationEngine(db)` rejeita referências `nil` com `ErrNilDatabase`.
- [ ] A chamada a `migrator.Run()` em banco zerado cria todas as 6 tabelas (`registrations`, `shared_rules`, `chats`, `messages`, `tools`, `voice_sessions`) e seus 5 índices.
- [ ] A tabela `messages` possui os campos de autoria composta `author_type` e `author_id`.
- [ ] A tabela `tools` possui o campo de versão `version TEXT NOT NULL DEFAULT '1.0.0'`.
- [ ] Após a execução da migration `v1`, a consulta `PRAGMA user_version;` retorna `1`.
- [ ] Executar `migrator.Run()` uma segunda vez em um banco já migrado não executa comandos DDL e retorna `nil` (idempotência).
- [ ] Em caso de erro em qualquer instrução DDL, a transação sofre Rollback e a versão permanece em `0`.
- [ ] A suíte de testes em `migrations_test.go` compila e passa com 100% de sucesso.

---

## 15. Estratégia de Testes Automatizados (`migrations_test.go`)

### 15.1. `TestMigrationEngine_RunV1`
- Instancia o `StorageEngine` em modo `:memory:`, executa `migrator.Run()` e assere:
  - `CurrentVersion()` retorna `1`.
  - Consultas em `sqlite_master` confirmam a existência de todas as 6 tabelas e 5 índices.

### 15.2. `TestMigrationEngine_Idempotency`
- Executa `migrator.Run()` em um banco já migrado e assere que a função retorna `nil` imediatamente sem erros.

### 15.3. `TestMigrationEngine_RollbackOnFailure`
- Simula uma falha em uma instrução DDL sintaticamente malformada em teste isolado e confirma que o banco permanece em `user_version = 0` sem tabelas parciais gravadas.
