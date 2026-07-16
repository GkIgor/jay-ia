# Fase 3: Ações e Comunicação Assíncrona

## Vida e Ações

**Objetivo**: tornar a Jay capaz de agir de forma orquestrada e enviar atualizações de estado em tempo real para a interface sem acoplar a lógica interna ao protocolo IPC.

## Entregas esperadas (Original)
- barramento de ferramentas
- integração com OpenClaw como executor (Postergado via ADR 0008)
- acesso controlado a terminal, arquivos e automação
- protocolo de ações para frontend

## Entregas Realizadas
- **Planner Purificado**: A lógica mental do planner está isolada no Core (`core/internal/planner`), decidindo estados de forma determinística pura sem efeitos colaterais diretos.
- **InternalBus**: Barramento interno baseado em Pub/Sub (Go Channels `[]chan Event`), permitindo múltiplos inscritos escutarem mudanças do Avatar.
- **Daemon Orchestrator**: O daemon (`core/internal/daemon`) atua como executor dos efeitos colaterais gerados pelo plano, emitindo eventos no bus e gerindo delays de simulação de processamento.
- **Broadcaster IPC**: O servidor IPC (`core/internal/ipc`) monitora conexões ativas de forma concorrente (`RWMutex`) e retransmite de forma fortemente tipada (`IPCEvent`) todas as atualizações de estado (`state.changed`, `animation.play`) ao Frontend C++.
- **Native Tool Provider**: Abstração de Tool Provider implementada via Native Provider para a Fase 1 (ADR 0008), permitindo comandos locais, leitura/escrita de arquivos e chamadas de API sem introduzir a complexidade inicial do OpenClaw.
- **Consentimento Multimodal de Segurança**: Fluxo de bloqueio de ferramenta que envia o evento `request.permission` (contendo `prompt` e `ref_id`) e aguarda a resposta do Frontend via socket. O frontend visual em Raylib desenha o modal com botões coloridos, aceitando tanto input de teclado (`Y`/`N`) quanto cliques do mouse em "Permitir"/"Negar", respondendo de forma tipada com a modalidade de fato usada.
- **Testes Unitários de Integração**: Teste unitário completo em `daemon_test.go` validando o fluxo de bloqueio, emissão do prompt, resposta do mock cliente via IPC e desbloqueio seguro do Daemon.

## Resultados Obtidos
Jay agora é capaz de executar ações locais sob rígido e granular consentimento de segurança do usuário. As decisões permanecem no Core e a interface gráfica (Raylib) reage de forma desacoplada exibindo o prompt de permissão, garantindo interatividade multimodal (teclado e clique de mouse) e mantendo total independência de stack técnica.
