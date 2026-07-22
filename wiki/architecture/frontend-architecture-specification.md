# Frontend Architecture Specification (Phase 4 Foundation)

**Projeto:** Jay Frontend C++20 GUI Engine — `jay-frontend-cpp`  
**Linguagem & Padrão:** C++20 (C++ Modules `.cppm`)  
**Biblioteca Gráfica:** Raylib (versão estritamente contida na camada Render)  
**Comunicação I/O:** Socket Unix Domain com protocolo IPC v1 em thread assíncrona  
**Status:** Arquitetura Base — Aguardando Revisão e Aprovação

---

## 1. Visão Arquitetural

O **Jay Frontend** é projetado não como um conjunto ad-hoc de telas, mas como um **Motor de Interface Gráfica Reativo, Modular e de Altíssima Performance (GUI Engine)**.

### Objetivos Arquiteturais Fundamentais:
1. **Desacoplamento Total de Framework Gráfico**: A biblioteca Raylib é uma dependência de infraestrutura estritamente confinada na camada de renderização (`Render`). Nenhum ViewModel, componente de estado, regra de negócio ou gerenciador de mensagens conhece ou importa `<raylib.h>`.
2. **Desempenho Zero-Allocation no Hot Loop (60 FPS Constantes)**: É proibida alocação dinâmica de memória (`new`, `std::make_shared`, `std::vector::push_back` que cause realocação) dentro dos métodos executados a cada frame (`Update` e `Render`). Toda a memória necessária para renderização, buffers de texto e estruturas de layout é pré-alocada ou reutilizada de arena/vetores pré-dimensionados.
3. **Fluxo Unidirecional de Dados (Unidirectional Data Flow)**:
   $$\text{State} \longrightarrow \text{ViewModel} \longrightarrow \text{Widget Tree} \longrightarrow \text{User Event} \longrightarrow \text{Command/Action} \longrightarrow \text{State}$$
4. **Isolamento Concorrente I/O vs. Loop Gráfico**: A comunicação via socket Unix com o daemon Go roda em uma thread dedicada. A troca de mensagens entre a thread IPC e a thread principal de renderização ocorre por meio de uma fila atômica lock-free / thread-safe, garantindo que o frame rate da Raylib nunca seja interrompido por atrasos de I/O ou bloqueios do SO.

---

## 2. Princípios de Projeto

- **RAII & Ownership Explícito**: Propriedade estrita baseada em `std::unique_ptr` para nós pai de widgets, views de não-propriedade (`non-owning pointers` / `std::string_view` / `std::span`) para passagem de dados de renderização, e `std::weak_ptr` para subscrições de callbacks.
- **Invalidação Reativa de Layout (Dirty Pattern)**: Geometrias e limites visuais não são recalculados a cada frame. Cálculos de layout só ocorrem quando uma marcação de "dirty" (`m_layoutDirty = true`) for disparada por alteração de estado ou redimensionamento de janela.
- **Separação Rígida entre Update e Render**:
  - `Update(deltaTime)`: Processa animações, interpolações físicas, estado dos botões e transições de foco. **Nenhuma instrução de desenho é executada aqui**.
  - `Render(RenderContext)`: Varre a árvore de widgets emitindo primitivas gráficas Raylib. **Nenhuma mutação de estado de negócio ocorre aqui**.

---

## 3. Diagrama Completo das Camadas

```
+-------------------------------------------------------------------------+
|                           APPLICATION LAYER                             |
|    (App Bootstrapper, Shell, Window Manager, Master Game Loop)          |
+-------------------------------------------------------------------------+
                                     |
                                     v
+-------------------------------------------------------------------------+
|                       PRESENTATION & WIDGET ENGINE                      |
|  (Widget Tree Node, Layout Engine, Focus Manager, Input Dispatcher)     |
+-------------------------------------------------------------------------+
             |                                             |
             v                                             v
+--------------------------+                 +----------------------------+
|    RAYLIB DRAW LAYER     |                 |      VIEW MODEL LAYER      |
| (Renderers, Shapes,      |                 |  (Reactive Bindings,       |
|  Fonts, Scissor, FBO)    |                 |   State Observer, Actions) |
+--------------------------+                 +----------------------------+
                                                           |
                                                           v
+-------------------------------------------------------------------------+
|                        APPLICATION STATE / DOMAIN                       |
|        (App State Store, Event Bus Local, Invalidation Notifier)        |
+-------------------------------------------------------------------------+
                                     |
                                     v
+-------------------------------------------------------------------------+
|                           IPC BOUNDARY                                  |
|   (Async IPC Thread, Lock-free SPSC RingBuffer, Message Envelopes v1)   |
+-------------------------------------------------------------------------+
```

