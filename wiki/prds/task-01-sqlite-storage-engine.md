# Especificação Técnica de Implementação: Task 01 — Infraestrutura SQLite Inicial (v3.0 - Edição Definitiva)

**Documento:** Technical Implementation Specification  
**Alvo:** `jay-ia/core/internal/storage`  
**Escopo Exclusivo:** Task 01 (StorageEngine, Driver SQLite, Pragmas e Ciclo de Vida da Conexão)  
**Status:** Especificação Técnica Oficial Definitiva  

---

## 1. Objetivo

### 1.1. Propósito da Task
O propósito exclusivo da **Task 01** é construir o gerenciador de infraestrutura e ciclo de vida da conexão com o banco de dados SQLite embarcado (`StorageEngine`) no backend Go do **Jay Core**.

### 1.2. Motivação da Existência
Atualmente, o Jay Core gerencia o seu estado temporário através de estruturas na memória RAM (*in-memory maps/slices*). Para evoluir o sistema para uma arquitetura orientada a recursos com persistência transacional e ACID, o Core necessita de uma camada de infraestrutura única, thread-safe, de alta performance e desacoplada, responsável exclusivamente por abrir, configurar e encerrar a conexão com o SQLite.

### 1.3. Motivação do Isolamento em relação às Próximas Tasks
Esta Task foi deliberadamente isolada de Migrations, DDLs, Repositórios e Regras de Negócio para respeitar o **Princípio da Responsabilidade Única (SRP)**:
1. Garante que problemas de I/O, drivers, caminho de arquivos e pragmas do SQLite sejam resolvidos e validados independentemente.
2. Evita o acoplamento prematuro entre a abertura do banco de dados e a criação de tabelas ou regras de domínio.
3. Fornece um componente infraestrutural reutilizável e estável que será injetado nas tarefas posteriores sem a necessidade de reescrever a lógica de abertura de conexão.

---

## 2. Contexto Arquitetural

### 2.1. Posicionamento no Backlog de Fases
A Task 01 é a **fundação absoluta** de todo o backlog da Fase 3.5. Ela representa a primeira das 17 Tasks planejadas para a reformulação do Jay Core.

