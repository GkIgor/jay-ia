# OpenClaw

# OpenClaw

## Papel Futuro no Sistema

O OpenClaw não faz parte da arquitetura fundamental atual da Jay. 

Este documento existe apenas para documentar uma possível integração futura através da interface **Tool Provider** (ver [`tool-provider.md`](tool-provider.md)).

## Contexto

Embora o OpenClaw seja um projeto altamente compatível com a visão de longo prazo da Jay, utilizá-lo como camada obrigatória para a Fase 1 adicionaria complexidade desnecessária (Gateway, skills, ciclo de vida).

Quando houver necessidade de funcionalidades mais avançadas de orquestração de plugins externos com forte controle e isolamento, o OpenClaw poderá coexistir ou substituir executores mais simples, mantendo-se o protagonismo do Tool Provider.
