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

## Configuração do VSCode e C++20 Modules (clangd)

O suporte a Módulos C++20 é sensível à sincronia de versões de AST. Para evitar erros como `ast_file_version_too_old` ou definições de tipos não resolvidas (como `Widget` interpretado como `int*`), siga as diretrizes abaixo:

### 1. Alinhamento de Versões (Recomendado LLVM 20)
Embora a especificação mínima seja LLVM 18, o **LLVM 20** oferece suporte significativamente superior e mais estável para C++20/23 Modules. Certifique-se de instalar as mesmas versões do compilador e do analisador:
```bash
sudo apt install clang-20 clangd-20
```

### 2. Evitar Conflitos de Extensões
Ao usar a extensão **clangd** no VSCode, desative o parser de IntelliSense nativo da extensão **C/C++ (Microsoft)** para que eles não disputem a análise dos cabeçalhos. No seu `.vscode/settings.json`, configure:
```json
{
  "C_Cpp.intelliSenseEngine": "disabled",
  "C_Cpp.default.cppStandard": "c++23",
  "C_Cpp.default.compilerPath": "/usr/bin/clang++-20",
  "C_Cpp.default.compileCommands": "${workspaceFolder}/build/compile_commands.json",
  "clangd.path": "/usr/bin/clangd-20",
  "clangd.arguments": [
    "--compile-commands-dir=${workspaceFolder}/build",
    "--background-index",
    "--clang-tidy",
    "--header-insertion=never",
    "--query-driver=/usr/bin/clang++*"
  ]
}
```

### 3. Diagnóstico e Resolução de Erros de AST (Troubleshooting)
Se o VSCode exibir o erro `clang(ast_file_version_too_old)` ou reclamar que os módulos `.pcm` não puderam ser importados:
1. **Verifique o Compilador do CMake**: O build do CMake deve usar exatamente a mesma versão do `clangd` configurada (ex: Clang 20).
2. **Execute um Rebuild Limpo**: Remova o diretório de build e recrie os arquivos:
   ```bash
   rm -rf build .cache
   CC=clang-20 CXX=clang++-20 cmake -B build -G Ninja -DCMAKE_BUILD_TYPE=Debug
   cmake --build build
   ```
3. **Limpe o Cache do Clangd**: Exclua a pasta `.cache` na raiz do projeto (onde o indexador armazena o histórico do AST).
4. **Reinicie o Serviço**: Abra a Paleta de Comandos (`Ctrl + Shift + P`) e execute `Clangd: Restart language server` ou recarregue a janela (`Developer: Reload Window`).
