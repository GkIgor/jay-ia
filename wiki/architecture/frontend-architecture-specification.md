# Frontend Architecture Specification (Phase 4 Foundation)

**Projeto:** Jay Frontend C++20 GUI Engine — `jay-frontend-cpp`  
**Linguagem & Padrão:** C++20 (C++ Modules `.cppm`)  
**Biblioteca Gráfica:** Raylib (versão estritamente contida na camada Render)  
**Comunicação I/O:** Socket Unix Domain com protocolo IPC v1 em thread assíncrona  
**Status:** Arquitetura Base — Revisada & Aprimorada

---

## 1. Visão Arquitetural

O **Jay Frontend** é projetado não como um conjunto ad-hoc de telas, mas como um **Motor de Interface Gráfica Reativo, Modular e de Altíssima Performance (GUI Engine)**.

### Objetivos Arquiteturais Fundamentais:
1. **Desacoplamento Total de Framework Gráfico**: A biblioteca Raylib é uma dependência de infraestrutura estritamente confinada na camada de renderização (`RenderContext`). Nenhum ViewModel, componente de estado, regra de negócio ou gerenciador de mensagens conhece ou importa `<raylib.h>`.
2. **Desempenho Zero-Allocation no Hot Loop (60 FPS Constantes)**: É estritamente proibida alocação dinâmica de memória (`new`, `delete`, `std::make_shared`, `std::vector::push_back` com re-alocação de heap) dentro do caminho crítico de código executado a cada frame (`Update` e `Render`). Toda a memória necessária para renderização, buffers de texto e estruturas de layout é pré-alocada ou reutilizada de vetores pré-dimensionados.
3. **Fluxo Unidirecional de Dados (Unidirectional Data Flow)**:
   $$\text{StateStore} \longrightarrow \text{ViewModel} \longrightarrow \text{Widget Tree} \longrightarrow \text{User Event} \longrightarrow \text{Command/Action} \longrightarrow \text{StateStore}$$
4. **Isolamento Concorrente I/O vs. Loop Gráfico**: A comunicação via socket Unix com o daemon Go roda em uma thread dedicada. A troca de mensagens entre a thread IPC e a thread principal de renderização ocorre por meio de uma fila atômica thread-safe/lock-free, garantindo que o frame rate da Raylib nunca seja interrompido por atrasos de I/O ou bloqueios do SO.

---

## 2. Princípios de Projeto

- **RAII & Ownership Explícito**: Propriedade estrita baseada em `std::unique_ptr` para nós pai de widgets, views de não-propriedade (`non-owning pointers` / `std::string_view` / `std::span`) para passagem de dados de renderização, e `std::weak_ptr` para subscrições de callbacks.
- **Invalidação Reativa de Layout (Dirty Pattern)**: Geometrias e limites visuais não são recalculados a cada frame. Cálculos de layout só ocorrem quando uma marcação de "dirty" (`m_layoutDirty = true`) for disparada por alteração de estado ou redimensionamento de janela.
- **Separação Rígida entre Update e Render**:
  - `Update(deltaTime)`: Processa animações, interpolações físicas, estado dos botões e transições de foco. **Nenhuma instrução de desenho é executada aqui**.
  - `Render(RenderContext)`: Varre a árvore de widgets emitindo primitivas gráficas Raylib. **Efeito colateral zero: O Render nunca modifica o State, ViewModel ou estado interno do Widget**.

---

## 3. Ownership dos Componentes Arquiteturais

Para evitar código espaguete e acoplamentos indevidos (ex: um Widget acessando diretamente o `IPCClient` ou a `StateStore`), os papéis de propriedade e lifetime são definidos estritamente:

```
+-----------------------------------------------------------------------------------+
| APPLICATION (Owner de Topo)                                                       |
|   ├── StateStore      (std::unique_ptr<StateStore>)                               |
|   ├── IPCClient       (std::unique_ptr<IPCClient>)                                |
|   ├── RenderContext   (std::unique_ptr<RenderContext>)                            |
|   └── Shell           (std::unique_ptr<Shell>)                                    |
|          ├── ViewModelRegistry (Mapa de ViewModels por tela/escopo)              |
|          └── WidgetTree Root   (std::unique_ptr<ContainerWidget>)                 |
|                 └── Sub-Widgets (std::unique_ptr<Widget> pertencente ao pai)     |
+-----------------------------------------------------------------------------------+
```

