# Frontend

## Responsabilidade

O frontend é o corpo da Jay.

Ele representa visualmente e sonoramente o agente, mas não decide por ele.

## Responsabilidades esperadas

- janela
- avatar 2D ou 3D
- áudio
- microfone
- animações
- renderização
- interpretação de ações vindas do Core
- solicitação de permissões ao sistema

## Regras arquiteturais

- o frontend é cliente do Core
- pode ser reiniciado sem derrubar a Jay
- nunca acessa diretamente memória ou conhecimento
- toda comunicação com o Core ocorre por protocolo definido

## Direção atual

O primeiro frontend deverá ser implementado em C++, mas a arquitetura deve permitir outros clientes no futuro.
