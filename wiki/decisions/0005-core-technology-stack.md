# ADR 0005: Stack Tecnológica do Core, CLI e SDK

## Status

Aceita

## Contexto

A arquitetura da Jay prevê componentes isolados. O `core` (Core Headless) operará como o cérebro da operação (daemon), o `cli` servirá para ferramentas operacionais locais e diagnósticos, e o `sdk` fornecerá contratos compartilhados para providers e plugins. Para implementarmos a **Fase 1 (Vida)**, era necessário definir qual linguagem de programação principal seria usada para esses componentes, respeitando requisitos como:
- Possibilidade de execução persistente como daemon sem grande overhead (baixo consumo de memória).
- Tipagem forte, legibilidade e estabilidade.
- Previsibilidade e facilidade para deploys semânticos / estáticos e isolados.
- Boa base de bibliotecas para sistemas, IPC, networking e execução local segura de ferramentas (OpenClaw).

## Decisão

Foi escolhida a linguagem **Go (Golang)** como a tecnologia primária para desenvolver o `core`, `cli` e `sdk`.

- **Core**: será implementado como um daemon (ex: `jayd`) responsável pelo loop contínuo e coordenação entre Personality, Memory, Knowledge, Planner, etc.
- **CLI**: será construído para conversar via RPC/IPC com o Core, permitindo inspecionar e interagir (ex: `jay status`, `jay config`).
- **SDK**: abrigará as interfaces exportadas, pacotes de conexão e protos genéricos (caso optemos por gRPC ou similar no futuro) para uso interno e por potenciais plugins.

## Consequências

- **Vantagens**: Go provê excelente suporte nativo a concorrência, rotinas assíncronas determinísticas e é compilado num binário único e leve, sendo ideal para rodar isolado num runtime restrito.
- **Limitações**: Sistemas gráficos (UI/Avatar) são menos idiomáticos em Go, no entanto, isso respeita a ADR 0001 (Independência entre Core e Frontend) que dita que o Frontend não estará no Core e possivelmente não usará Go, se comunicando apenas via protocolo.
- **Organização**: O repositório passará a ser um módulo Go na raiz principal, para unificar as dependências enquanto o projeto for monorepo.
