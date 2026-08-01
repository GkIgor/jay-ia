# PRD: Redesign UI - Luminous Glass (Frontend C++)

Este documento define os requisitos e o plano de execução para o redesign completo da interface gráfica do frontend C++ (Raylib) do Jay, adotando o design system "Luminous Glass" gerado no Stitch. 

O objetivo é transformar a interface em um produto esteticamente premium, focado em performance, sem alterar a lógica de negócios ou os use cases existentes do Core Go. O trabalho será realizado na branch `feature/luminous-glass-ui`.

---

## 1. Objetivos Visuais e Critérios de Aceite

O objetivo não é apenas aplicar cores, mas construir uma fundação visual escalável. 

**Critérios de Aceite:**
* Todos os painéis principais (Sidebar, TopBar, ChatFeed, InputArea) utilizam o novo sistema arquitetural visual.
* Todos os botões e componentes interativos seguem o novo Design System.
* Não existem componentes visuais remanescentes do tema anterior.
* A interface utiliza exclusivamente os novos Design Tokens.
* Nenhum widget da aplicação utiliza cores, fontes ou tamanhos hardcoded (tudo deve provir do Design System/Theme).

---

## 2. Requisitos Não Funcionais (Performance e Estabilidade)

Sendo uma interface construída sobre Raylib (Immediate Mode/Retained Hybrid), a performance é crítica:
* **Efeitos de Blur:** O processamento de shaders (ex: Glass blur) deve ser estritamente limitado às áreas necessárias e otimizado para não gargalar a GPU.
* **Ciclo de Vida de Shaders:** Shaders devem ser carregados uma única vez na inicialização da aplicação, jamais recriados por frame.
* **Gerenciamento de Fontes:** Fontes e texturas devem ser carregadas e cacheadas na inicialização, proibindo recarregamentos durante a execução.
* **Alocação de Memória:** Evitar alocações dinâmicas (ex: instanciação de novos objetos ou vetores não pré-alocados) durante o método `Render()`.
* **Framerate:** Preservar a taxa de atualização (FPS) da interface em hardware comum, mantendo responsividade instantânea.
* **Escalabilidade Visual:** Preparar o feed de chat para lidar com múltiplas mensagens através de culling/virtualização (desenhar apenas o que está na tela).

---

## 3. UI Foundation & Design System

A interface deixará de usar primitivas estáticas para consumir um sistema de design centralizado, estruturado nos seguintes pilares:

### 3.1. Theme & Design Tokens
* **Color Palette:** Definição clara de tokens semânticos (Primary, Surface, Accents, Text, Background).
* **Sizing & Spacing:** Escala baseada em múltiplos de 8px (4, 8, 12, 16, 24, 32, 40, 48, 64).
* **Radius:** Definição semântica de cantos arredondados (sm, md, lg, xl, full).
* **Elevation:** Sistema de elevação semântica (0 a 4).
* **Tokens Adicionais:** O tema também centraliza a definição de `Icons`, `Animation Tokens` (curvas e durações), `Shadow Tokens` e `Effect Tokens`.

### 3.2. Decoration System
Quase toda a estética visual da interface provém de decoração (como no CSS). O RenderContext não desenha "Glass", mas sim um `BoxDecoration`.
* O sistema implementará objetos de decoração ricos para serem acoplados a qualquer widget de layout (Panels, Containers).
* **Propriedades suportadas:** `Background` (cores sólidas), `Gradient` (linear/radial), `Border` (ex: 1px translúcida no topo), `Glow`, `Shadow` (outer/inner), `Blur` (intensidade do backdrop-filter), e `Opacity`.

### 3.3. Typography
* O sistema deve suportar múltiplas famílias tipográficas (ex: Inter para interface, Geist Mono para código).
* Definição de estilos tipográficos semânticos: `Headline`, `Title`, `Body`, `Label`, `Mono`.

### 3.4. Motion System
* O frontend deve suportar transições padronizadas para micro-interações.
* **Estados Suportados:** Hover fade, opacity transition, pulse, focus ring, slide, ease.
* Componentes interativos não "piscam" de um estado para outro; eles transitam através de durações (durations) predefinidas pelo sistema.

### 3.5. Componentes Core (Styles, não hardcodes)
Os widgets devem ser instanciados com estilos, e não implementações visuais fixas:
* **Button:** Suporte a variantes semânticas (`Primary`, `Secondary`, `Glass`, `Ghost`, `Danger`, `Glow`).
* **Panel:** Componente abstrato de painel governado por um `PanelStyle` (`Default`, `Glass`, `Elevated`, `Transparent`), permitindo que a mesma fundação sirva para painéis de ferramentas, configurações ou sidebars apenas trocando a decoração.
* **ChatBubble:** Representação estruturada de mensagens (`User`, `Assistant`, `Tool`).
* **InputArea / CodeBlock:** Contêineres de interação e visualização de dados complexos.

