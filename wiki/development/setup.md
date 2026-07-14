# Ambiente

## Estado atual

O repositório está na Fase 3 de desenvolvimento (Planner Puro e Push IPC estruturado).

## Pré-requisitos e Dependências

Para colaborar no desenvolvimento da Jay, é necessário instalar as seguintes ferramentas no seu ambiente local:

### 1. Backend (Go Core)
- **Go (1.21+)**: Linguagem base do daemon, cli e sdk.
- **socat** (opcional, para testes de socket): `sudo apt install socat`

### 2. Frontend (C++ Avatar)
- **Clang (18+)**: Compilador necessário com suporte nativo a C++23 Modules.
- **clang-tools-18**: Requerido para a varredura e mapeamento de dependências de módulos do C++ (`clang-scan-deps`). Instale via:
  ```bash
  sudo apt install clang-tools-18
  ```
- **CMake (3.28+)**: Gerador de build que suporta nativamente C++ CXX_MODULES.
- **Ninja Build System**: Gerador recomendado para compilação concorrente rápida e compatível com mapeamento de módulos C++. Instale via:
  ```bash
  sudo apt install ninja-build
  ```
- **Raylib & nlohmann-json**: São baixados e compilados automaticamente pelo CMake através de `FetchContent`. No entanto, para compilar a Raylib no Linux, certifique-se de ter as seguintes bibliotecas gráficas de desenvolvimento instaladas:
  ```bash
  sudo apt install libasound2-dev libx11-dev libxrandr-dev libxi-dev libxcursor-dev libxinerama-dev libxkbcommon-dev
  ```

## Leitura obrigatória

Antes de criar setup definitivo, qualquer colaborador deve consultar:

- `wiki/README.md`
- `wiki/index.md`
- `wiki/vision/`
- arquitetura e ADRs relevantes

## Fluxo de início recomendado

1. Ler a wiki base.
2. Identificar a fase e o PRD relacionados ao trabalho.
3. Confirmar se já existe ADR cobrindo a decisão necessária.
4. Só então iniciar estrutura, código ou automação.

## Regra operacional

Se uma tarefa exigir uma decisão ainda não formalizada, o trabalho deve parar no ponto de ambiguidade, solicitar orientação humana e registrar o resultado na wiki antes de seguir.