---

## 4. Fluxo do Game Loop (Master Loop)

O loop principal executa a exatamente 60 FPS (ou sincronizado com a taxa de atualização do monitor), composto por 5 fases determinísticas em sequência:

```
[ Início do Frame ]
       │
       ▼
 1. Poll Inputs          --> Raylib captura eventos brutos de mouse/teclado e alimenta o InputDispatcher
       │
       ▼
 2. Drain IPC Queue      --> Drena mensagens da fila thread-safe vinda do Socket Unix em lotes (sem bloqueio)
       │
       ▼
 3. Update (dt)          --> Avança animações, simulação gráfica e re-avalia geometrias dirty (Layout Engine)
       │
       ▼
 4. Render (Context)     --> BeginDrawing() -> Renderiza Árvore de Widgets por profundidade -> EndDrawing()
       │
       ▼
[ Fim do Frame ]
```

---

## 5. Fluxo de Eventos

Existem dois fluxos distintos de eventos na arquitetura:

1. **Eventos de Input (Baixo Nível - Raylib)**:
   - Capturados na fase `Poll Inputs`.
   - Propagados da raiz até as folhas da Árvore de Widgets (**Tunneling**), ou da folha focada até a raiz (**Bubbling**). Se um widget consome o evento (`event.handled = true`), a propagação encerra imediatamente.
2. **Eventos de Domínio/Estado (Alto Nível - Dispatcher)**:
   - Disparados pela fila IPC ou por ações do usuário.
   - Trafegam via `EventBus` assíncrono interno e notificam Observers / ViewModels registrados.

---

## 6. Fluxo de Renderização

A renderização segue o padrão **Depth-First Traversal (Pré-ordem)** na Árvore de Widgets:

```
RenderContext (contém Fontes, Tempos, Scissor Rect Ativo, Theme Palette)
  │
  ├─> Window / Screen Root -> BeginScissorMode(...)
  │      ├─> Container Panel 1
  │      │      ├─> Widget A (Renderiza fundo e bordas)
  │      │      └─> Widget B (Renderiza texto/ícone)
  │      └─> Container Panel 2
  │             └─> Overlay / Modal (Renderizado no topo)
  └─> EndScissorMode()
```

- **Scissor Stack**: O `RenderContext` mantém uma pilha de recortes rect (Scissor Stack) para garantir que widgets filhos nunca desenhem fora dos limites de seus painéis pai.

---

## 7. Fluxo de Atualização de Estado

A mutação de estado nunca ocorre de forma direta dentro de um componente visual:

1. O Usuário interage com um `Widget` (clique ou digitação).
2. O `Widget` invoca um comando no seu `ViewModel` associado.
3. O `ViewModel` dispara uma `Action` para a `StateStore` ou envia uma mensagem RPC via `IPCClient`.
4. A `StateStore` atualiza o estado centralizado e emite notificação de alteração.
5. Os `ViewModels` ouvintes recebem a notificação e marcam os `Widgets` correspondentes como `Dirty`.
6. No próximo ciclo de `Update`, os `Widgets` dirty recalculam seu conteúdo e layout.

---

## 8. Arquitetura de Widgets

A infraestrutura GUI é construída sobre uma classe base abstrata e pura `Widget`:

```cpp
// Representação conceitual do contrato base de Widget em C++20
export module jay.engine.widget;

import jay.engine.types;
import jay.engine.render_context;
import jay.engine.events;

export namespace jay::engine {

class Widget {
public:
    virtual ~Widget() = default;

    virtual void Init() {}
    virtual void Update(float deltaTime) {}
    virtual void Layout(const BoxConstraints& constraints) {}
    virtual void Render(RenderContext& ctx) const = 0;
    virtual bool OnEvent(const InputEvent& event) { return false; }

    void MarkLayoutDirty() { m_layoutDirty = true; }
    bool IsLayoutDirty() const { return m_layoutDirty; }

    void SetBounds(const Rect& bounds) { m_bounds = bounds; }
    Rect GetBounds() const { return m_bounds; }

protected:
    Rect m_bounds{0.0f, 0.0f, 0.0f, 0.0f};
    bool m_layoutDirty{true};
    bool m_visible{true};
    bool m_enabled{true};
};

} // namespace jay::engine
```

