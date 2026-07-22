# Convenções e Padrões Arquiteturais da Camada de API e Serviços (`core/internal/api` e `core/internal/service`)

**Documento:** Especificação Arquitetural de Referência  
**Pacotes Go:** `core/internal/api` e `core/internal/service`  
**Escopo:** Padrões imperativos para todos os adaptadores RPC (Handlers) e serviços de aplicação (Services)

---

## 1. Arquitetura em Três Camadas

Toda operação do protocolo IPC no Jay Core atravessa três camadas estritamente desacopladas:

1. **Camada de Protocolo / Roteamento (`core/internal/api/router.go`)**:
   - Ponto de entrada único dos frames JSON do socket IPC.
   - Responsável pela validação de formato do `RequestEnvelope`, isolamento de panics (`recover`), execução de middlewares e tradução automática de erros de domínio Go em códigos `ErrorCode` IPC.

2. **Camada de Adaptadores RPC (`core/internal/api/*_handler.go`)**:
   - Desserializa payloads específicos via `ipc.UnmarshalPayload`.
   - Invoca a camada de `Service` correspondente repassando o contexto e argumentos.
   - Mapeia structs de entidade de domínio para DTOs do SDK usando funções helper puras (`toXDTO`).
   - Constrói e retorna `*ipc.ResponseEnvelope`.
   - **Proibição**: Handlers NUNCA executam I/O de banco diretamente e NUNCA tomam decisões de autorização de recursos.

3. **Camada de Serviço de Aplicação (`core/internal/service/*_service.go`)**:
   - Orquestra os casos de uso do domínio.
   - Aplica autorização declarativa consultando `PermissionEvaluator.Evaluate`.
   - Executa I/O através dos Repositórios do pacote `storage`.
   - Retorna erros Go de domínio padrão (`storage.ErrNotFound`, `storage.ErrAlreadyExists`, `storage.ErrForbidden`, etc.).
   - **Proibição**: Services NUNCA dependem de pacotes de protocolo IPC (`sdk/ipc`) ou adaptadores HTTP/RPC.

---

## 2. Regras de Payloads e DTOs

1. **Payloads Dedicados por Comando**:
   - Cada comando do protocolo possui DTOs de solicitação e resposta exclusivos em `sdk/ipc/messages.go` (ex: `GetRegistrationRequest` / `GetRegistrationResponse`).
   - É proibido reutilizar o payload de um comando de escrita para um comando de leitura.

2. **DTO Mappers Isolados**:
   - A conversão de entidades de banco (`storage.Registration`, `storage.Chat`, `storage.Message`, `storage.Tool`) em DTOs IPC (`ipc.RegistrationDTO`, `ipc.ChatDTO`, etc.) é feita por funções helper privadas no pacote `core/internal/api` (ex: `toRegistrationDTO`).

---

## 3. Garantias de Autorização e Identidade

1. **Identidade do Requisitante**:
   - A identidade do requisitante é extraída do envelope (`req.ClientID`).
2. **Autorização na Camada de Serviço**:
   - A verificação `Evaluator.Evaluate` ocorre exclusivamente dentro dos métodos da camada de `Service`.
   - Operações de escrita em regras de compartilhamento (`UpdateSharedRules`) forçam que o `requesterID` seja a única fonte da verdade para o proprietário das regras.
