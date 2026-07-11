# Core

## Responsabilidade

O Core é a identidade operacional da Jay.

Ele coordena:

- personalidade
- conversa
- planejamento
- memória
- conhecimento
- aprendizado
- execução de ações
- estado interno

## Papel na arquitetura

O Core é o ponto de coordenação do sistema.

Ele integra:

- percepção
- decisão
- continuidade
- execução indireta
- relação com clientes e providers

## Relações principais

O Core:

- consulta memória para preservar continuidade
- consulta conhecimento para responder com base autoritativa
- utiliza o Planner para decidir próximos passos
- aciona providers para executar ações
- emite eventos para frontends por protocolo

## Propriedades desejadas

O Core deve permanecer:

- headless
- persistente
- observável
- desacoplado de frontend específico
- desacoplado de provider específico

## O que o Core não faz

- não desenha interface
- não depende de bibliotecas gráficas
- não acessa diretamente recursos do host sem passar por adaptadores
- não delega sua identidade a uma LLM

## Propriedades esperadas

- execução persistente como daemon
- operação headless
- reconexão de clientes
- suporte a múltiplos providers
- previsibilidade e rastreabilidade
