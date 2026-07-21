# ADR 0005: Arquitetura de Layout Reativo e Otimização de Churn do Heap loop no Frontend

## Status
Aprovado

## Contexto
O frontend C++ (Raylib) opera por padrão como uma GUI de Modo Imediato (Immediate Mode GUI - IMGUI), desenhando a tela inteira 60 vezes por segundo. No entanto, seções dinâmicas como o feed de mensagens do chat não sofrem modificações a cada quadro.
Recalcular a quebra de linha (`WrapText`), realizar medições de largura de fonte com a GPU (`MeasureTextEx`) e criar substrings temporários (`basic_string::substr` para renderizar o fundo de seleção de texto) a cada frame gera um consumo de CPU/GPU inaceitável e um churn massivo de alocações no heap (milhares de alocações por segundo detectadas pelo Heaptrack).

## Decisão
Fica estabelecido que:
1. **Padrão de Layout Reativo (Flutter-like)**:
   - A fase de **Layout** (computacionalmente pesada) deve ser totalmente separada da fase de **Pintura/Desenho** (leve).
   - O layout das mensagens (`m_renderList`) deve ser recalculado no método `RebuildLayout` apenas se houver mudança de estado:
     - Mudança no tamanho da janela (`screenWidth` ou `screenHeight`).
     - Acréscimo ou remoção de mensagens no feed (`messages.size() != m_prevMsgCount`).
     - Alteração no estado do avatar / carregamento de resposta do agente (`isWaiting`).
     - Ações explícitas de expansão/colapso de blocos (ex: expandir o `ToolGroup`).
   - Se o estado não mudar, a fase de layout é pulada e as coordenadas de desenho pré-calculadas são renderizadas diretamente, com custo de CPU e alocação de Heap nulos no frame loop.
2. **Layout Caching**:
   - Os dados estáticos de largura (`measuredWidth`) e quebra de linhas de texto (`lines`) devem ser cacheados em um mapa indexado pelo ID único de cada mensagem, evitando recálculos em builds futuros.
3. **Otimização de Strings sem Alocação no Heap**:
   - Operações em loop de frames que necessitam ler substrings ou textos dinâmicos (como a renderização de seleção de caracteres para realce azul) não devem usar `std::string::substr`, pois isso aloca memória temporária na Heap a cada iteração.
   - Deve ser utilizado um buffer de caracteres persistente (estilo `std::string measureBuffer`) pré-alocado fora do loop (`reserve(256)`), atualizando seu conteúdo via `assign` para aproveitar o Small String Optimization (SSO) e reuso de capacidade, forçando a contagem de alocações no heap a ser zero.

## Consequências
- A média geral de alocações no Heap da aplicação despencou de **8.196 / segundo** para **2.802 / segundo** (uma queda de ~65% de todo o overhead de execução do programa).
- O consumo de CPU do renderizador de chat tornou-se O(1) na maior parte do tempo de frame ordinário.
- Maior estabilidade e redução drástica do consumo de buffers temporários nos drivers OpenGL da placa de vídeo (Mesa).
