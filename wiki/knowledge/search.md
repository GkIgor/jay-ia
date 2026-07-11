# Busca

## Objetivo

Busca é o mecanismo de recuperação de contexto, não um substituto para curadoria documental.

## Regra

Agentes não devem carregar a wiki inteira no contexto por padrão.

Eles devem localizar e consultar apenas os trechos relevantes para a decisão atual.

## Consequência

Boa organização da wiki é requisito para boa recuperação de contexto.

## Escopos de busca

Existem ao menos dois escopos relevantes:

### Busca na wiki do projeto

Usada para:

- decisões arquiteturais
- responsabilidades de componentes
- contratos técnicos
- planejamento de fases

### Busca na base futura da Jay

Usada para:

- conhecimento aprendido
- fatos operacionais
- documentação curada
- memória consultável

## Estratégia desejada

Busca deve priorizar:

- relevância
- escopo correto
- contexto mínimo suficiente
- rastreabilidade da origem

## Regra para agentes

Antes de responder ou decidir, o agente deve confirmar se está buscando no lugar correto:

- wiki do projeto, quando a pergunta é sobre arquitetura e decisão
- base da Jay, quando a pergunta é sobre conhecimento operacional aprendido

## Anti-padrão

Usar contexto excessivo em vez de recuperação seletiva tende a piorar custo, latência e precisão.
