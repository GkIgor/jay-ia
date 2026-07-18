# ADR 0004: Infraestrutura de Depuração de Memória (ASan) e Compilação para Depuradores

## Status
Aprovado

## Contexto
O ecossistema Jay está crescendo em complexidade estrutural (threads para IPC, filas de mensagens, manipulação dinâmica de layouts). Vazamentos de memória silenciosos ou falhas de segmentação no frontend em C++ (Raylib) e condições de corrida no Core (Go) podem ocorrer. Para investigar e prevenir estes problemas, é essencial dispor de um modo de execução com ferramentas de análise de memória ativas e compilações adequadas para depuradores.

## Decisão
Fica estabelecido que:
1. O ecossistema Jay suportará nativamente compilações em modo depuração (Debug Mode) parametrizadas nos Makefiles do projeto.
2. **Frontend C++ (ASan)**:
   - Quando compilado com `CMAKE_BUILD_TYPE=Debug`, o frontend injetará as flags `-fsanitize=address -fno-omit-frame-pointer` do AddressSanitizer (ASan) do Clang.
   - Isso garante que qualquer violação de acesso à memória (use-after-free, buffer overflows, memory leaks) derrube o frontend exibindo um relatório com a linha exata e a pilha de chamadas do erro.
3. **Core Go (Delve / Race)**:
   - Introduzido o alvo de build `build-debug` no Go que passa flags de compilação `-gcflags="all=-N -l"` para desativar otimizações e inlining, permitindo inspeção precisa de variáveis locais com o Delve (dlv).
4. **Facilidade de Uso**:
   - Criação de novos alvos de entrada de build e execução (`make build-debug` e `make run-debug`) nos arquivos Makefile raiz e secundários do projeto, unificando a inicialização de depuração do ecossistema.

## Consequências
- Detecção imediata de vazamentos de memória durante o uso do frontend C++.
- Compatibilidade nativa do Core Go com depuradores.
- Inicialização facilitada com uma única chamada de comando para depuração.
