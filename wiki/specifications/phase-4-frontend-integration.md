# Especificação Técnica — Fase 4: C++ Frontend IPC Integration & Modern GUI

**Projeto:** Jay Frontend C++20 — `jay-frontend-cpp`  
**Tecnologias:** C++20 (Modules `.cppm`), Raylib, `nlohmann::json`, Socket Unix Domain  
**Status:** Aguardando aprovação

---

## Estrutura de Tasks da Fase 4

### Módulo 1: Infraestrutura de Protocolo v1 & Async IPC (`src/ipc`)
- **Task 18: IPC Protocol v1 Serializer & Envelope Engine** (`src/ipc/protocol.cppm`, `messages.cppm`)
- **Task 19: Async RPC Client & Pending Requests Pool** (`src/ipc/ipc_client.cppm`)
- **Task 20: Typed Push Event Dispatcher** (`src/ipc/event_dispatcher.cppm`)

### Módulo 2: Integração da Interface de Chat & Histórico (`src/features/chat`)
- **Task 21: Registration & Chat Lifecycle Integration** (`src/app/app.cppm`)
- **Task 22: Message Feed Sync & Real-time Processing** (`src/features/chat/chat_renderer.cppm`, `chat_input.cppm`)

### Módulo 3: Modal de Permissões & Reatividade do Avatar (`src/features/permissions` & `src/features/avatar`)
- **Task 23: Interactive Permissions Modal** (`src/features/permissions/permissions_modal.cppm`)
- **Task 24: Avatar Motion State Machine Integration** (`src/features/avatar/avatar_renderer.cppm`)

### Módulo 4: Validação E2E & Polimento Visual
- **Task 25: Full Stack E2E Validation & Polishing**
