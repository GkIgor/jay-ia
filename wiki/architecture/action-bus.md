# Action Bus

## Responsabilidade

O Action Bus é o mecanismo conceitual pelo qual a Jay expressa ações e eventos para componentes externos.

## Uso principal

- comunicação com frontends
- solicitações de permissão
- notificações
- ações visuais
- pedidos de recursos locais

## Papel na arquitetura

O Action Bus separa decisão interna de execução concreta.

Ele permite que o Core:

- descreva intenções
- peça recursos
- coordene respostas visuais e sonoras
- consulte estado externo

sem depender diretamente de toolkit gráfico, API de desktop ou implementação de frontend específica.

## Relação com o frontend

O frontend interpreta as ações recebidas e devolve resultados, falhas, permissões ou estados do ambiente.

## Relação com segurança

Como o Core não acessa recursos sensíveis diretamente, o Action Bus se torna ponto importante de mediação arquitetural entre intenção e permissão.

## Propriedade central

Jay pede.

O cliente interpreta.

Isso mantém separação entre intenção e execução concreta.
