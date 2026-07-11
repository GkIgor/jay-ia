# API de Frontend

## Objetivo

Definir o contrato entre o Core e qualquer frontend.

## Princípios

- o frontend é cliente
- o Core é headless
- o protocolo é o produto
- múltiplos frontends devem ser possíveis

## Capacidades esperadas

O frontend pode:

- conectar ao Core
- receber eventos
- enviar entrada do usuário
- reportar estado do ambiente
- solicitar permissões ao sistema

## Restrições

- o frontend não acessa memória diretamente
- o frontend não toma decisões arquiteturais
- o frontend não define a identidade da Jay