- **Composição em Árvore**: Container Widgets mantêm contêineres de propriedade `std::vector<std::unique_ptr<Widget>> m_children`.

---

## 9. Arquitetura de Layout

O cálculo de layout adota o modelo **Box Constraints** (semelhante a motores modernos como Flutter/Web Flexbox):

1. **Passo 1 (Top-Down Constraints)**: O pai passa restrições de tamanho mínimo e máximo (`BoxConstraints`) para o filho através de `Widget::Layout(constraints)`.
2. **Passo 2 (Bottom-Up Size)**: O filho determina seu tamanho preferido e o ajusta dentro das restrições recebidas.
3. **Passo 3 (Parent Position Assignment)**: O pai posiciona o filho calculando suas coordenadas relatórias e atribui o `Rect` final via `SetBounds()`.

### Algoritmo de Invalidação (Dirty Flag):
- Cada nó da árvore possui a flag `m_layoutDirty`.
- Se um widget altera seu texto ou tamanho, ele chama `MarkLayoutDirty()`, que propaga a invalidação para o seu nó pai.
- No ciclo de `Update`, apenas as subárvores marcadas como `m_layoutDirty == true` executam o algoritmo de relayout.

---

## 10. Arquitetura de Input & Gerenciamento de Foco

O `FocusManager` gerencia o foco do teclado de forma determinística:

- **Nó Focado Ativo**: Apenas um widget por janela pode possuir o foco de entrada por vez (`Widget* m_focusedWidget`).
- **Navegação por Tabulação**: O `FocusManager` mantém uma lista ordenada por índice de foco (`tabIndex`) e responde às teclas `Tab` e `Shift+Tab`.
- **Hit-Testing por Mouse**: Em cliques de ponteiro, o `InputDispatcher` realiza o teste de colisão (*Point-in-Rect Hit-Testing*) de cima para baixo na árvore visual para atribuir o foco ao widget clicado.

---

## 11. Arquitetura de Animação

O `AnimationEngine` é responsável por transições visuais fluidas:

- **Interpoladores (Tweens)**: Animações de posição, opacidade, cor e escala utilizam o tipo `Tween<T>` com funções de suavização configuráveis (`EaseLinear`, `EaseInOutCubic`, `EaseOutBounce`, `Spring`).
- **Tick Sem Frame-Drop**: O tempo da animação é incrementado no ciclo de `Update(deltaTime)` baseado no tempo delta acumulado, garantindo que animações não acelerem nem desacelerem se a taxa de quadros oscilar.

---

## 12. Arquitetura de IPC

A comunicação com o daemon Go utiliza o **Protocolo IPC v1 (RequestEnvelope / ResponseEnvelope)**:

- **Socket Client Thread**: Thread secundária exclusiva que gerencia o ciclo de vida da conexão Unix Socket (`net.Dial("unix", socketPath)`).
- **Envelopes Tipados**:
  - Solicitação: `RequestEnvelope { protocol_version: 1, request_id, client_id, type, payload }`
  - Resposta: `ResponseEnvelope { protocol_version: 1, request_id, type, status, error, payload }`
- **Matching de Respostas RPC**: O `IPCClient` mantém um mapa thread-safe de callbacks pendentes associados ao `request_id`. Quando uma resposta chega, a mensagem é pareada e despachada para a fila da thread principal.

---

## 13. Modelo de Threading & Thread Safety

```
+-------------------------------------------------------+
|                     MAIN THREAD                       |
|   (Raylib Graphics, Event Loop, Widgets, ViewModels)  |
+-------------------------------------------------------+
       │                                     ▲
       │ Envia Chamada RPC                   │ Processa Eventos /
       │ (Lock-free Push)                    │ Respostas IPC
       ▼                                     │
+-------------------------------------------------------+
|            SPSC RING BUFFER / QUEUE                   |
+-------------------------------------------------------+
       │                                     ▲
       ▼ Recebe Requisições                  │ Grava Respostas /
+-------------------------------------------------------+
|                     IPC THREAD                        |
|    (Unix Socket Reader/Writer, JSON Serialization)    |
+-------------------------------------------------------+
```

