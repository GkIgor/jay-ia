# Frontend Architecture Specification (Phase 4 Foundation)

**Projeto:** Jay Frontend C++20 GUI Engine — `jay-frontend-cpp`  
**Linguagem & Padrão:** C++20 (C++ Modules `.cppm`)  
**Biblioteca Gráfica:** Raylib (versão estritamente contida na camada Render)  
**Comunicação I/O:** Socket Unix Domain com protocolo IPC v1 em thread assíncrona  
**Status:** Arquitetura Base — Aprovada (Review 3 — Nota 10/10)

---

## 1. Visão Arquitetural & Fluxo Unidirecional

$$\text{StateStore} \longrightarrow \text{ViewModel} \longrightarrow \text{Widget Tree} \longrightarrow \text{User Event} \longrightarrow \text{Use Case} \longrightarrow \text{Action} \longrightarrow \text{StateStore}$$

---

## 2. Camada de Use Cases (Workflows de Casos de Uso)

```
+------------------+         +--------------------------+         +-----------------+
|     Widget       |  ────>  |        ViewModel         |  ────>  |    Use Case     |
| (Componente UI)  |         | (Estado UI & Formatação) |         | (Caso de Uso)   |
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

## 3. Política de Atualização Otimista com Rollback e Tradução de Falhas

1. **Update Otimista**: O `UseCase` despacha uma `Action` para a `StateStore` alterando o estado para pendente (`MessagePending`).
2. **Disparo Assíncrono**: O `UseCase` envia a RPC pelo `IPCClient`.
3. **Resolução de Estado**:
   - **Sucesso**: O `UseCase` despacha confirmação (`MessageSent`).
   - **Falha/Timeout**: O `UseCase` captura a falha de rede/socket, traduz em um erro de domínio e despacha estado de falha (`MessageFailed`), permitindo retry do usuário.

---

## 4. Regras Rígidas de Direção de Dependência

$$\text{CORRETO: } \text{Widget} \longrightarrow \text{ViewModel} \longrightarrow \text{UseCase} \longrightarrow \text{IPCClient} \text{ / } \text{StateStore}$$

$$\text{PROIBIDO: } \text{Widget} \xlongrightarrow{\text{NÃO!}} \text{IPCClient} \quad \text{ou} \quad \text{Widget} \xlongrightarrow{\text{NÃO!}} \text{StateStore} \quad \text{ou} \quad \text{StateStore} \xlongrightarrow{\text{NÃO!}} \text{IPCClient}$$