```
[Task 01: StorageEngine (Infraestrutura SQLite)]
                         │
                         ▼
[Task 02: Migration Engine & DDL Base]
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

### 2.2. Dependências Futuras
- **Task 02 (Migration Engine)**: Necessita da conexão ativa fornecida pelo `StorageEngine` para ler e atualizar o `PRAGMA user_version` e executar os scripts DDL.
- **Tasks 03 a 07 (Repositórios)**: Necessitam da referência de conexão mantida pelo `StorageEngine` para realizar operações de leitura e escrita.
- **Task 16 (Daemon `jayd`)**: Necessita da função de inicialização e gerenciamento de ciclo de vida do `StorageEngine` na partida da aplicação.

---

## 3. Responsabilidades

### 3.1. O que Pertence à Task 01
- Gerenciar o ciclo de vida da conexão SQLite (`NewStorageEngine`, `Open`, `Close`, `DB`).
- Garantir comportamento idempotente para `Open()` e `Close()`.
- Registrar e carregar o driver Go puro de SQLite (`modernc.org/sqlite`).
- Aplicar de forma determinística os pragmas imutáveis de infraestrutura (`journal_mode=WAL`, `foreign_keys=ON`, `busy_timeout=5000`, `synchronous=NORMAL`).
- Garantir a criação defensiva de diretórios pai no sistema de arquivos para evitar falhas de inicialização do SQLite.
- Tratar e empacotar erros de I/O e conexão específicos de infraestrutura.

### 3.2. O que NÃO Pertence à Task 01
- **NENHUMA** criação de tabelas ou execução de comandos DDL (`CREATE TABLE`, `CREATE INDEX`).
- **NENHUMA** execução de migrations ou leitura/mutação de `PRAGMA user_version`.
- **NENHUMA** definição de structs de repositório (`RegistrationRepository`, `ChatRepository`, etc.).
- **NENHUMA** definição de modelos de domínio ou entidades (`Chat`, `Message`, `Registration`).
- **NENHUMA** execução de queries de negócio (`SELECT`, `INSERT`, `UPDATE`, `DELETE`).
- **NENHUMA** dependência ou contrato com o protocolo IPC ou transporte Unix Domain Socket.
- **NENHUMA** regra de IA, LLM, prompt ou agentes.

---

## 4. Princípios Arquiteturais Fundamentais

### 4.1. Princípio de Posse Exclusiva (*Ownership*)
> **O `StorageEngine` possui ownership exclusivo sobre o ciclo de vida da conexão SQLite. Nenhum outro componente do sistema pode criar, substituir ou encerrar essa conexão.**

### 4.2. Princípio da Concorrência Delegada
> **O `StorageEngine` não implementa sincronização própria nem utiliza mutexes internos. A segurança para acesso concorrente é fornecida nativamente pela implementação thread-safe de `database/sql` da biblioteca padrão do Go, enquanto o gerenciamento do ciclo de vida da conexão permanece sob responsabilidade exclusiva da função de entrada do daemon.**

---

## 5. Orquestração do Boot da Aplicação e Migrations

Uma das definições mais importantes para a evolução do Jay Core é o papel de cada componente durante o boot do processo.

### 5.1. Quem chama o Migration Engine?
O `StorageEngine` **NÃO** chama e **NÃO** conhece o `MigrationEngine` (Task 02). A responsabilidade de orquestrar a sequência de partida pertence à função principal do Daemon (`main()` em `core/cmd/jayd/main.go`).

```text
               main() [Daemon jayd]
                 │
                 ├── 1. Instancia e Abre a Infraestrutura
                 │      engine := storage.NewStorageEngine(cfg)
                 │      if err := engine.Open(); err != nil { exit(1) }
                 │
                 ├── 2. Executa as Migrações de Esquema (Task 02)
                 │      migrator := storage.NewMigrationEngine(engine.DB())
                 │      if err := migrator.Run(); err != nil { exit(1) }
                 │
                 ├── 3. Instancia os Repositórios e Serviços (Tasks 03-15)
                 │      chatRepo := storage.NewChatRepository(engine.DB())
                 │
                 └── 4. Executa o Loop Principal e Encerramento Gracioso
                        defer engine.Close()
```

---

## 6. Arquitetura das Camadas

### 6.1. Visão Geral
A camada de armazenamento é estruturada de forma que o restante da aplicação nunca interaja diretamente com drivers internos de baixo nível. O `StorageEngine` atua como o único guardião da conexão.

```text
+-------------------------------------------------------------+
|                     Camadas Superiores                      |
|       (Migration Engine / Repositórios / Daemon Jayd)       |
+-------------------------------------------------------------+
                              │
                              │ Consome *sql.DB via engine.DB()
                              ▼
+-------------------------------------------------------------+
|               StorageEngine (core/internal/storage)         |
|  - Gerencia Ciclo de Vida Idempotente (Open/Close)          |
|  - Aplica Pragmas Fixos de Infraestrutura                   |
|  - Mantém a referência única de *sql.DB                     |
+-------------------------------------------------------------+
                              │
                              │ Utiliza a biblioteca padrão Go
                              ▼
+-------------------------------------------------------------+
|                     database/sql (Stdlib)                   |
+-------------------------------------------------------------+
                              │
                              │ Registra o driver CGO-Free
                              ▼
+-------------------------------------------------------------+
|                 modernc.org/sqlite (Driver)                 |
+-------------------------------------------------------------+
                              │
                              │ Operações de Arquivo/I/O
                              ▼
