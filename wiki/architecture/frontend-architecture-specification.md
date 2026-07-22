# Frontend Architecture Specification (Phase 4 Foundation)

**Projeto:** Jay Frontend C++20 GUI Engine — `jay-frontend-cpp`  
**Linguagem & Padrão:** C++20 (C++ Modules `.cppm`)  
**Biblioteca Gráfica:** Raylib (versão estritamente contida na camada Render)  
**Comunicação I/O:** Socket Unix Domain com protocolo IPC v1 em thread assíncrona  
**Status:** Arquitetura Base — Revisada & Aprimorada (Review 2)

---

## 1. Visão Arquitetural

O **Jay Frontend** é projetado não como um conjunto ad-hoc de telas, mas como um **Motor de Interface Gráfica Reativo, Modular e de Altíssima Performance (GUI Engine)**.

---

## 2. Ownership dos Componentes Arquiteturais

```
+-----------------------------------------------------------------------------------+
| APPLICATION (Owner de Topo)                                                       |
|   ├── StateStore          (std::unique_ptr<StateStore>)                           |
|   ├── IPCClient           (std::unique_ptr<IPCClient>)                            |
|   ├── ApplicationServices (std::unique_ptr<AppServiceGroup>)                      |
|   ├── RenderContext       (std::unique_ptr<RenderContext>)                        |
|   └── Shell               (std::unique_ptr<Shell>)                                |
|          ├── ViewModelRegistry (std::unique_ptr<ViewModel> por tela/container)    |
|          └── WidgetTree Root   (std::unique_ptr<ContainerWidget>)                 |
|                 └── Sub-Widgets (std::unique_ptr<Widget> pertencente ao pai)     |
+-----------------------------------------------------------------------------------+
```

---

## 3. Camada de Application Services (Casos de Uso)

```
+------------------+         +--------------------------+         +-----------------+
|     Widget       |  ────>  |        ViewModel         |  ────>  |   App Service   |
| (Componente UI)  |         | (Estado UI & Formatação) |         |  (Caso de Uso)  |
+------------------+         +--------------------------+         +--------┬--------+
                                                                           │
                                                    ┌──────────────────────┴──────────────────────┐
                                                    ▼                                             ▼
                                          +-------------------+                         +-------------------+
                                          |    IPCClient      |                         |    StateStore     |
                                          | (Envio de RPCs)   |                         | (Mutação de Dados)|
                                          +-------------------+                         +-------------------+
```

---

## 4. Regras Rígidas de Direção de Dependência

$$\text{CORRETO: } \text{Widget} \longrightarrow \text{ViewModel} \longrightarrow \text{AppService} \longrightarrow \text{IPCClient} \text{ / } \text{StateStore}$$

$$\text{PROIBIDO: } \text{Widget} \xlongrightarrow{\text{NÃO!}} \text{IPCClient} \quad \text{ou} \quad \text{Widget} \xlongrightarrow{\text{NÃO!}} \text{StateStore} \quad \text{ou} \quad \text{StateStore} \xlongrightarrow{\text{NÃO!}} \text{IPCClient}$$
