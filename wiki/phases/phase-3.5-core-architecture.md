# Fase 3.5: Consolidação da Arquitetura do Core (SQLite & Resource Engine)

## Consolidação Arquitetural do Core

**Objetivo**: Transformar o Jay Core em um serviço estritamente reativo (**Pull-Only**) e orientado a recursos, eliminando toda a persistência temporária em memória e consolidando o armazenamento em banco de dados **SQLite** embarcado com suporte a concorrência, permissões declarativas e contratos IPC fortemente tipados.

---

## Entregas Esperadas

- **Serviço Orientado a Recursos**: Estado 100% persistido em SQLite (WAL mode) com entidades `Registration`, `SharedRule`, `Chat`, `Message`, `Tool` e `VoiceSession`.
- **Comunicação Estritamente Reativa (Pull-Only)**: Eliminação de pushes/broadcasts não solicitados pelo Core. Toda interação é iniciada pelo Consumidor no modelo Requisição-Resposta.
- **Identidades Lógicas Registradas (`Registration`)**: Desacoplamento entre sessões de transporte (sockets IPC) e identidades lógicas dos clientes ("O Core não sabe quem é ninguém").
- **Motor de Permissões Declarativo (`Permission Evaluator`)**: Avaliação baseada no solicitante, escopo do recurso e casamento de padrões (`EXACT`, `PREFIX`, `WILDCARD`, `REGEX`).
- **Autoria Composta de Mensagens (`AuthorType` + `author_id`)**: Suporte a autores do tipo `REGISTRATION`, `AGENT`, `TOOL` e `SYSTEM`.
- **Desacoplamento entre Message Service e Chat Processing Service**: Separação do CRUD/Sync de mensagens da execução da IA/Agente.
- **Ferramentas Versionadas (`version`)**: Suporte a versionamento de capacidades no repositório de ferramentas.

---

## PRD de Referência

O plano detalhado e as especificações completas de contratos IPC e schemas SQL encontram-se em:
- [`prds/core-resource-architecture.md`](../prds/core-resource-architecture.md) / [`prds/phase-3.5-core-architecture.md`](../prds/phase-3.5-core-architecture.md)

---

## Resultado Esperado

Ao final da Fase 3.5, o Jay Core será um serviço extremamente resiliente, ACID, desacoplado de qualquer tecnologia de frontend ou cliente específico, com capacidade de reinicialização transparente e pronto para servir como a base definitiva para as próximas fases.
