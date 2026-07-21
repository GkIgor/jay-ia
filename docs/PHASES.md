# Resumo das Fases do Projeto Jay

Este documento detalha o progresso atual do desenvolvimento do projeto, destacando o que já foi concluído nas Fases 1 e 2, e o que ainda está pendente para o encerramento total de ambas as fases.

## Fase 1: Concluída ✅

A Fase 1 teve como foco o estabelecimento da fundação do backend (Core) do projeto, garantindo uma arquitetura robusta e independente (headless). Os principais itens concluídos foram:

*   **Core Headless em Go:** Implementação do núcleo do agente em Go, operando sem interface gráfica.
*   **Interface MemoryStore:** Criação de uma interface flexível para gerenciamento de memória (armazenamento e recuperação de contexto).
*   **Estrutura Socket IPC (Inter-Process Communication):** Estabelecimento da comunicação entre processos via Sockets, permitindo que o Core se comunique de forma eficiente com o Frontend e outras possíveis integrações.

## Fase 2: Concluída ✅

A Fase 2 focou na criação da interface do usuário (Frontend) utilizando tecnologias de alto desempenho e modernas, conectando-se ao Core desenvolvido na Fase 1. Os principais itens concluídos foram:

*   **Frontend em C++23:** Desenvolvimento da interface gráfica utilizando os recursos mais recentes da linguagem C++, incluindo o uso extensivo de Módulos C++ (`.cppm`).
*   **Gerenciamento de Dependências com CMake:** Utilização do `CMake` e `FetchContent` para integração nativa das bibliotecas dependentes, especificamente `Raylib` (para renderização gráfica) e `nlohmann-json` (para manipulação de dados JSON).
*   **Thread de Cliente IPC Resiliente:** Implementação de uma thread dedicada no frontend para gerenciar a comunicação com o Core via IPC, projetada para ser resiliente a falhas de conexão.
*   **Arquitetura Desacoplada:**
    *   **EventDispatcher:** Sistema de despacho de eventos para roteamento de mensagens internas de forma eficiente.
    *   **Avatar State Machine:** Máquina de estados para gerenciar os comportamentos e animações do avatar de forma independente da lógica principal.

## Fase 3: Concluída ✅

A Fase 3 estabeleceu o sistema de execução nativa de ferramentas e function calling do LLM:

*   **Native Tool Provider:** Abstração de execução nativa de ferramentas (`read_file`, `write_file`, `run_command`, `list_dir`).
*   **Consentimento Multimodal de Segurança:** Fluxo de autorização bloqueante no Frontend Raylib.
*   **LLM Router Agnóstico & Function Calling Multi-Turno:** Loop no Daemon com limites de iteratividade e suporte multi-modelo.

## Fase 3.5: Em Desenvolvimento 🚀 (Consolidação do Core)

A Fase 3.5 consolida o trabalho fundamental acumulado nas Fases 1, 2 e 3, migrando a arquitetura do Core para um **Serviço Orientado a Recursos** estritamente reativo (**Pull-Only**) com estado 100% persistido em **SQLite**:

*   **Persistência SQLite & Transações (ACID):** Eliminação de mapas/slices em memória. Tabelas WAL para `registrations`, `shared_rules`, `chats`, `messages`, `tools` e `voice_sessions`.
*   **Protocolo Pull-Only & Identidades Lógicas (`Registration`):** Desacoplamento entre conexões efêmeras de socket e identidades conhecidas.
*   **Motor de Permissões Declarativo (`SharedRule`):** Avaliador estrito por escopo e padrões de casamento (`EXACT`, `PREFIX`, `WILDCARD`, `REGEX`).
*   **Message Service & Chat Processing Service:** Separação do CRUD de mensagens da execução de IA.
*   **PRD de Referência:** [`wiki/prds/core-resource-architecture.md`](../wiki/prds/core-resource-architecture.md)

## O Que Falta para Completar Totalmente as Fases 1, 2 e 3 🚧

*   **CI/CD (Integração e Entrega Contínuas):** Configuração de pipelines automatizados para testes, build e validação de ambos os lados (Core e Frontend).
*   **Polimento Visual & Scroll:** Melhorias na interface gráfica utilizando os novos componentes de layout de texto e scrollbar.
*   **Testes Unitários para o Frontend e Core SQLite:** Implementação de suítes de testes automatizados para a camada de persistência SQLite, Permission Evaluator e componentes C++.
