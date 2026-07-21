# Especificação Técnica de Implementação: Task 01 — Infraestrutura SQLite Inicial (v4.0 - Padrão Aberto Maduro)

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

## 2. Não-Objetivos (*Non-Goals*)

Esta Task delimita estritamente suas fronteiras funcionais. **Esta Task não pretende**:
- Implementar pool de conexões customizado além do fornecido pela biblioteca padrão do Go.
- Implementar mecanismos de *retry* ou reconexão automática em segundo plano.
- Implementar estratégias de cache de dados em RAM.
- Implementar wrappers ou abstrações arbitrárias sobre o `database/sql`.
- Implementar métricas de telemetria, tracing ou registradores de log ruidosos nesta camada.
- Executar scripts DDL ou mutações de versão de banco.

---

## 3. Contexto Arquitetural

### 3.1. Posicionamento no Backlog de Fases
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

### 3.2. Dependências Futuras
- **Task 02 (Migration Engine)**: Necessita da conexão ativa fornecida pelo `StorageEngine` para ler e atualizar o `PRAGMA user_version` e executar os scripts DDL.
- **Tasks 03 a 07 (Repositórios)**: Necessitam da referência de conexão mantida pelo `StorageEngine` para realizar operações de leitura e escrita.
- **Task 16 (Daemon `jayd`)**: Necessita da função de inicialização e gerenciamento de ciclo de vida do `StorageEngine` na partida da aplicação.

---

## 4. Princípios Arquiteturais Fundamentais

### 4.1. Princípio de Posse Exclusiva (*Ownership*)
> **O `StorageEngine` possui ownership exclusivo sobre o ciclo de vida da conexão SQLite. Nenhum outro componente do sistema pode criar, substituir ou encerrar essa conexão.**

### 4.2. Princípio da Concorrência Delegada
> **O `StorageEngine` não implementa sincronização própria nem utiliza mutexes internos. A segurança para acesso concorrente é fornecida nativamente pela implementação thread-safe de `database/sql` da biblioteca padrão do Go, enquanto o gerenciamento do ciclo de vida da conexão permanece sob responsabilidade exclusiva da função de entrada do daemon.**

### 4.3. Consequência Arquitetural da Exposição do `*sql.DB`
> **A exposição direta do ponteiro `*sql.DB` através do método `engine.DB()` faz parte da arquitetura oficial do Jay Core. Repositórios e componentes de persistência utilizarão diretamente a API da biblioteca padrão de Go (`database/sql`), evitando camadas de abstração secundárias que apenas encapsulem a mesma interface sem agregar comportamento.**

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

## 7. Organização e Coesão dos Arquivos

A implementação deve permanecer estritamente coesa, evitando a fragmentação desnecessária em múltiplos pacotes ou arquivos diminutos:

```text
core/internal/storage/
├── engine.go        # Struct StorageEngine, Config, Open, Close, DB e Pragmas
├── errors.go        # Definição dos erros sentinela de infraestrutura de storage
└── engine_test.go   # Suíte de testes unitários, idempotência, falhas e pragmas
```

---

## 8. Diagrama de Estados do Ciclo de Vida (Modelo Conceitual)

> **Nota de Engenharia**: O diagrama de estados a seguir é um **modelo conceitual de documentação** para descrever o comportamento em tempo de execução. Na implementação em Go, ele é representado de forma simples e de zero-overhead verificando se a referência `db` é nula (`e.db == nil` vs `e.db != nil`), sem a necessidade de declarar enums ou flags de estado redundantes.

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

---

## 9. Invariantes Arquiteturais

Os seguintes invariantes são **garantias absolutas** que a implementação da Task 01 deve satisfazer em tempo de execução:

1. **Invariante de Imutabilidade de Configuração**: O objeto `Config` é copiado por valor na instanciação via `NewStorageEngine(cfg)` e torna-se estritamente imutável durante todo o ciclo de vida do `StorageEngine`.
2. **Invariante de Pós-Condição de Abertura**: Se `Open()` retornar `nil`, o método `DB()` retornará garantidamente um ponteiro não-nulo (`engine.DB() != nil`) e todos os pragmas estarão aplicados.
3. **Invariante de Pós-Condição de Encerramento**: Após a execução de `Close()`, a chamada a `DB()` retornará sempre `nil`.
4. **Invariante de Validade do Ponteiro**: O ponteiro retornado por `DB()` permanece válido apenas enquanto o ciclo de vida do `StorageEngine` estiver no estado `Ready` (ativo).
5. **Invariante de Idempotência de Inicialização e Encerramento**: Chamadas repetidas a `Open()` ou `Close()` em sequência são operações inofensivas (*no-op*) que retornam `nil`.
6. **Invariante de Isolamento de Boot**: A chamada a `Open()` é estritamente single-threaded durante a sequência de inicialização do Daemon em `main()`. Chamadas concorrentes a `Open()` em múltiplas goroutines são proibidas.

