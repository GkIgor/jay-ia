# Estrutura do Projeto

## Estado atual

Hoje o projeto é centrado em documentação.

## Estrutura conceitual esperada

```text
jay/
├── core/
├── frontend/
├── cli/
├── sdk/
├── plugins/
└── wiki/
```

## Responsabilidades esperadas

### `core/`

Contém a identidade operacional da Jay.

Responsabilidades esperadas:

- personalidade
- planner
- memória
- conhecimento
- learning
- roteamento de providers
- IPC
- ciclo principal do agente

### `frontend/`

Contém clientes locais da Jay.

Responsabilidades esperadas:

- renderização
- avatar
- áudio
- microfone
- interpretação do protocolo de ações
- mediação com recursos do sistema

### `cli/`

Ferramentas operacionais para inspecionar, iniciar, diagnosticar e operar a Jay.

### `sdk/`

Contratos compartilhados para plugins, providers e integrações.

### `plugins/`

Extensões opcionais do sistema, sempre subordinadas aos contratos do Core.

### `wiki/`

Fonte de verdade do projeto.

## Organização interna desejada

Quando código for adicionado, a estrutura deve refletir fronteiras arquiteturais já definidas na wiki, e não apenas conveniência de curto prazo.

## Anti-padrões

Evitar:

- colocar lógica de Core no frontend
- acoplar OpenClaw ao centro da arquitetura
- misturar memória e conhecimento no mesmo módulo conceitual
- criar pastas guiadas apenas por biblioteca ou framework

## Observação

Essa estrutura representa a direção arquitetural atual e pode evoluir antes da implementação concreta, desde que continue coerente com a wiki e os ADRs.
