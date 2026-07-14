# Fase 2: Corpo

## Corpo

Objetivo: dar presença visual e vocal à Jay sem acoplar interface ao Core.

## Entregas esperadas
- frontend inicial
- avatar básico
- voz
- animações simples
- estados visuais
- comunicação Core ↔ frontend

## Entregas Realizadas
- **Interface Gráfica com Raylib**: Aplicação nativa C++ em Raylib (`jay-frontend`) que inicializa uma janela de exibição do avatar.
- **C++23 Modules**: Configuração moderna em CMake e Ninja usando C++23 Modules para isolar os componentes (`avatar`, `renderer`, `ipc_client`, `event_dispatcher`).
- **Resiliência IPC**: Cliente socket em C++ rodando em thread de segundo plano com reconexão automática resiliente ao socket do Core.
- **Mapeamento de Estado**: Motor visual do Avatar desenhando círculos coloridos com base no estado mental transmitido pelo socket.

## Resultado
Jay ganhou um corpo visual independente que se comunica via IPC de forma resiliente, totalmente desacoplado da lógica de inteligência do Core.
