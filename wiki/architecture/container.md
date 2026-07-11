# Ambiente Isolado

## Objetivo

Jay deve operar em um ambiente persistente e isolado do sistema operacional do usuário.

## Requisitos arquiteturais

- ambiente próprio
- usuário dedicado
- ausência de privilégios administrativos
- filesystem próprio
- persistência de dados
- isolamento do host

## Observação importante

A tecnologia concreta de isolamento ainda não foi decidida.

Podman, LXC/LXD, systemd-nspawn ou outra alternativa poderão ser avaliados depois por ADR específico.

## Intenção

O isolamento não é detalhe operacional. Ele é parte da identidade da Jay como habitante de um ambiente próprio.
