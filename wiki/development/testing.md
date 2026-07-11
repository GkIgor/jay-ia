# Testes

## Estado atual

Ainda não existe suíte formal de testes.

## Regra presente

Mudanças de arquitetura devem ser validadas primeiro por consistência documental:

- wiki coerente
- ADRs alinhados
- índice atualizado
- ausência de contradição entre visão, arquitetura e PRDs

## Critério mínimo atual

Uma mudança só está pronta quando:

- a decisão está documentada
- os documentos afetados permanecem coerentes entre si
- o índice continua navegável
- não existe conflito evidente entre README, architecture, specifications e ADRs

## Estratégia futura

Quando houver código, a estratégia de testes deverá separar ao menos:

- testes de Core
- testes de frontend
- testes de protocolo IPC
- testes de providers
- testes de integração entre Core e clientes

## Regra de arquitetura testável

A forma como o projeto for implementado deve preservar testabilidade por desacoplamento.

Se um componente não puder ser testado isoladamente, isso costuma indicar acoplamento excessivo.

## Evolução esperada

Quando houver código, este documento deverá separar testes de Core, frontend, integração e providers.
