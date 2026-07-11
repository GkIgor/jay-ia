# Segurança

## Princípios

- mínimo privilégio
- isolamento por padrão
- permissões explícitas para recursos sensíveis
- desacoplamento entre Core e host
- rastreabilidade de decisões e ações

## Regras de alto nível

- Jay não opera com sudo como requisito normal.
- Acesso a tela, microfone e webcam deve ocorrer via frontend e mecanismos explícitos de permissão.
- O Core não deve acessar diretamente bibliotecas de sistema gráfico.
- Limites devem existir para evitar que a Jay comprometa o host.

## Observação

Segurança aqui não é apenas proteção contra ataque externo. É também proteção contra ação incorreta do próprio agente.
