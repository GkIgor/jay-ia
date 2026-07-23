# Especificação Técnica — Frontend Functional Inventory & Feature Specification

**Projeto:** Jay Frontend C++20 — `jay-frontend-cpp`  
**Status:** Inventário Funcional Aprovado (Nota 10/10)

---

## 1. Inventário Funcional & Matriz de Rastreabilidade

### 1.1. Shell & Navegação Global
- **SHELL-001** [CORE]: TabBar Superior (Navegação "AVATAR" e "CHAT" por clique ou `Ctrl+Tab`)
- **SHELL-002** [UX]: Resizing Responsivo (Ajuste de proporção 35% Avatar / 65% Chat via `BoxConstraints`)

### 1.2. Chat Feature — Message Feed
- **CHAT-001** [CORE]: Navegação por Scroll (Mouse wheel, `PageUp`/`PageDown` e drag de scrollbar)
- **CHAT-002** [UX]: Auto-Scroll Inteligente (Rola para base se enviado pelo usuário; preserva histórico se lendo mensagens antigas)
- **CHAT-003** [CORE]: Balões Responsivos (Desenho isolado sem deformar cantos arredondados)
- **CHAT-004** [CORE]: Diferenciação Semântica por cores (Usuário, Jay, Erro)
- **CHAT-005** [UX]: Reticências de Status (Piscamento animado para estado `Thinking`)
- **CHAT-006** [NICE]: Trim de Mensagem (Remoção automática de whitespace/newlines das pontas)
- **CHAT-007** [UX]: Seleção de Texto em Balões (Seleção por clique e arraste com destaque por linha)
- **CHAT-008** [UX]: Copiar Mensagem Rápido (Botão visual com confirmação "check" por 2s e cópia UTF-8)
- **CHAT-009** [CORE]: Ferramentas Retráteis (Tool Groups expansíveis com contadores e status por ação)

### 1.3. Chat Feature — Campo de Entrada (Chat Input & Text Editor)
- **CHAT-010** [CORE]: Edição UTF-8 & Desfazer/Refazer (`Ctrl+Z`, `Ctrl+Y`, `Ctrl+Shift+Z`)
- **CHAT-011** [UX]: Preservação de Coluna Preferida (`preferredColumn` com setas ↑ ↓)
- **CHAT-012** [CORE]: Seleção Multilinha no Input (`Ctrl+A`, mouse drag, duplo/triplo clique, `Shift+Clique`)
- **CHAT-013** [CORE]: Atalhos e Clipboard (`Ctrl+X`, `Ctrl+C`, `Ctrl+V`, deleção por palavra `Ctrl+Backspace/Delete`)
- **CHAT-014** [CORE]: Quebra de Linha Automática (Word Wrap e divisão de tokens gigantes)
- **CHAT-015** [UX]: Scroll Interno no Input (Scrollbar vertical automática e ajuste de clique sob scroll)
- **CHAT-016** [CORE]: Bloqueio do Botão Enviar (Desabilitação do botão e da tecla `Enter` durante processamento)

### 1.4. Avatar Feature
- **AVA-001** [CORE]: Máquina de Estados de Cores (Idle, Thinking, Executing, Sleeping)
- **AVA-002** [UX]: Animação de Pulso & Halo (Pulso suave contínuo e halo translúcido)

### 1.5. Permissions Feature
- **PERM-001** [CORE]: Overlay Glassmorphism (Modal centralizado bloqueando o fundo)
- **PERM-002** [CORE]: Resposta Rápida por Teclado/Clique (Botões e teclas `Y` / `N`)

---

## 2. Tasks de Portabilidade (Tasks 36 a 41)

- [ ] **Task 36: Rich Text Editor & Keyboard Shortcuts** (`CHAT-010`, `CHAT-011`, `CHAT-012`, `CHAT-013`, `CHAT-014`, `CHAT-015`)
- [ ] **Task 37: Rich Message Bubble & Copy Icon** (`CHAT-003`, `CHAT-004`, `CHAT-005`, `CHAT-006`, `CHAT-007`, `CHAT-008`)
- [ ] **Task 38: Collapsible Tool Group Widget** (`CHAT-009`)
- [ ] **Task 39: Intelligent Auto-Scroll & Scrollbar Drag** (`CHAT-001`, `CHAT-002`)
- [ ] **Task 40: Keyboard Shortcuts & Glassmorphism Overlay** (`PERM-001`, `PERM-002`)
- [ ] **Task 41: Full Feature Homologation Matrix**
