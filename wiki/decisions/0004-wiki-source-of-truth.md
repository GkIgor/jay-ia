# ADR 0004: Wiki como Fonte de Verdade

## Status

Aceita

## Contexto

O projeto depende de continuidade entre humanos e agentes de IA. Decisões importantes não podem ficar apenas em código, conversa ou contexto efêmero.

## Decisão

A wiki deste repositório é a fonte oficial de verdade para princípios, arquitetura, decisões e planejamento do projeto.

Quando algo não estiver documentado e a decisão for necessária, deve-se consultar um humano e registrar o resultado no local apropriado da wiki.

## Consequências

- agentes devem consultar a wiki antes de decidir
- divergências entre implementação e arquitetura precisam ser explicitadas
- `wiki/index.md` deve permanecer atualizado como índice vivo da documentação
