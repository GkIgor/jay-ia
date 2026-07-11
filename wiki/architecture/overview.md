# Visão Geral

## Resumo

A arquitetura da Jay segue um modelo de agente persistente, modular e desacoplado. O Core coordena raciocínio, memória, conhecimento e execução. O frontend representa corpo e interface. Ferramentas externas entram como provedores substituíveis.

## Estrutura Conceitual

```text
Usuário
   ↓
Frontend
   ↓
Jay Core
   ├── Personality
   ├── Conversation
   ├── Planner
   ├── Memory
   ├── Knowledge
   ├── Learning
   ├── LLM Router
   └── Tool Bus
         ├── OpenClaw Provider
         ├── MCP Provider
         ├── Native Provider
         └── HTTP/API Provider
```

## Regras Arquiteturais

- O Core é headless.
- O frontend é cliente do Core.
- A LLM não é fonte oficial de conhecimento.
- O ambiente de execução da Jay é isolado do host.
- A arquitetura não deve depender de uma tecnologia específica de container, frontend ou provider.