### 3.6. Componentes de Layout e Estrutura
Para compor a UI de forma flexível e responsiva sem recalcular matemática de tela manualmente (`x`, `y`, `width`, `height`), o sistema utilizará (ou criará) primitivas de layout modernas, similares ao Flexbox/Flutter:
* **Row / Column:** Containers direcionais baseados em Flex para alinhamento horizontal e vertical, suportando espaçamentos (`gap` do Design System) e alinhamentos (`center`, `start`, `end`, `space-between`).
* **Stack:** Componente de sobreposição (Z-index local) para empilhar filhos, como badges sobre ícones.
* **Padding:** Widget decorador para injetar margens internas (`container-padding`, `stack-md`) baseadas na escala semântica.
* **Spacer / Divider:** Componentes de expansão dinâmica flex e linhas de separação baseadas em tokens.

### 3.7. Overlay System
O sistema de sobreposições gerenciará tudo que "flutua" acima da interface principal de forma unificada:
* **Overlays suportados:** `Toast`, `Dialog`, `Tooltip`, `Context Menu`, `Dropdown`, e `Floating Card`.
* O sistema garante captura de foco (modal) e click-away (fechar ao clicar fora).

### 3.8. Scroll System
Para suportar o histórico de mensagens e menus longos, o sistema precisa de uma infraestrutura robusta de rolagem:
* Componentes essenciais: `ScrollableView`, gerenciado por um `ScrollController`.
* **Capacidades:** `Scrollbar` estilizável, rolagem suave (momentum/kinetic scrolling), e gerenciamento de `Viewport` com virtualização (reutilização de widgets fora da tela) para garantir performance in listas infinitas.

---

## 4. Arquitetura de Renderização e Assets

### 4.1. Resource Pipeline (CMake)
* Durante o build, o CMake deverá copiar automaticamente a pasta `assets/` para o diretório de saída da aplicação, preservando sua estrutura integralmente. Isso garante que o binário encontre seus recursos em tempo de execução.
* A estrutura de assets suportada explícita será: `assets/fonts/`, `assets/icons/`, `assets/images/`, `assets/shaders/`, e `assets/themes/`.

### 4.2. Shader Pipeline e Desacoplamento
* O `RenderContext` base da aplicação será responsável estritamente pelas primitivas gráficas e gerenciamento de estado do renderizador (batching, draw calls).
* Efeitos avançados (como o Glass Blur) serão coordenados por uma **ShaderPipeline** dedicada.
* Componentes de alto nível (como um `GlassSidebar`) invocam o pipeline de efeitos de forma desacoplada, garantindo que o `RenderContext` não conheça lógicas específicas de "vidro" ou "glow".

### 4.3. Pipeline de UI e Ciclo de Vida (Lifecycle)
O fluxo da aplicação segue o ciclo de vida estruturado da árvore de componentes:
* **Widget Lifecycle:** `Construct` (Instanciação) -> `Attach` (Conexão à árvore) -> `Measure` / `Layout` (Cálculo geométrico) -> `Update(dt)` (Matemática de animação e estados) -> `Render` (Desenho na tela) -> `Detach` (Remoção da árvore) -> `Destroy`.
* **A Regra de Ouro da Engine (Constraints):** O cálculo geométrico obedece estritamente ao paradigma *Constraints Down, Sizes Up*. O pai envia as restrições (`BoxConstraints`), e o filho decide seu próprio tamanho dentro desse limite. Isso define o núcleo arquitetural do frontend responsivo.

### 4.4. Árvore de Widgets Conceitual (Widget Tree)
A arquitetura base será montada com a seguinte hierarquia visual:
```text
Window
│
├── OverlayRoot (Toasts, Modais flutuantes)
│
└── ShellBase (Background global, Shaders atmosféricos)
    │
    ├── TopBar (Logo superior, pill de conexão, abas)
    │
    ├── Sidebar (Navegação principal, histórico)
    │
    └── ContentArea (Container dinâmico de views)
           ├── AvatarView
           ├── ChatView
           │    ├── ScrollableView (Virtualizado)
           │    │    ├── MessageItem (User)
           │    │    ├── MessageItem (Assistant)
           │    │    └── MessageItem (Tool)
           │    └── InputArea (Input, Attach, Send)
           └── SettingsView
```

---

## 5. Ordem de Implementação (Fases)

Para evitar retrabalho e construir uma fundação sólida, a implementação seguirá rigorosamente esta ordem:

1. **Asset Pipeline:** Configuração do CMake para cópia estruturada de recursos e scaffolding da pasta `assets/`.
2. **Design Tokens & Theme:** Criação das estruturas base de cores, espaçamentos, raio e elevação.
3. **Typography System:** Módulo de carregamento, cacheamento e renderização semântica de fontes.
4. **Shader Pipeline:** Infraestrutura centralizada para carregamento, compilação e aplicação de shaders (Blur, Glow) com foco em performance.
5. **Motion System:** Base de temporização (`dt`) e interpolação para animações e transições de estado.
6. **Core Components (Foundation):** Criação dos widgets primitivos baseados no tema (GlassPanel, Styled Button, Status Chip, InputField).
7. **Layout & Navigation:** Estruturação da nova `Sidebar` e `TopBar` utilizando os novos painéis e botões.
8. **Chat System:** Virtualização da lista de mensagens e implementação dos `ChatBubbles` (User, Assistant, Tool Result).
9. **Shell Integration:** Conectar a nova árvore de widgets (Widget Tree) ao container principal.
10. **Polish & Feedback:** Ajuste fino de animações (pulse, loading, tool transitions) e otimização final de draw calls.
