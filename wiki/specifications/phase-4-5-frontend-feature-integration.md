# Especificação Técnica Refinada — Fase 4.5: C++ Frontend Feature Integration & Decoupling

**Projeto:** Jay Frontend C++20 — `jay-frontend-cpp`  
**Status:** Aprovado pós-Review (Nota 9,8/10)

---

## Estrutura Sequencial das Tasks da Fase 4.5

### Módulo A — Avatar Feature (`src/features/avatar`)
- [ ] **Task 27: Avatar ViewModel & State Transition Widget** (`src/features/avatar/avatar_viewmodel.cppm`, `avatar_widget.cppm`)

### Módulo B — Chat Feature (`src/features/chat`)
- [ ] **Task 28: Message Bubble Component Widget** (`src/features/chat/message_bubble_widget.cppm`)
- [ ] **Task 29: Chat Feed Widget & ViewModel Composition** (`src/features/chat/chat_viewmodel.cppm`, `chat_widget.cppm`)
- [ ] **Task 30: Chat Input & Editor Widget Integration** (`src/features/chat/chat_input_widget.cppm`)

### Módulo C — Permissions Feature (`src/features/permissions`)
- [ ] **Task 31: Permissions Modal Overlay Widget** (`src/features/permissions/permissions_widget.cppm`)

### Módulo D — Composição, Validação e Limpeza (`src/app/`)
- [ ] **Task 32: Shell Component Integration** (`src/app/shell.cppm`)
- [ ] **Task 33: Architecture Parity Validation** (Validação intermediária com código antigo preservado)
- [ ] **Task 34: Legacy Code Cleanup & Deletion** (Deleção de `src/app/renderer.cppm` e legados)
- [ ] **Task 35: Full Stack E2E Verification** (Validação final E2E com daemon Go `jayd`)
