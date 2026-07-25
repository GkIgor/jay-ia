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

## 3. Filosofia da GUI Engine

> A GUI Engine fornece capacidades reutilizáveis. A aparência de um componente deve ser resultado da composição entre Layout, Theme, Decoration, Motion e Effects, e **nunca** de lógica visual embutida em widgets específicos. Se houver vontade de criar um "GlassButton", a fundação obriga a questionar: *"Isso é realmente um novo widget, ou apenas o widget genérico `Button` configurado com um `ButtonStyle::Glass`?"*.

---

## 4. UI Foundation & Design System

A interface consumirá um sistema de design centralizado, estruturado nos seguintes pilares:

### 4.1. Theme & Design Tokens
* **Color Palette:** Definição clara de tokens semânticos (Primary, Surface, Accents, Text, Background).
* **Sizing & Spacing:** Escala baseada em múltiplos de 8px (4, 8, 12, 16, 24, 32, 40, 48, 64).
* **Radius:** Definição semântica de cantos arredondados (sm, md, lg, xl, full).
* **Elevation:** Sistema de elevação semântica (0 a 4).
* **Tokens Adicionais:** O tema também centraliza a definição de `Icons`, `Animation Tokens` (curvas e durações), `Shadow Tokens` e `Effect Tokens`.

### 4.2. Decoration System
O RenderContext não desenha "Glass", mas sim um `BoxDecoration`.
* O sistema implementa objetos de decoração ricos para serem acoplados a qualquer widget.
* **Importante:** As `Decorations` são imutáveis e reutilizáveis, evitando realocações a cada frame.
* **Propriedades suportadas:** `Background` (cores sólidas), `Gradient` (linear/radial), `Border` (ex: 1px translúcida no topo), `Glow`, `Shadow` (outer/inner), `Blur` (intensidade do backdrop-filter), e `Opacity`.

### 4.3. Typography
* Múltiplas famílias tipográficas (Inter para interface, Geist Mono para código).
* Estilos tipográficos semânticos: `Headline`, `Title`, `Body`, `Label`, `Mono`.

### 4.4. Motion System
* Infraestrutura orientada a objetos para animações baseadas no tempo, contendo: `AnimationController`, `Tween`, `Curve` (funções de interpolação) e `Duration`.
* Componentes interativos não "piscam" de um estado para outro; eles fluem naturalmente.

### 4.5. Layout System
Primitivas de infraestrutura estritamente voltadas a posicionamento e espaçamento (sem responsabilidade visual):
* Contêineres de alinhamento: `Row`, `Column`, `Stack`, `Padding`, `Align`, `Spacer`, `SizedBox`, e `Divider`.

### 4.6. Componentes Core (Styles e States)
Os componentes visuais puros reagem aos **Estados** para inferir sua aparência:
* **WidgetState:** Todo componente obedece ao paradigma de estado (`Normal`, `Hover`, `Pressed`, `Focused`, `Disabled`, `Selected`).
* **Button:** Baseado em variantes (`Primary`, `Secondary`, `Glass`, `Ghost`, `Danger`, `Glow`).
* **Panel:** Componente base guiado pelo `PanelStyle` (`Default`, `Glass`, `Elevated`, `Transparent`).
* **ChatBubble / InputArea:** Componentes estruturados de domínio.

### 4.7. Focus System
* O `FocusManager` captura de forma unificada os eventos de teclado (TAB, atalhos, navegação por setas) garantindo a acessibilidade e retendo o estado de `Focused` (WidgetState) para injeção visual.

### 4.8. Overlay System
O sistema gerenciará tudo que flutua acima da interface principal no eixo Z global:
* **Overlays suportados:** `Toast`, `Dialog`, `Tooltip`, `Context Menu`, `Dropdown`, e `Floating Card`.

### 4.9. Scroll System
Para suportar o histórico de mensagens e menus longos, o sistema provê uma infraestrutura de rolagem de alta fidelidade:
* **Componentes:** `ScrollableView` governado por um `ScrollController`.
* **Capacidades:** `Scrollbar` customizável, rolagem com Inércia (`Momentum`), e gerenciamento de `Viewport` utilizando `Virtualização` (reutilização de widgets em background) para escalabilidade infinita de mensagens.

