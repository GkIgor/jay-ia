# Action Protocol

## Objetivo

Definir como a Jay expressa intenções para componentes externos.

## Modelo conceitual

A Jay não executa interface diretamente.

Ela emite ações como:

- `avatar.wave`
- `avatar.smile`
- `speech.start`
- `notify`
- `request.microphone`

## Propriedades desejadas

- formato serializável
- legível por humanos
- independente de frontend específico
- extensível

## Regra

O protocolo descreve intenção, não implementação visual.
