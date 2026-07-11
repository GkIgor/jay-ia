# IPC

## Objetivo

Definir a fronteira de comunicação entre o Core e clientes externos, especialmente frontends.

## Direção atual

O protocolo deve ser simples, observável e desacoplado da implementação interna.

JSON sobre socket local é a direção atual mais coerente, mas esta escolha ainda não deve ser tratada como decisão irreversível sem ADR.

## Requisitos

- suporte a eventos
- suporte a comandos e respostas
- reconexão
- mensagens inspecionáveis
- baixo acoplamento entre linguagens
