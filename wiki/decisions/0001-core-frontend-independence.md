# ADR 0001: Independência entre Core e Frontend

## Status

Aceita

## Contexto

Jay precisa existir como agente persistente, inclusive sem interface gráfica. O frontend representa corpo e interação local, mas não deve concentrar identidade, memória ou decisão.

## Decisão

O Core da Jay será completamente independente de qualquer frontend.

O frontend será tratado como cliente conectado por protocolo definido.

## Consequências

- o Core pode rodar headless
- o frontend pode reiniciar sem derrubar a Jay
- múltiplos frontends tornam-se possíveis
- o protocolo Core ↔ frontend torna-se peça central da arquitetura