### Tabela de Ownership e Lifetime:

| Componente | Proprietário (Owner) | Tipo de Ponteiro de Propriedade | Referência Usada por Clientes |
|---|---|---|---|
| `StateStore` | `Application` | `std::unique_ptr<StateStore>` | Non-owning reference (`StateStore&`) no `ViewModel` |
| `IPCClient` | `Application` | `std::unique_ptr<IPCClient>` | Non-owning reference no `StateStore` / Handlers de Ação |
| `RenderContext` | `Application` | `std::unique_ptr<RenderContext>` | Emprestado via `RenderContext&` no `Widget::Render()` |
| `Shell` | `Application` | `std::unique_ptr<Shell>` | Non-owning reference no `Application` |
| `ViewModel` | `Shell / Registry` | `std::shared_ptr<ViewModel>` | `std::weak_ptr<ViewModel>` ou ref não-proprietária em `Widget` |
| `Widget Tree Root`| `Shell` | `std::unique_ptr<Widget>` | Non-owning reference (`Widget*`) no `FocusManager` |
| `Child Widget` | `Parent Widget` | `std::unique_ptr<Widget>` | Non-owning reference (`Widget*`) na subárvore visual |

---

## 4. Arquitetura do State Engine (StateStore)

A `StateStore` é o coração reativo e fonte única de verdade da aplicação, dividida estritamente em **Domain State** e **Transient UI State**:

### 4.1. Domain State vs. Transient UI State
1. **Domain State (Centralizado na `StateStore`)**:
   - Dados persistentes e de negócio recebidos do IPC Core (ex: lista de conversas, mensagens de chat, histórico, status da conexão, catálogo de ferramentas).
   - Vive na `StateStore` principal.
2. **Transient UI State (Local no Widget / ViewModel)**:
   - Dados de efemeridade visual puramente gráfica (ex: se um botão está sob hover, posição do cursor de digitação, progresso de scroll da barra, estado de piscamento do cursor).
   - Vive encapsulado dentro do `Widget` ou seu `ViewModel`, sem poluir a `StateStore` de domínio.

### 4.2. Fluxo de Atualização Reativa e Snapshots Parciais
- A `StateStore` emite eventos granulares de notificação de alteração (ex: `OnChatMessagesUpdated`).
- Os `ViewModels` escutam seletores parciais do estado e, ao detectar alterações relevantes, marcam a flag `m_layoutDirty = true` dos widgets associados.

---

## 5. Regras Rígidas de Direção de Dependência

Fica proibida qualquer violação de camadas ou "atalhos" diretos:

$$\text{CORRETO: } \text{Widget} \longrightarrow \text{ViewModel} \longrightarrow \text{StateStore} \longrightarrow \text{IPCClient}$$

$$\text{PROIBIDO: } \text{Widget} \xlongrightarrow{\text{NÃO!}} \text{IPCClient} \quad \text{ou} \quad \text{Widget} \xlongrightarrow{\text{NÃO!}} \text{StateStore}$$

---

## 6. Definição Rígida de "Zero Allocation" no Hot Loop (60 FPS)

O **Hot Loop** é definido como o caminho de execução repetido 60 vezes por segundo dentro das fases `Update(dt)` e `Render(ctx)`:

### Operações Estritamente PERMITIDAS no Hot Loop:
- Alteração de memória de buffers previamente reservados (`vector.clear()` sem alterar `capacity()`).
- Reutilização de buffers de caracteres UTF-8 fixos via `std::string::assign()`.
- Reutilização de instâncias em Object Pools ou Arenas de alocação estática.
- Passagem de dados por valor escalar, referências const ou views (`std::string_view`, `std::span`).

### Operações Estritamente PROIBIDAS no Hot Loop:
- Invocação de `new`, `delete`, `malloc()`, `free()`.
- Instanciação dinâmica de smart pointers (`std::make_shared`, `std::make_unique`).
- Crescimento dinâmico de `std::vector` ou `std::unordered_map` que provoque realocação no Heap.
- Concatenação ou instanciação de `std::string` temporárias durante o desenho.
- Alocação de lambdas dinâmicas com capturas que provoquem alocação no Heap.
