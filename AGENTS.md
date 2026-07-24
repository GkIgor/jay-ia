# Repository Guidelines

## Wiki-First Rule

This repository exists to build **Jay IA**, and [`wiki/`](wiki/) is the project’s source of truth for agents and humans. When implementation, planning, or a technical decision is unclear, consult the wiki first. If the answer is not documented, **do not invent policy**: ask a human, then record the answer in the correct wiki file.

Treat [`wiki/README.md`](wiki/README.md) as the project’s initial context and [`wiki/index.md`](wiki/index.md) as the living index of the book. If a page is added, renamed, split, or becomes important to navigation, update [`wiki/index.md`](wiki/index.md) in the same change.

## Required Reading Order

Before making decisions, read in this order:
- [`wiki/README.md`](wiki/README.md): project identity, goals, and initial architecture.
- [`wiki/index.md`](wiki/index.md): navigation and documentation hierarchy.
- [`wiki/vision/`](wiki/vision/): non-negotiable principles and non-goals.
- [`wiki/architecture/`](wiki/architecture/) and [`wiki/specifications/`](wiki/specifications/): component responsibilities and contracts.
- [`wiki/prds/`](wiki/prds/), [`wiki/phases/`](wiki/phases/), and [`wiki/decisions/`](wiki/decisions/): execution plan and recorded decisions.

Use the Karpathy-style documentation model here: the index points to the right chapter, and decisions should be grounded in the relevant chapter, not scattered through chat history.

## How Agents Should Work

For any task:
1. Find the relevant wiki pages before coding or proposing architecture.
2. Follow documented decisions even if implementation is missing.
3. If wiki and code disagree, assume the wiki is intended truth and flag the mismatch.
4. If knowledge is missing, ask a human for direction.
5. After clarification, register the result in the appropriate wiki location.
6. If there is any pending item, deferred improvement, test gap, mock, `TODO`, or unresolved follow-up, record it in the wiki and update that record when it is resolved or changed.
7. **Regra de Alteração de Código (CRÍTICA - OBRIGATÓRIA)**: Sempre que alterar qualquer linha de código (seja em Go ou C++), o agente **deve obrigatoriamente**:
   - Executar o comando de build correspondente (`go build` / `make build` ou `cmake --build build`) para garantir que o código compila perfeitamente e sem avisos.
   - Formatar o arquivo modificado usando as ferramentas padrão do repositório. Para Go, use `gofmt`. Para C++, **é obrigatório** executar `clang-format -i <arquivo>` seguindo a especificação do `.clang-format` na raiz do projeto (que define o estilo Google e **recuo estrito de 2 espaços**).
   - Nenhuma tarefa ou código C++ será considerado concluído ou passível de commit sem a execução prévia do `clang-format` no arquivo.
   - **Atenção especial ao C++**: A configuração do `clang-format` deve sempre respeitar a legibilidade e a semântica de espaçamento em blocos e namespaces, garantindo que o código não fique aglomerado sem quebras de linha lógicas (espaços verticais).
8. **Regra de Comentários em Código (CRÍTICA)**:
   - **Proibido adicionar comentários explicativos de código óbvio**, comentários inline redundantes ou cabeçalhos decorativos.
   - O código deve ser auto-explicativo por meio de nomenclaturas explícitas e legíveis.
   - Comentários são permitidos **exclusivamente** para documentar decisões técnicas não-óbvias (o *motivo*/*WHY*, e nunca o *o quê*/*WHAT*). Se o código for legível por si só, **não adicione comentários**.

Good destinations:
- vision changes: [`wiki/vision/`](wiki/vision/)
- architecture decisions: [`wiki/architecture/`](wiki/architecture/)
- formal decisions: [`wiki/decisions/`](wiki/decisions/)
- protocols/formats: [`wiki/specifications/`](wiki/specifications/)
- delivery planning: [`wiki/prds/`](wiki/prds/) or [`wiki/phases/`](wiki/phases/)
- operational follow-ups, deferred work, and test/mocks/TODO tracking: the most relevant wiki page for the topic, or a dedicated entry in [`wiki/future/`](wiki/future/), [`wiki/development/`](wiki/development/), or the applicable ADR/PRD

## Processo de Planejamento e Sincronização da Wiki (CRÍTICO)

O planejamento e a execução do agente devem sempre refletir e atualizar a wiki.
Siga estritamente as regras abaixo:

1. **Plano na Wiki**: Todo progresso de planejamento e evolução do projeto deve ser registrado na wiki (como em `phases/`, `prds/` ou documentação do componente).
2. **Steps e Subtasks**: Para cada plano em andamento, quebre a execução estruturalmente em *steps*, *tasks* e *subtasks* granulares documentadas na própria wiki (ou em um log vinculado a ela).
3. **Atualização Pós-Modificação**: Sempre que finalizar uma modificação de código ou comportamento sistêmico, consulte as páginas da wiki correspondentes e atualize-as para garantir que refletem o estado funcional atual.
4. **Wiki Sempre Atualizada**: A wiki é a fonte primária da verdade. É inaceitável que o código evolua enquanto a documentação fica defasada. A atualização não é uma etapa opcional, mas o pilar obrigatório do seu ciclo contínuo.

## Project Structure

Current repository structure is documentation-first:
- [`wiki/`](wiki/): source of truth and planning system
- [`wiki/knowledge/`](wiki/knowledge/): how Jay learns and curates knowledge
- [`wiki/development/`](wiki/development/): setup, style, testing, release notes

Application code and tests are now established. Their structure is documented in [`wiki/development/project-structure.md`](wiki/development/project-structure.md).

## Editing, Validation, and Commits

Keep Markdown concise, explicit, and non-duplicative. Prefer kebab-case filenames like `action-bus.md`; keep ADRs numbered like `0001-core-frontend-independence.md`.

Before finishing a change:
- verify the relevant wiki page was consulted
- update the wiki if a new decision was made
- update [`wiki/index.md`](wiki/index.md) if navigation changed
- avoid leaving important decisions only in commits or conversations

Use short imperative commits, for example: `docs: record memory ownership decision`.