---

## 10. Configurações Fixas vs Configuráveis da Infraestrutura

### 10.1. Configurações Configuráveis (`Config` Struct)
Apenas estes dois parâmetros são passíveis de alteração externa via configuração:
- `DatabasePath` (String): Caminho do arquivo SQLite no disco ou `:memory:` para ambiente de testes.
- `BusyTimeoutMs` (Integer): Timeout de bloqueio em ms (Default: `5000`).

### 10.2. Ordem Determinística de Aplicação dos Pragmas Fixos
No momento da execução do `Open()`, os pragmas fixos e imutáveis são aplicados exatamente nesta ordem:

```text
Open()
  │
  ▼
PRAGMA foreign_keys = ON;
  │
  ▼
PRAGMA busy_timeout = 5000;
  │
  ▼
PRAGMA synchronous = NORMAL;
  │
  ▼
PRAGMA journal_mode = WAL;
  │
  ▼
Ready (Conexão configurada com sucesso)
```

---

## 11. Justificativa da Criação Automática de Diretórios

Em sistemas Unix/Linux, se o caminho do banco apontar para um diretório que não existe (ex: `~/.jay/data/jay_core.db`), o driver SQLite nativo falha imediatamente com `unable to open database file`. O `StorageEngine` executa defensivamente `os.MkdirAll(parentDir, 0700)` antes da abertura para evitar falhas no boot.

---

## 12. Especificação dos Contratos das Funções Públicas

### 12.1. `NewStorageEngine(config Config) (*StorageEngine, error)`
- Construtor. Valida a configuração, copia o `Config` por valor e aloca a struct. Retorna `ErrInvalidConfig` se o caminho for vazio.

### 12.2. `engine.Open() error`
- Inicializa a conexão SQLite, garante diretórios pai no SO e aplica os pragmas na ordem determinística. É idempotente. Retorna `ErrStorageInitialization` empacotado em caso de falha.

### 12.3. `engine.Close() error`
- Encerra a conexão `*sql.DB` e limpa a referência (`db = nil`). É idempotente.

### 12.4. `engine.DB() *sql.DB`
- Retorna o ponteiro de conexão `*sql.DB` ativo se o engine estiver inicializado e aberto, ou `nil` caso contrário.

---

## 13. Critérios de Aceite Objetivos

- [ ] A implementação em `core/internal/storage` permanece limpa e coesa.
- [ ] O erro sentinela `ErrStorageInitialization` é utilizado para empacotar falhas de inicialização.
- [ ] O método `DB()` possui a assinatura `DB() *sql.DB`.
- [ ] `Open()` e `Close()` são idempotentes.
- [ ] Os pragmas `foreign_keys=ON`, `busy_timeout=5000`, `synchronous=NORMAL` e `journal_mode=WAL` são aplicados na ordem especificada.
- [ ] A suíte de testes em `engine_test.go` inclui o teste de re-inicialização (`Open -> Close -> Open -> Close`) com 100% de sucesso.

---

## 14. Estratégia de Testes Automatizados (`engine_test.go`)

### 14.1. Testes de Idempotência e Re-inicialização (`TestStorageEngine_OpenClose_Idempotency`)
- Executa `Open()` 2x seguidas -> Ambas retornam `nil`.
- Executa `Close()` 2x seguidas -> Ambas retornam `nil`.
- Executa a sequência completa de re-inicialização: `Open()` -> `Close()` -> `Open()` -> `Close()`, validando que a reconexão funciona perfeitamente sem leaks.

### 14.2. Teste de Validação dos Pragmas (`TestStorageEngine_PragmasValidation`)
- Abre o engine em modo `:memory:`, executa consultas `PRAGMA ...` e assere os valores gravados.

### 14.3. Testes Negativos e de Falha (`TestStorageEngine_NegativeCases`)
- Passa um caminho sem permissão de escrita e assere que `Open()` retorna `ErrStorageInitialization`.
- Passa `DatabasePath` vazio e assere `ErrInvalidConfig`.