- **Garantia de Thread Safety**: **Nenhum método da Raylib é chamado fora da Main Thread**. A comunicação entre a `IPC Thread` e a `Main Thread` é estritamente unidirecional através de filas atômicas com trava zero ou `std::mutex` com duração de bloqueio de nanossegundos (`std::lock_guard` rápido apenas para mover elementos do vetor).

---

## 14. Modelo de Ownership e Lifetime (RAII)

- **Árvore de Widgets**: Os nós pai possuem a propriedade exclusiva dos nós filhos através de `std::unique_ptr<Widget>`. Quando o nó pai é destruído, toda a sua subárvore visual é desalocada automaticamente via destrutores RAII nativos.
- **Callbacks & Subscrições**: Observadores usam `std::weak_ptr` para evitar vazamentos de memória ou referências penduradas (*dangling references*) se um widget for destruído enquanto um evento assíncrono estiver a caminho.

---

## 15. Convenções para Módulos C++20

Todos os arquivos do projeto utilizam obrigatoriamente a sintaxe nativa de **C++20 Modules (`.cppm`)**:

- **Sintaxe de Declaração**:
  ```cpp
  module;
  // Inclusões de cabeçalhos C/C++ tradicionais de terceiros (ex: <vector>, <memory>) entram no Global Module Fragment
  #include <memory>
  #include <vector>

  export module jay.engine.widget;

  // Importação de módulos internos do projeto
  import jay.engine.types;

  export namespace jay::engine {
      // Declaração e exportação de tipos e classes
  }
  ```
- **Nomenclatura de Arquivo**: Nome do arquivo sempre em `snake_case` com extensão `.cppm` (ex: `render_context.cppm`, `text_layout.cppm`).
- **Nomenclatura de Módulo**: Namespace pontilhado iniciado por `jay.` (ex: `export module jay.engine.layout;`).

---

## 16. Convenções de Nomenclatura

| Elemento | Convenção | Exemplo |
|---|---|---|
| Módulos C++20 | `jay.subsistema.componente` | `jay.engine.widget` |
| Arquivos `.cppm` | `snake_case.cppm` | `render_context.cppm` |
| Namespaces | `jay::subsistema` | `jay::engine` |
| Classes / Structs | `PascalCase` | `RenderContext`, `Widget` |
| Métodos / Funções | `PascalCase` | `MarkLayoutDirty()`, `Render()` |
| Variáveis Membro | `m_camelCase` | `m_layoutDirty`, `m_bounds` |
| Variáveis Locais | `camelCase` | `deltaTime`, `targetWidth` |
| Constantes / Enums | `PascalCase` ou `ALL_CAPS` | `ProtocolVersionCurrent` |

---

## 17. Estrutura Completa de Diretórios

```
jay-frontend-cpp/
├── CMakeLists.txt
├── src/
│   ├── main.cpp                              [Ponto de Entrada Nativo]
│   ├── app/
│   │   ├── application.cppm                  [Bootstrapper Master Loop & Window Config]
│   │   └── shell.cppm                        [Window Manager & Layout Raiz]
│   ├── engine/
│   │   ├── types.cppm                        [Structs de Geometria: Rect, Vec2, Color]
│   │   ├── render_context.cppm               [Abstração de Desenho e Scissor Stack]
│   │   ├── widget.cppm                       [Classe Base Abstrata Widget]
│   │   ├── layout_engine.cppm                [Cálculo de Restrições BoxConstraints & Dirty Flags]
│   │   ├── focus_manager.cppm                [Gerenciador de Foco e Tabulação]
│   │   ├── input_dispatcher.cppm             [Distribuidor de Eventos de Teclado/Mouse]
│   │   └── animation_engine.cppm             [Interpoladores Tweens & Easing Functions]
│   ├── shared/
│   │   ├── widgets/
│   │   │   ├── panel.cppm                    [Container Genérico de Layout]
│   │   │   ├── label.cppm                    [Widget de Exibição de Texto com Layout Engine]
│   │   │   ├── button.cppm                   [Widget Interativo de Botão]
│   │   │   ├── text_input.cppm               [Widget de Caixa de Entrada Multilinha]
│   │   │   └── scroll_container.cppm         [Container com Scrollbar Integrada]
│   │   └── theme/
│   │       └── theme_palette.cppm            [Tokens de Cores, Fontes e Espaçamentos]
│   └── ipc/
│       ├── protocol.cppm                     [Tipos e Envelopes do Protocolo v1]
│       ├── ipc_client.cppm                   [Thread de Socket Unix & Request Pool]
│       └── event_dispatcher.cppm             [Dispatcher de Eventos Push]
└── tests/
    ├── unit/
    │   ├── test_layout_engine.cpp            [Testes de Invalidação e Geometria]
    │   ├── test_ipc_protocol.cpp             [Testes de Serialização de Envelopes]
    │   └── test_focus_manager.cpp            [Testes de Navegação por Foco]
```

