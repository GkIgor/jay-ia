# API de Plugins

## Objetivo

Definir como extensões e providers interagem com a Jay sem acoplamento indevido ao Core.

## Direção atual

Plugins devem conversar com interfaces estáveis, não com detalhes internos de implementação.

## Expectativas

- contrato claro
- isolamento de responsabilidade
- versionamento futuro
- capacidade de adicionar novos providers de ferramenta, LLM, voz ou conhecimento

## Observação

O desenho detalhado da API de plugins ainda está em aberto.
