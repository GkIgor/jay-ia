# Indexação

## Objetivo

Permitir que agentes e humanos encontrem contexto relevante sem ler toda a wiki a cada tarefa.

## Papel do `wiki/index.md`

O `wiki/index.md` funciona como índice vivo da documentação, no estilo índice de livro.

Ele deve:

- apontar para as seções corretas
- explicar a ordem de leitura
- refletir mudanças estruturais da wiki

## Papel arquitetural do índice

O índice não é apenas uma página de navegação.

Ele funciona como mecanismo de entrada para agentes, reduzindo ambiguidade sobre onde cada decisão deve ser encontrada.

## Regras de manutenção

`wiki/index.md` deve ser atualizado quando houver:

- criação de nova seção relevante
- mudança de estrutura
- renomeação de documentos importantes
- alteração na ordem de leitura recomendada

## Indexação para agentes

Além da leitura humana, a wiki deve ser organizável para recuperação parcial de contexto.

Isso implica:

- títulos claros
- documentos focados
- responsabilidade única por página
- ausência de duplicação desnecessária

## Critério de qualidade

Um agente deve conseguir localizar a página certa com pouca exploração e baixo risco de interpretar o documento errado.

## Direção futura

Além do índice humano, o projeto deverá evoluir para mecanismos de recuperação contextual adequados ao uso por agentes.