---

## 5. Arquitetura de Renderização e Assets

### 5.1. Resource Pipeline (CMake)
* Durante o build, o CMake deverá copiar automaticamente a pasta `assets/` para o diretório de saída da aplicação, preservando sua estrutura integralmente. Isso garante que o binário encontre seus recursos em tempo de execução.
* A estrutura explícita será: `assets/fonts/`, `assets/icons/`, `assets/images/`, `assets/shaders/`, e `assets/themes/`.
* O sistema de **Theme** será responsável por carregar instâncias de recursos (como fontes ou ícones) através do Asset Pipeline, atuando de forma genérica em vez de conhecer caminhos absolutos no disco.

### 5.2. Shader Pipeline e Desacoplamento
* O `RenderContext` base da aplicação será responsável estritamente pelas primitivas gráficas e gerenciamento de estado do renderizador (batching, draw calls).
* Efeitos avançados (como o Glass Blur) serão coordenados por uma **ShaderPipeline** dedicada.
* Componentes (como uma `Sidebar` instanciando um painel `Glass`) invocam o pipeline de efeitos de forma desacoplada, garantindo que o `RenderContext` não conheça lógicas específicas de materiais ou glow.

### 5.3. Pipeline de UI e Ciclo de Vida (Lifecycle)
O fluxo da aplicação segue o ciclo de vida estruturado da árvore de componentes:
* **Widget Lifecycle:** `Construct` (Instanciação) -> `Attach` (Conexão à árvore) -> `Measure` / `Layout` (Cálculo geométrico) -> `Update(dt)` (Matemática de animação e estados) -> `Render` (Desenho na tela) -> `Detach` (Remoção da árvore) -> `Destroy`.
* **A Regra de Ouro (Constraints):** O cálculo obedece estritamente ao *Constraints Down, Sizes Up*. O pai envia as restrições (`BoxConstraints`), e o filho decide seu próprio tamanho.
  * **Atenção:** Widgets *nunca* acessam diretamente o tamanho da janela (ex: `GetScreenWidth()`) para calcular sua geometria. Toda informação espacial provém das Constraints recebidas do pai, garantindo testabilidade e modularidade.

### 5.4. Árvore de Widgets Conceitual (Widget Tree)
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

## 6. Ordem de Implementação (Fases)

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

---

## 7. Observação Crítica: Complexidade Visual (Browser vs Raylib)

Alcançar uma interface com visual idêntico ao renderizado por navegadores modernos exige um esforço gráfico significativamente maior do que o desenvolvimento web tradicional. Ao contrário do navegador, que encapsula décadas de matemática gráfica (CSS `backdrop-filter`, `box-shadow` com antialiasing, etc.), o Raylib é um framework de baixo nível que requer que construamos essas capacidades do zero para evitar uma "aparência dura de 2005":

1. **Antialiasing e Geometria (SDFs):** O desenho nativo de primitivas geométricas pode apresentar serrilhados. Será necessário o uso de matemática de SDF (*Signed Distance Fields*) ou multi-sampling acoplado a shaders para obter cantos perfeitamente arredondados e nítidos.
2. **Efeito Glass (Multi-pass Blur):** Não há suporte nativo para desfoque de fundo (blur) no Raylib. A implementação exige desenhar os elementos da tela em uma textura intermediária (FBO / `RenderTexture2D`), submetê-la a um Fragment Shader GLSL de desfoque Gaussiano ou Kawase, e então compor a cena.
3. **Glows e Sombras Suaves:** A ausência de box-shadows com decaimento exponencial orgânico exige a simulação via geração de malhas com gradientes (vertex colors) ou shaders dedicados de espalhamento de luz para criar as emissões (glows) naturais dos botões.
4. **Física de Interface (Motion):** Interfaces modernas reagem com inércia e elasticidade. As interações (`hover`, `focus`, aberturas) devem ser controladas pelo tempo (`dt`) usando funções de *easing* (Curvas Bezier, molas/spring physics) calculadas a cada frame, substituindo as trocas estáticas instantâneas.

Essa complexidade intrínseca valida e exige que as fases de **Fundação** (Shader Pipeline e Motion System) sejam concluídas e bem arquitetadas antes da criação de qualquer componente visual final.
