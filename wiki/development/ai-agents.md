# Processo de Revisão por Agentes de IA

Para garantir o rigor técnico, a estética premium e a fidelidade arquitetural do Projeto Jay, foram instituídas três personas de agentes de auditoria. 

Estes perfis de agentes estão definidos no diretório `agents/` na raiz da workspace:
- **UX Agent** (`agents/ux-agent.md`): Audita a qualidade visual, usabilidade (UX) e transições visuais no frontend Raylib.
- **Code-Review Agent** (`agents/code-review-agent.md`): Executa análise estática de segurança, vazamentos de memória e prevenção de bugs lógicos.
- **Architect Agent** (`agents/architect-agent.md`): Verifica a conformidade com Clean Code, SOLID e alinhamento do código com a Wiki.

---

## Como Utilizar as Personas de Auditoria

Sempre que uma implementação for **longa, complexa ou baseada em um plano (implementation plan)**, você como agente IA ativo deve realizar o seguinte fluxo de auditoria antes de declarar a tarefa concluída:

1. **Auto-Revisão de Arquitetura**:
   - Abra e leia o arquivo `agents/architect-agent.md` da workspace.
   - Avalie se as novas structs, interfaces e separação de responsabilidades seguem estritamente os princípios SOLID e o mapeamento de recursos documentado na Wiki.
2. **Auto-Revisão de Código**:
   - Abra e leia o arquivo `agents/code-review-agent.md`.
   - Analise o código produzido em busca de memory leaks (C++), concorrência inadequada em chamadas IPC ou erros ignorados sem justificativa (Go).
3. **Auto-Revisão de UX (Frontend)**:
   - Abra e leia o arquivo `agents/ux-agent.md`.
   - Caso a tarefa envolva qualquer widget ou interface visual, avalie se os elementos gráficos respeitam o design system *Luminous Glass* (backdrop blur de 20px+, bordas de brilho de 1px e paleta de cores Midnight).

Após realizar as auto-revisões, o agente ativo deve incluir um resumo com o parecer de cada uma dessas três personas na mensagem de entrega final ao usuário.
