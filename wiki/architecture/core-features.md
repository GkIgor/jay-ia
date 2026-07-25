# Core Feature Inventory & Functional Specifications

Este documento apresenta o inventário funcional e as especificações de comportamento do **Jay Core (Go Daemon)**. Ele descreve as capacidades operacionais do backend e serve como referência para integração de qualquer frontend ou cliente CLI.

---

## 1. Inventário de Recursos e Funcionalidades do Core

### 1.1. Persistência & SQLite Storage Engine
O Jay Core é um serviço persistente livre de estado volátil em memória. Todo o ciclo de vida dos recursos é gerido por um banco de dados SQLite embarcado.

- **Concorrência ACID (WAL Mode)**: Abertura da base de dados com pragmas otimizados de concorrência (`PRAGMA journal_mode = WAL`, `PRAGMA synchronous = NORMAL` e `PRAGMA busy_timeout = 5000`).
- **Integridade Referencial**: Chaves estrangeiras ativas e validadas em nível de banco de dados (`PRAGMA foreign_keys = ON`).
- **Migrações Automáticas**: Sistema nativo de migração incremental baseado em arquivos DDL ordenados, controlado e versionado através do pragma `PRAGMA user_version`.
- **Performance de Busca**: Índices otimizados para ordenação sequencial de mensagens por chat (`idx_messages_chat_seq`) e busca de propriedades de chat/ferramentas.
- **Estratégias de Exclusão**:
  - *Soft Delete* (atualização lógica de status para `DELETED`) para as entidades `Chat` e `Message`, preservando a integridade cronológica de sequências.
  - *Hard Delete* (deleção física no banco de dados) para as entidades `Registration` e `SharedRule` (esta com deleção cascateada `ON DELETE CASCADE` se o registro pai for removido).

---

### 1.2. Protocolo de Comunicação IPC (Estritamente Reativo / Pull-Only)
A comunicação do Core segue o modelo clássico de Requisição-Resposta.

- **Sem Push / Broadcast**: O Core nunca envia dados espontaneamente ou por canais de subscrição ativos. Toda transação de dados é iniciada pelo cliente por meio de requisições de consulta (*Pull*).
- **Envelope Padronizado (JSON-over-Socket)**: Toda mensagem trafega no formato JSON sobre UNIX Domain Socket (ou TCP) envelopada em estruturas estritamente tipadas (`RequestEnvelope` / `ResponseEnvelope`).
- **Garantia de Identidade de Mensagem**: Cada envelope possui um `request_id` (UUID gerado pelo cliente) e um código numérico de ação (`type`) para roteamento.
- **Versionamento de Protocolo**: Controle de compatibilidade integrado através do campo `protocol_version`.
- **Roteamento Interno (Router RPC)**: Mapeamento centralizado de mensagens tipadas para handlers específicos que despacham as operações de banco de dados e serviços.

---

### 1.3. Gestão de Identidades Lógicas (`Registration`)
Desacoplamento físico e lógico entre sessões de transporte de rede e identidades operacionais.

- **Identidade Persistente**: A entidade `Registration` identifica logicamente o cliente (ex: `jay_client_cpp`, `jay_client_cli`). O socket físico pode cair e reconectar livremente sem que a identidade ou seus recursos associados deixem de existir.
- **Agnóstico à Aplicação**: O Core não possui código ou regras específicas para clientes como "Slack" ou "VSCode". Todos são tratados uniformemente como identidades lógicas registradas.

---

### 1.4. Motor de Permissões Declarativo (`Permission Evaluator`)
Avaliação e bloqueio dinâmico de acessos de leitura, escrita e execução sobre recursos de terceiros.

- **Regras de Compartilhamento (`SharedRule`)**: Proprietários de recursos podem declarar regras de compartilhamento definindo alvos, escopos e ações permitidas.
- **Escopos Granulares**: Definição de permissões por escopo de dados (`ALL`, `CHATS`, `MESSAGES`, `TOOLS`).
- **Bitmask de Ações**: Controle por operações binárias (`READ` = 1, `WRITE` = 2, `EXECUTE` = 4, `ADMIN` = 8).
- **Casamento de Padrões (Match Types)**:
  - `EXACT`: Casamento direto de strings.
  - `PREFIX`: Validação por prefixo (ex: `slack_*`).
  - `WILDCARD`: Suporte a globbing simplificado (`*` e `?`).
  - `REGEX`: Avaliação por expressões regulares completas em Go (`regexp`).

---

### 1.5. Serviço de Conversa & Processamento de IA (Planner & LLM Router)
A inteligência da conversa e a tomada de decisões da IA estão isoladas de efeitos colaterais diretos.

- **Decisão Pura (Planner)**: O motor mental (`LLMPlanner`) constrói planos de ação baseados em raciocínio puro, livre de execuções imperativas de arquivos ou rede em seu núcleo.
- **Roteador Multi-Modelos (`llm.Client`)**: Abstração agnóstica que desacopla o Core de SDKs específicos, permitindo chaveamento dinâmico entre provedores como OpenRouter, Gemini API nativa, ou mocks de testes locais.
- **Loop de Raciocínio Multi-Turno**: Capacidade do Daemon de receber chamadas de ferramentas (*tool calls*), despachar sua execução e reinjetar os resultados de volta no assistente de forma cíclica em um loop persistente de planejamento.
- **Limites de Proteção Contra Loops**: Limite estrito de iterações de planejamento (`MaxPlanningIterations = 5`) para evitar consumo de cota e travamento de processamento.
- **Isolamento de Histórico (`ConversationManager`)**: Gestão centralizada do histórico que traduz mensagens do banco SQLite para o formato esperado pelo modelo em tempo de execução.

---

### 1.6. Barramento de Ferramentas Local (`Tools`)
Registro dinâmico de capacidades locais expostas pelos clientes para uso da IA.

- **Autodeclaração Versionada**: Clientes podem registrar capacidades no Core contendo metadados de versão (SemVer) e validação de parâmetros de entrada especificada via JSON Schema.
- **Fluxo de Consentimento de Segurança**: Quando a IA decide executar uma ferramenta crítica, o Core suspende o loop de processamento, cria uma solicitação de permissão pendente e aguarda que o cliente envie uma resposta RPC de aprovação (via clique de mouse ou atalho de teclado). O processamento só prossegue após consentimento explícito.

---

### 1.7. Daemon & CLI
- **Execução Headless (`jayd`)**: Daemon em segundo plano controlado por PID local e socket de comunicação UDS.
- **Cliente Utilitário CLI (`jay`)**: Cliente de console compilável que consome a API do Core para fins de administração, depuração, envio de mensagens e automação de scripts.
