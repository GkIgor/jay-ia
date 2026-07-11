# ADR 0002: Ambiente Isolado Persistente

## Status

Aceita

## Contexto

Jay deve operar com independência e segurança, sem depender de acesso irrestrito ao host do usuário.

## Decisão

Jay executará em ambiente isolado, persistente, com usuário dedicado, sem privilégios administrativos e com filesystem próprio.

## Não decisão

A tecnologia concreta de isolamento ainda não foi escolhida.

## Consequências

- a arquitetura não depende de um runtime específico
- o isolamento passa a ser requisito estrutural
- futuras escolhas de implementação deverão respeitar os mesmos limites
