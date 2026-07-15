# ADR 0003: OpenClaw como Executor de Ferramentas

## Status

Postergada (Superseded by ADR 0008 for Phase 1)

## Contexto

OpenClaw é valioso como mecanismo de execução de ferramentas, mas não deve concentrar planejamento, identidade ou memória da Jay.

## Decisão

OpenClaw será tratado como provider de execução dentro do barramento de ferramentas da Jay.

Ele não será o cérebro do sistema.

## Consequências

- Jay pode evoluir sem depender estruturalmente de OpenClaw
- outros providers poderão coexistir ou substituí-lo
- responsabilidades ficam mais claras entre Core, memória, conhecimento e execução