+-------------------------------------------------------------+
|                 SQLite Engine & Banco (.db)                 |
+-------------------------------------------------------------+
```

---

## 7. Organização dos Diretórios e Arquivos

### 7.1. Estrutura Enxuta Proposta
Para evitar fragmentação desnecessária e manter a coesão máxima do pacote, a estrutura conterá **exatamente 3 arquivos**:

```text
core/internal/storage/
├── engine.go        # Struct StorageEngine, Configuração, Open, Close, DB e Pragmas
├── errors.go        # Definição dos erros sentinela de infraestrutura de storage
└── engine_test.go   # Suíte de testes unitários, idempotência, falhas e pragmas
```

### 7.2. Detalhamento dos Arquivos

#### `engine.go`
- **Responsabilidade**: Concentrar a struct `Config`, a struct `StorageEngine`, o construtor `NewStorageEngine`, os métodos de ciclo de vida (`Open`, `Close`, `DB`) e a rotina privada de aplicação dos pragmas fixos.
- **O que NUNCA deve existir nele**: Métodos CRUD de tabelas específicas, lógica de migração de versões ou parsers de protocolo IPC.

#### `errors.go`
- **Responsabilidade**: Definir variáveis de erro sentinela exportadas (`ErrInvalidConfig`, `ErrDatabaseOpenFailed`, `ErrPragmaFailed`).
- **O que NUNCA deve existir nele**: Erros de regra de negócio ou de protocolo.

#### `engine_test.go`
- **Responsabilidade**: Validar de forma automatizada se a abertura, aplicação de pragmas, idempotência, testes negativos de falhas e encerramento funcionam perfeitamente.

---

## 8. Diagrama de Estados do Ciclo de Vida (*Lifecycle State Machine*)

O `StorageEngine` opera sob quatro estados bem definidos:

```text
 [Uninitialized] ── NewStorageEngine() ──> [Created]
                                             │  │
                                      Open() │  │ Open() [Falha]
                                   [Sucesso] │  │
                                             ▼  ▼
                                          [Ready]  [Failed]
                                             │
                                          Close()
                                             │
                                             ▼
                                          [Closed]
