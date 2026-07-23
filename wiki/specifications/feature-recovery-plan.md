# Especificação Técnica — Plano de Recuperação e Portabilidade de Features

**Projeto:** Jay Frontend C++20 — `jay-frontend-cpp`  
**Status:** Aguardando Aprovação  

---

## Árvore de Features & Subfeatures

```text
Jay Frontend
├── Shell & Navegação Global
│   ├── TabBar Superior (Navegação "AVATAR" e "CHAT")
│   │   ├── Alternância por clique de mouse
│   │   └── Atalho global de teclado (Ctrl + Tab)
│   └── Window Resizing Responsivo
│       ├── Ajuste de proporção de layout (35% Avatar / 65% Chat)
│       └── Re-layout automático da árvore de widgets (BoxConstraints::Tight)
│
├── Chat Feature
│   ├── Feed de Mensagens (Message Feed)
│   │   ├── Scroll Controller & Scrollbar Principal
│   │   │   ├── Rolagem por Wheel do mouse (WheelSensChat = 140px)
│   │   │   ├── Rolagem por teclado (PageUp / PageDown)
│   │   │   └── Polegar (Thumb) da Scrollbar com largura de 8px e arrasto contínuo por mouse (Slide Drag)
│   │   ├── Auto-Scroll Inteligente & Contextual
│   │   │   ├── Scroll automático para a base se a mensagem foi enviada pelo Usuário
│   │   │   ├── Scroll mantido na posição atual se o usuário subiu para ler o histórico ao receber resposta do Assistente (margem de 50px do fundo)
│   │   │   └── Rolagem suave até a base quando o usuário estiver no final
│   │   ├── Gestão de Balões de Mensagem (Message Bubbles)
│   │   │   ├── Contêiner Responsivo de Texto (desenho de texto isolado sem deformar o raio dos cantos BubbleCornerRadius = 12px)
│   │   │   ├── Diferenciação semântica visual (Usuário: direita/azul PrimaryDark, Jay: esquerda/escuro JayBubble, Erro: vermelho Danger)
│   │   │   ├── Trim automático de whitespace e quebras de linha nas extremidades das mensagens
│   │   │   └── Suporte a mensagens de status (Piscamento de reticências animadas para estado Thinking)
│   │   ├── Seleção de Texto em Balões
│   │   │   ├── Seleção de texto por clique e arraste com o mouse dentro dos balões do feed
│   │   │   └── Realce visual da porção selecionada com retângulos semi-transparentes em cada linha física
│   │   ├── Ícone de Copiar Mensagem
│   │   │   ├── Botão visual de cópia rápida no canto inferior da bolha
│   │   │   ├── Feedback visual de confirmação ao copiar (ícone muda temporariamente para "check" de sucesso Executing por 2 segundos)
│   │   │   └── Copia o texto bruto original da mensagem para o Clipboard do sistema em UTF-8
│   │   └── Collapsible Tool Groups (Grupos de Ferramentas Retráteis)
│   │       ├── Agrupamento sequencial de ações de ferramentas (ToolAction) em um balão único retrátil (ToolGroupBubble)
│   │       ├── Indicador dinâmico de status por ação (ícone de check verde Success para sucesso, X vermelho Danger para falha)
│   │       ├── Exibição de mensagem de erro técnica por ferramenta falhada
│   │       ├── Cabeçalho retrátil com ícone de triângulo e contador de ações ("N ações executadas")
│   │       ├── Expansão/Colapso ao clicar no cabeçalho
│   │       └── Prevenção de seleção acidental e de falhas de memória em linhas vazias/retráteis
│   │
│   └── Campo de Entrada (Chat Input & Text Editor)
│       ├── Motor de Edição de Texto Multilinha (TextEditor + TextBuffer)
│       │   ├── Suporte a Unicode UTF-8 completo (char32_t) com delimitação correta de palavras (IsWordBoundary)
│       │   ├── Buffer de desfazer e refazer (Undo/Redo via Ctrl+Z, Ctrl+Y, Ctrl+Shift+Z com snapshots lógicos)
│       │   ├── Cursor piscante (Caret) com alinhamento vertical rigoroso (StepY = 18px) sem deriva ou desvio de linha
│       │   └── Preservação da coluna visual preferida (preferredColumn) ao navegar com setas para cima/baixo por linhas de tamanhos diferentes
│       ├── Seleção de Texto no Input
│       │   ├── Seleção por atalho global (Ctrl + A)
│       │   ├── Seleção por arraste de mouse (Mouse Drag) de âncora até cursor
│       │   ├── Seleção por Shift + Setas / Shift + Home / Shift + End
│       │   ├── Seleção por duplo clique (seleciona palavra inteira)
│       │   ├── Seleção por triplo clique (seleciona linha física inteira)
│       │   ├── Seleção por Shift + Click (estende seleção até a posição do clique)
│       │   └── Realce visual azul por linha física parcial sem vazamento do scissor
│       ├── Teclas de Atalho & Edição Rápida
│       │   ├── Inserção e substituição de texto (substitui seleção ativa ao digitar)
│       │   ├── Deleção simples (Backspace, Delete)
│       │   ├── Deleção por palavra (Ctrl + Backspace, Ctrl + Delete)
│       │   ├── Navegação rápida por palavra (Ctrl + Left, Ctrl + Right)
│       │   └── Operações de Clipboard integradas em UTF-8 (Ctrl + C, Ctrl + V, Ctrl + X)
│       ├── Layout & Scroll Interno do Input
│       │   ├── Quebra automática de linha (Word Wrap) por palavras com TextLayout sem modificar o buffer de texto
│       │   ├── Quebra forçada de tokens grandes (Oversized Tokens) como URLs, Base64 e hashes contínuos para não estourar as bordas
│       │   ├── Scrollbar interna vertical automática que surge quando o texto excede o limite de linhas da caixa (aceita mouse wheel WheelSensInput = 54px e arraste de mouse)
│       │   └── Hit-testing de seleção de mouse sob scroll ativo (ajuste exato de coordenada localY + inputScrollOffset)
│       └── Bloqueio & Feedback do Botão Enviar
│           ├── Desabilitação automática do botão "ENVIAR" quando Jay está processando (Thinking ou Executing)
│           └── Feedback visual de botão desabilitado (Theme::Border / Theme::TextSec) e bloqueio da tecla Enter
│
├── Avatar Feature
│   ├── Máquina de Estados Visuais do Avatar
│   │   ├── Estados visuais de núcleo: Idle (Azul Primary), Thinking (Amarelo Warning), Executing (Verde Success), Sleeping (Azul translúcido)
│   │   └── Reatividade a eventos do daemon (state.changed, animation.play - smile, nod, wave)
│   └── Animações & Efeitos Gráficos
│       ├── Animação de pulso contínuo (expansão/contração suave de escala via Tween com Easing::InOutCubic)
│       ├── Halo externo translúcido com canal Alpha suave
│       └── Interpolação suave de cor de núcleo via TweenColor ao alternar de estado
│
└── Permissions Feature
    ├── Modal de Consentimento (Permissions Modal)
    │   ├── Overlay Glassmorphism sobre toda a tela com backdrop semi-transparente bloqueando o fundo
    │   ├── Exibição do título ("Solicitação de Permissão"), prompt explicativo e ref_id do recurso solicitado
    │   └── Botões semânticos de ação: "Permitir" (Verde AllowBtn) e "Negar" (Vermelho DenyBtn)
    └── Teclas de Atalho & Interatividade
        ├── Atalho por teclado rápido: tecla Y para permitir, tecla N para negar
        ├── Envio do envelope IPC permission.response contendo ref_id, allowed e modality ("keyboard" ou "click")
        └── Fechamento imediato do modal e desbloqueio da tela ao responder
```

---

## Tasks de Portabilidade (36 a 41)

- [ ] **Task 36: Rich Text Editor & Keyboard Shortcuts Portability** (`src/features/chat/chat_input_widget.cppm`)
- [ ] **Task 37: Rich Message Bubble & Copy Icon Portability** (`src/features/chat/message_bubble_widget.cppm`)
- [ ] **Task 38: Collapsible Tool Group Widget Portability** (`src/features/chat/tool_group_widget.cppm`)
- [ ] **Task 39: Intelligent Auto-Scroll & Scrollbar Drag Portability** (`src/features/chat/chat_widget.cppm`)
- [ ] **Task 40: Keyboard Shortcuts & Glassmorphism Overlay Portability** (`src/features/permissions/permissions_widget.cppm`)
- [ ] **Task 41: Full Feature Parity & E2E Validation**
