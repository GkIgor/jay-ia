# Repository Guidelines

## Wiki-First Rule

This repository exists to build **Jay IA**, and `wiki/` is the project’s source of truth for agents and humans. When implementation, planning, or a technical decision is unclear, consult the wiki first. If the answer is not documented, **do not invent policy**: ask a human, then record the answer in the correct wiki file.

Treat `wiki/README.md` as the project’s initial context and `wiki/index.md` as the living index of the book. If a page is added, renamed, split, or becomes important to navigation, update `wiki/index.md` in the same change.

## Required Reading Order

Before making decisions, read in this order:
- `wiki/README.md`: project identity, goals, and initial architecture.
- `wiki/index.md`: navigation and documentation hierarchy.
- `wiki/vision/`: non-negotiable principles and non-goals.
- `wiki/architecture/` and `wiki/specifications/`: component responsibilities and contracts.
- `wiki/prds/`, `wiki/phases/`, and `wiki/decisions/`: execution plan and recorded decisions.

Use the Karpathy-style documentation model here: the index points to the right chapter, and decisions should be grounded in the relevant chapter, not scattered through chat history.

## How Agents Should Work

For any task:
1. Find the relevant wiki pages before coding or proposing architecture.
2. Follow documented decisions even if implementation is missing.
3. If wiki and code disagree, assume the wiki is intended truth and flag the mismatch.
4. If knowledge is missing, ask a human for direction.
5. After clarification, register the result in the appropriate wiki location.

Good destinations:
- vision changes: `wiki/vision/`
- architecture decisions: `wiki/architecture/`
- formal decisions: `wiki/decisions/`
- protocols/formats: `wiki/specifications/`
- delivery planning: `wiki/prds/` or `wiki/phases/`

## Project Structure

Current repository structure is documentation-first:
- `wiki/`: source of truth and planning system
- `wiki/knowledge/`: how Jay learns and curates knowledge
- `wiki/development/`: setup, style, testing, release notes

Application code and tests are not established yet. When they are added, document their structure in `wiki/development/project-structure.md`.

## Editing, Validation, and Commits

Keep Markdown concise, explicit, and non-duplicative. Prefer kebab-case filenames like `action-bus.md`; keep ADRs numbered like `0001-core-frontend-independence.md`.

Before finishing a change:
- verify the relevant wiki page was consulted
- update the wiki if a new decision was made
- update `wiki/index.md` if navigation changed
- avoid leaving important decisions only in commits or conversations

Use short imperative commits, for example: `docs: record memory ownership decision`.