```

### 8.1. Transições e Comportamentos Detalhados

| Estado Origem | Operação | Estado Destino | Comportamento / Retorno |
|---|---|---|---|
| *None* | `NewStorageEngine()` | `Created` | Instancia a struct e valida a configuração. `db` permanece `nil`. |
| `Created` | `Open()` (Sucesso) | `Ready` | Abre a conexão no SQLite, aplica pragmas fixos e armazena `*sql.DB`. |
| `Created` | `Open()` (Falha) | `Failed` | Falha no I/O ou Pragma. `db` permanece `nil` e retorna `error`. |
| `Failed` | `Open()` | `Ready` ou `Failed` | **Permite re-tentativa**: Executa novamente o processo de abertura. |
| `Ready` | `Open()` | `Ready` | **Idempotente**: Retorna `nil` imediatamente sem reabrir. |
| `Ready` | `DB()` | `Ready` | Retorna o ponteiro `*sql.DB` ativo garantidamente não-nulo. |
| `Ready` | `Close()` | `Closed` | Fecha a conexão `*sql.DB` e limpa a referência (`db = nil`). |
| `Closed` | `Close()` | `Closed` | **Idempotente (No-Op)**: Retorna `nil` imediatamente. |
| `Closed` | `DB()` | `Closed` | Retorna `nil` diretamente. |
| `Created` / `Failed` | `DB()` | Estado Mantido | Retorna `nil` diretamente. |

---

## 9. Invariantes Arquiteturais

Os seguintes invariantes são **garantias absolutas** que a implementação da Task 01 deve satisfazer em tempo de execução:

1. **Invariante de Pós-Condição de Abertura**: Se `Open()` retornar `nil`, o método `DB()` retornará garantidamente um ponteiro não-nulo (`engine.DB() != nil`) e todos os pragmas estarão aplicados.
2. **Invariante de Pós-Condição de Encerramento**: Após a execução de `Close()`, a chamada a `DB()` retornará sempre `nil`.
3. **Invariante de Assinatura Limpa (`DB() *sql.DB`)**: O método `DB()` não retorna erro. Ele é uma consulta de propriedade simples do engine. Chamadores usam a invariante de que se o engine foi inicializado pelo `main()`, `engine.DB()` entrega a conexão pronta.
4. **Invariante de Idempotência de Inicialização**: Chamar `Open()` múltiplas vezes em um engine em estado `Ready` retorna `nil` sem causar vazamentos.
5. **Invariante de Idempotência de Encerramento**: Chamar `Close()` múltiplas vezes em um engine em estado `Closed` retorna `nil` sem causar pânicos.
6. **Invariante de Isolamento de Esquema**: O `StorageEngine` nunca executa instruções DDL (`CREATE TABLE`, `ALTER TABLE`, `DROP TABLE`) e nunca modifica o `PRAGMA user_version`.

---

## 10. Discussão Arquitetural: Assinatura `DB() *sql.DB` e Exposição de Conexão

### 10.1. A Questão da Assinatura sem Erro: `DB() *sql.DB`
> *Por que utilizar `DB() *sql.DB` sem retornar `( *sql.DB, error )`?*

#### Racional da Escolha:
Em Go idiomático, se o ciclo de vida garante que `Open()` é executado com sucesso durante o boot da aplicação em `main()`, o `StorageEngine` assume o estado `Ready`. Tratar erros em toda e qualquer chamada a `DB()` dentro dos repositórios criaria código defensivo ruidoso e desnecessário (ex: `db, _ := engine.DB()` ou `if err != nil { panic(err) }`).

Com a assinatura `DB() *sql.DB`:
- Se o engine estiver no estado `Ready`, retorna `*sql.DB`.
- Se o engine estiver `Closed` ou não inicializado, retorna `nil`.

### 10.2. A Questão da Exposição do `*sql.DB`
O `StorageEngine` expõe o `*sql.DB` padrão do Go para que Repositórios e o Migration Engine possam utilizá-lo nativamente com suporte a transações (`db.BeginTx`), sem a necessidade de criar métodos wrapper customizados que duplicariam a standard library do Go.

---

## 11. Configurações Fixas vs Configuráveis da Infraestrutura

Para evitar ambiguidade sobre o que pode ou não ser alterado via arquivos de configuração ou flags, o `StorageEngine` define uma separação rígida:

### 11.1. Configurações Configuráveis (`Config` Struct)
Apenas estes dois parâmetros são passíveis de alteração externa:
- `DatabasePath` (String): Caminho do arquivo SQLite no disco ou `:memory:` para ambiente de testes.
- `BusyTimeoutMs` (Integer): Timeout de bloqueio em ms (Default: `5000`).

### 11.2. Configurações Fixas e Imutáveis da Infraestrutura
Estes pragmas são **HARDCODED** na implementação privada do `StorageEngine` e nunca serão expostos para alteração por configuração:
- `journal_mode = WAL`: Inegociável para garantir concorrência de leituras sem bloqueio.
- `foreign_keys = ON`: Inegociável para garantir integridade referencial no SQLite.
- `synchronous = NORMAL`: Inegociável para garantir balanço ótimo de performance e segurança em modo WAL.

---

## 12. Justificativa da Criação Automática de Diretórios

### 12.1. O Problema no SQLite Native
Em sistemas Unix/Linux, se o desenvolvedor ou usuário especificar um caminho para o banco em uma pasta que não existe (ex: `~/.jay/data/jay_core.db`), o driver SQLite nativo falha imediatamente com o erro genérico `unable to open database file`.

### 12.2. A Solução no `StorageEngine`
Antes de invocar `sql.Open()`, se `DatabasePath` for um caminho de arquivo físico (e não `:memory:`), o `StorageEngine` executa a checagem e criação das pastas pai (`os.MkdirAll(parentDir, 0700)`).

**Justificativa**: Esta é uma conveniência defensiva de infraestrutura I/O. Ela evita que a aplicação falhe no boot por falta de pastas pré-criadas no ambiente do usuário, mantendo o isolamento de filesystem dentro do pacote `storage`.

---

## 13. Escolha do Driver SQLite: `modernc.org/sqlite`

- **Ausência Total de CGO (CGO-Free / Pure Go)**: O driver é uma transpilação direta do código C do SQLite para Go puro.
- **Compilação Cruzada Simplificada**: Permite compilar o Jay Core para qualquer plataforma (`GOOS=linux`, `GOOS=darwin`, `GOOS=windows`) sem necessidade de toolchains C (`gcc`).
- **Segurança e Estabilidade**: Evita incompatibilidades de bibliotecas dinâmicas C em containers ou no SO do usuário.

---

## 14. Especificação dos Contratos das Funções Públicas

### 14.1. `NewStorageEngine(config Config) (*StorageEngine, error)`
- Construtor. Valida a configuração e aloca a struct. Retorna `ErrInvalidConfig` se o caminho for vazio.

### 14.2. `engine.Open() error`
- Inicializa a conexão SQLite, garante diretórios pai no SO e aplica os pragmas fixos. É idempotente (retorna `nil` se em estado `Ready`). Altera para `Failed` se o I/O ou pragma falhar.

### 14.3. `engine.Close() error`
- Encerra a conexão `*sql.DB` e limpa a referência (`db = nil`). É idempotente (retorna `nil` se em estado `Closed`).

### 14.4. `engine.DB() *sql.DB`
- Retorna o ponteiro de conexão `*sql.DB` ativo se em estado `Ready`, ou `nil` caso contrário.

---

## 15. Futuras Responsabilidades e Limites Estritos

Para evitar desvio de escopo (*scope creep*) à medida que o projeto evoluir:

1. **Task 02 (Migration Engine)**: Poderá chamar `engine.DB()` para ler `PRAGMA user_version` e criar tabelas. O `StorageEngine` **NUNCA** será modificado para conter código de migração.
2. **Tasks 03-07 (Repositórios)**: Receberão `engine.DB()` para executar queries CRUD. O `StorageEngine` **NUNCA** será modificado para adicionar métodos de repositório.
3. **Task 01 é Imutável pós-aprovação**: Concluída esta Task, seu código em `engine.go` permanecerá estável durante todas as fases seguintes.

---

## 16. Critérios de Aceite Objetivos (Checklist de Conclusão)

- [ ] O pacote `core/internal/storage` contém **exatamente 3 arquivos**: `engine.go`, `errors.go` e `engine_test.go`.
- [ ] O método `DB()` possui a assinatura exata `DB() *sql.DB`.
- [ ] O método `IsClosed()` **NÃO** existe.
- [ ] Nenhum `sync.RWMutex` é utilizado na struct `StorageEngine`.
- [ ] `Open()` é totalmente idempotente (retorna `nil` se já `Ready`).
- [ ] `Close()` é totalmente idempotente (retorna `nil` se já `Closed`).
- [ ] Os pragmas `journal_mode=WAL`, `foreign_keys=ON`, `busy_timeout=5000` e `synchronous=NORMAL` são aplicados no `Open()`.
- [ ] Se a abertura de I/O ou Pragma falhar, o estado passa para `Failed` e `db` permanece `nil`.
- [ ] O engine cria as pastas pai no sistema de arquivos para caminhos físicos.
- [ ] A suíte de testes em `engine_test.go` compila e passa com 100% de sucesso (incluindo testes de falhas negativas).

---

## 17. Estratégia de Testes Automatizados (`engine_test.go`)

### 17.1. Teste de Idempotência (`TestStorageEngine_OpenClose_Idempotency`)
- Executa `Open()` 2x seguidas -> Valida que ambas retornam `nil`.
- Executa `Close()` 2x seguidas -> Valida que ambas retornam `nil`.

### 17.2. Teste de Validação dos Pragmas (`TestStorageEngine_PragmasValidation`)
- Abre o engine em modo `:memory:`, executa consultas `PRAGMA ...` e assere os valores: `journal_mode` (wal/memory), `foreign_keys` == 1, `busy_timeout` == 5000.

### 17.3. Testes Negativos e de Falha (`TestStorageEngine_NegativeCases`)
- **Caminho sem Permissão**: Passa um caminho sem permissão de escrita (ex: `/proc/invalid_path/db.sqlite`) e assere que `Open()` retorna erro e transiciona para o estado `Failed`.
- **Configuração Inválida**: Passa `DatabasePath` vazio e assere retorno `ErrInvalidConfig`.
- **Acesso pós-Close**: Executa `Close()` e assere que `engine.DB()` retorna `nil`.

### 17.4. Teste de Criação de Estrutura de Pastas (`TestStorageEngine_CreateParentDirectories`)
- Passa um caminho apontando para subpastas inexistentes em `t.TempDir()` e assere que a abertura cria os diretórios automaticamente.