---

## 18. Responsabilidades de Cada Módulo

- **`jay.app.application`**: Inicializa a janela da Raylib, mantém o Master Game Loop e gerencia o tempo delta.
- **`jay.engine.widget`**: Define a interface base e o ciclo de vida dos componentes visuais.
- **`jay.engine.layout_engine`**: Executa a resolução de constraints e invalidação dirty de geometrias.
- **`jay.engine.render_context`**: Empacota primitivas de renderização e mantém o estado da pilha de recorte scissor.
- **`jay.engine.focus_manager`**: Controla o estado global do componente em foco.
- **`jay.engine.input_dispatcher`**: Converte inputs brutos do SO em estruturas de eventos tipadas.
- **`jay.engine.animation_engine`**: Interpola propriedades numéricas ao longo do tempo.
- **`jay.ipc.ipc_client`**: Executa I/O assíncrona com o socket Unix do daemon em background.

---

## 19. Regras de Dependência Entre Módulos

```
                                 ┌─────────────────────────┐
                                 │  jay.app.application    │
                                 └────────────┬────────────┘
                                              │
                                              ▼
                                 ┌─────────────────────────┐
                                 │    jay.app.shell        │
                                 └────────────┬────────────┘
                                              │
                                              ▼
                                 ┌─────────────────────────┐
                                 │  jay.shared.widgets     │
                                 └────────────┬────────────┘
                                              │
                                              ▼
┌─────────────────────────┐      ┌─────────────────────────┐      ┌─────────────────────────┐
│   jay.engine.widget     │◄─────┤ jay.engine.layout_engine│◄─────┤   jay.ipc.ipc_client    │
└────────────┬────────────┘      └─────────────────────────┘      └────────────┬────────────┘
             │                                                                 │
             ▼                                                                 ▼
┌─────────────────────────┐                                       ┌─────────────────────────┐
│ jay.engine.render_ctx   │                                       │   jay.ipc.protocol      │
└────────────┬────────────┘                                       └─────────────────────────┘
             │ (ÚNICO LOCAL PERMITIDO)
             ▼
      <raylib.h> / Raylib C-API
```

### Regras Estritas de Importação:
1. **Nenhum módulo da camada `jay.engine` (exceto `render_context`) ou `jay.ipc` ou ViewModels pode importar `<raylib.h>`**.
2. **Dependências Circulares São Proibidas**: A dependência flui sempre de cima para baixo: `App` -> `Widgets` -> `Engine` -> `Types`.
3. **Módulos `jay.ipc` não possuem dependências com módulos visuais/widgets**.

---

## 20. Estratégia de Testes Headless (Sem Janela Gráfica)

Toda a arquitetura do motor GUI é projetada para ser testada sem depender do servidor gráfico X11/Wayland ou de janelas abertas:

1. **Testes do Layout Engine**: Instancia nós de `Widget`, atribui `BoxConstraints` e valida os retângulos `Rect` resultantes através de asserções C++.
2. **Testes do Focus Manager**: Injeta eventos de teclado sintéticos e valida a transferência do ponteiro de foco entre nós.
3. **Testes de Serialização IPC**: Valida a construção de envelopes `RequestEnvelope` (v1) e a desserialização de `ResponseEnvelope` usando `nlohmann::json` sem abrir sockets.
4. **Testes de Invalidação (Dirty Flag)**: Modifica o texto de um label e verifica se a flag `m_layoutDirty` foi propagada até a raiz.
