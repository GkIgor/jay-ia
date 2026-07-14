# Fase 3: Ações e Comunicação Assíncrona

## Vida e Ações

**Objetivo**: tornar a Jay capaz de agir de forma orquestrada e enviar atualizações de estado em tempo real para a interface sem acoplar a lógica interna ao protocolo IPC.

## Entregas esperadas (Original)
- barramento de ferramentas
- integração com OpenClaw como executor
- acesso controlado a terminal, arquivos e automação
- protocolo de ações para frontend

## Entregas Realizadas (Push IPC e Inteligência Desacoplada)
- **Planner Purificado**: A lógica mental do planner está isolada no Core (`core/internal/planner`), decidindo estados de forma determinística pura sem efeitos colaterais diretos.
- **InternalBus**: Barramento interno baseado em Pub/Sub (Go Channels `[]chan Event`), permitindo múltiplos inscritos escutarem mudanças do Avatar.
- **Daemon Orchestrator**: O daemon (`core/internal/daemon`) atua como executor dos efeitos colaterais gerados pelo plano, emitindo eventos no bus e gerindo delays de simulação de processamento.
- **Broadcaster IPC**: O servidor IPC (`core/internal/ipc`) monitora conexões ativas de forma concorrente (`RWMutex`) e retransmite de forma fortemente tipada (`IPCEvent`) todas as atualizações de estado (`state.changed`, `animation.play`) ao Frontend C++.

## Resultados Obtidos
Jay agora atualiza o frontend de forma assíncrona, transitando graficamente entre estados e executando animações disparadas por comandos recebidos no socket Unix, mantendo total independência de stack técnica.
