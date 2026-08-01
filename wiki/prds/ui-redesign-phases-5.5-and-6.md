# PRD Tático: Fase 5.5 (Arquitetura & Hardening) e Fase 6 (Compositor FBO)

Após a consolidação das Fases 1 a 5, o ecossistema visual de C++ (SDF, Constraints, Tipografia e Estados) demonstrou altíssima qualidade técnica. Contudo, auditorias revelaram que a progressão direta para a Fase 6 original resultaria em uma escalabilidade comprometida (falta de primitivas geométricas) e vazamentos de abstração (Widgets tocando diretamente na GPU).

Este documento delineia a **Fase 5.5 (Hardening)** para selar a arquitetura antes da **Fase 6**, garantindo que o framework permaneça limpo, preditivo e altamente sofisticado visualmente (estética Luminous Glass verdadeira).

---

## 1. Fase 5.5: Endurecimento Arquitetural e Layout Geométrico

A filosofia da nossa Engine exige que *"Widgets nunca desenhem shaders por conta própria"* e *"Widgets não adivinham matemática X/Y sozinhos"*.

### 1.1. Inversão de Dependência (Blindagem do RenderContext)
* **Problema**: `Panel` e `Button` instanciavam e alimentavam o pipeline gráfico chamando o `ShaderPipeline` diretamente, acoplando-os à GPU e quebrando o encapsulamento.
* **Solução Técnica**: 
  O `RenderContext` centraliza todo o *draw call* de alto nível com a assinatura:
  `void DrawDecoratedBox(Rect bounds, const BoxDecoration& style)`
  O contexto decide se aplica o Shader SDF, se passa a cor sólida ou se ignora o desenho. Os widgets apenas entregam o "O que" (`BoxDecoration`), e o contexto decide o "Como".
* **Impacto**:
  - [render_context.cppm](file:///home/gk_igor_dev/development/opensource/jay/jay-frontend-cpp/src/engine/render_context.cppm)
  - [button.cppm](file:///home/gk_igor_dev/development/opensource/jay/jay-frontend-cpp/src/shared/widgets/button.cppm) (Remoção da lógica OpenGL)
  - [panel.cppm](file:///home/gk_igor_dev/development/opensource/jay/jay-frontend-cpp/src/shared/widgets/panel.cppm) (Remoção da lógica OpenGL)

### 1.2. Primitivas de Restrição de Espaço (Layout System)
* **Problema**: Sem elementos que governem X e Y automaticamente, criar uma tela complexa exige cálculos manuais propensos a quebras em redimensionamentos de janela.
* **Solução Técnica**: 
  Implementação de caixas puras de roteamento de limites matemáticos (`BoxConstraints`):
  * **`Padding`**: Encolhe as constraints recebidas subtraindo margens internas antes de repassar ao filho, e desloca a coordenada X,Y do filho.
  * **`Column` (Vertical Flexbox)**: Divide sequencialmente a altura máxima entre os filhos, baseando-se em seus tamanhos intrínsecos e atualizando deltas de posição no `SetBounds()`.
  * **`Row` (Horizontal Flexbox)**: Igual à Column, mas fatiando o limite da largura (`maxW`).
  * **`Stack` (Z-Index Absoluto)**: Renderiza filhos empilhados um sobre o outro (usado para colocar Textos sobre Paineis, ou Modais acima da aplicação).
* **Impacto**:
  - [layout_primitives.cppm](file:///home/gk_igor_dev/development/opensource/jay/jay-frontend-cpp/src/shared/layout/layout_primitives.cppm)

### 1.3. O Motor Analítico de Hit-Testing Global
* **Problema**: O método antigo usava `m_bounds.Contains(mouseX, mouseY)`. Se um Modal (Fundo Desfocado) for instanciado acima da tela principal, o clique atravessa o desfoque e aciona botões escondidos atrás.
* **Solução Técnica**:
  A raiz (`Shell` ou `InputDispatcher`) executa um algoritmo de *Hit-Testing* descendo a árvore de trás pra frente (`rbegin()` para o topo do eixo Z até o fundo). Componentes com `SetAbsorbsEvents(true)` agem como `MouseRegion`, engolindo a propagação de eventos e bloqueando cliques abaixo.

---

## 2. Fase 6: Shell & Glass Compositor (O Desfoque Puro)

Com a geometria trancada e a injeção gráfica centralizada na Fase 5.5, implementaremos a estética real do Glassmorphism.

### 2.1. O Compositor em RenderTexture2D (FBO)
* **Estratégia**: O Raylib não permite borrifar um efeito de desfoque *apenas* naquilo que já foi pintado abaixo de um botão de forma direta. O `Shell` cria um Framebuffer Object (`RenderTexture2D`) para renderizar a cena opaca e aplicar o shader de desfoque de fundo (Glass Blur).

### 2.2. A Injeção de Sampler Duplo (Blur sem Vazamento)
* **Solução para o Vazamento**:
  1. A camada inferior da interface é capturada pela FBO principal.
  2. Essa FBO inteira sofre o *Multi-pass Kawase Blur* em FBOs secundários gerando a textura `BlurredBackground`.
  3. Quando o `RenderContext->DrawDecoratedBox` for acionado para um botão ou painel Glass, ele invoca o `rounded_rect.fs` enviando a textura `BlurredBackground` como variável `uniform sampler2D`.
  4. O fragment shader de borda suave pega as cores já borradas dessa textura usando as coordenadas UV absolutas da tela (`gl_FragCoord.xy / screenSize.xy`), mas aplica o recorte do raio SDF para matar os cantos, gerando o vidro perfeito sem bordas quadradas.

* **Impacto**:
  - [shell.cppm](file:///home/gk_igor_dev/development/opensource/jay/jay-frontend-cpp/src/app/shell.cppm) (Gerenciador de FBO base)
  - [rounded_rect.fs](file:///home/gk_igor_dev/development/opensource/jay/jay-frontend-cpp/assets/shaders/rounded_rect.fs) (Inclusão de textura de base)
  - [shader_pipeline.cppm](file:///home/gk_igor_dev/development/opensource/jay/jay-frontend-cpp/src/engine/shader_pipeline.cppm) (Orquestração do Kawase e Texturas)
