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

## Propriedades desejadas

- comunicação local simples de debugar
- sem dependência de stack gráfica
- boa compatibilidade entre Go e clientes externos
- transporte estável para múltiplos tipos de mensagem

## Tipos de mensagem

O canal IPC deve suportar pelo menos:

- eventos emitidos pelo Core
- comandos emitidos pelo frontend
- respostas correlacionadas
- erros estruturados
- anúncio de capacidades
- heartbeat ou mecanismo equivalente de presença

## Fluxo conceitual de conexão

```text
frontend conecta
↓
frontend anuncia capacidades
↓
Core aceita sessão
↓
troca de eventos e comandos
↓
desconexão ou reconexão
```

## Sessão

Cada cliente conectado deve ser tratado como sessão independente.

Isso permite:

- múltiplos frontends
- reconexão
- isolamento de falhas
- capacidades diferentes por cliente

## Tratamento de falhas

O canal IPC deve permitir distinguir:

- falha de transporte
- mensagem inválida
- ação recusada
- recurso indisponível
- permissão negada

## Observabilidade

Mensagens devem ser inspecionáveis sem ferramentas excessivamente complexas.

Isso facilita:

- depuração
- testes
- desenvolvimento de novos clientes
- análise de regressões

## Limites desta especificação

Este documento define a fronteira e as propriedades do canal.

Ele não congela ainda:

- framing final
- política de autenticação local
- serialização definitiva além da direção atual
- detalhes binários de streaming de mídia
