# Especificação Técnica — Fase 4.5: C++ Frontend Feature Integration & Decoupling

**Projeto:** Jay Frontend C++20 — `jay-frontend-cpp`  
**Status:** Aguardando aprovação  

---

## Estrutura de Tasks da Fase 4.5

### Módulo A — Integração do Avatar e Máquina de Estados Visual (`src/features/avatar`)
- [ ] **Task 27: Avatar Widget & State Transition VM** (`src/features/avatar/avatar_viewmodel.cppm`, `avatar_widget.cppm`)

### Módulo B — Integração do Chat, Mensagens e Feed Dinâmico (`src/features/chat`)
- [ ] **Task 28: Chat View & Feed Components Migration** (`src/features/chat/chat_viewmodel.cppm`, `chat_widget.cppm`, `message_bubble_widget.cppm`)
- [ ] **Task 29: Chat Input Widget & Text Editor Integration** (`src/features/chat/chat_input_widget.cppm`)

### Módulo C — Modal de Permissões e Decoupling Final (`src/features/permissions` & `src/app`)
- [ ] **Task 30: Permissions Widget & Modal Glassmorphism Integration** (`src/features/permissions/permissions_widget.cppm`)
- [ ] **Task 31: Shell Integration & Legacy Cleanup** (`src/app/shell.cppm`, remoção de `renderer.cppm` e código legado obsoleto)
