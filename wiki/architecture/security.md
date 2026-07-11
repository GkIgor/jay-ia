# Segurança

## Princípios

- mínimo privilégio
- isolamento por padrão
- permissões explícitas para recursos sensíveis
- desacoplamento entre Core e host
- rastreabilidade de decisões e ações

## Escopo

Segurança, neste projeto, inclui:

- proteção do host
- limitação de dano causado pelo próprio agente
- proteção de recursos sensíveis
- previsibilidade operacional
- fronteiras claras entre componentes

## Regras de alto nível

- Jay não opera com sudo como requisito normal.
- Acesso a tela, microfone e webcam deve ocorrer via frontend e mecanismos explícitos de permissão.
- O Core não deve acessar diretamente bibliotecas de sistema gráfico.
- Limites devem existir para evitar que a Jay comprometa o host.

## Fronteiras de confiança

### Core

É a autoridade comportamental da Jay, mas não deve receber acesso irrestrito ao sistema.

### Frontend

Pode mediar acesso a recursos locais, mas não deve concentrar identidade ou memória.

### Providers de ferramenta

Devem operar dentro de limites claros e substituíveis.

### Ambiente isolado

É a principal fronteira estrutural de contenção de dano.

## Recursos sensíveis

Devem ser tratados com fluxos explícitos de consentimento quando aplicável:

- microfone
- câmera
- tela
- automação de entrada
- notificações intrusivas

## Riscos arquiteturais relevantes

- acoplamento entre Core e recursos do host
- expansão de privilégios por conveniência
- uso de ferramentas sem fronteira clara
- conhecimento incorreto levando a ação insegura
- ambiguidade operacional em tarefas sensíveis

## Resposta arquitetural esperada

Para mitigar esses riscos, a arquitetura deve favorecer:

- isolamento
- providers substituíveis
- pedidos explícitos de permissão
- documentação como fonte de verdade
- possibilidade de interromper ou escalar decisões

## Observação

Segurança aqui não é apenas proteção contra ataque externo. É também proteção contra ação incorreta do próprio agente.
