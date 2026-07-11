# API de Frontend

## Objetivo

Definir o contrato entre o Core e qualquer frontend.

## Princípios

- o frontend é cliente
- o Core é headless
- o protocolo é o produto
- múltiplos frontends devem ser possíveis

## Papel do frontend

O frontend representa o corpo e a superfície local da Jay.

Ele existe para:

- apresentar avatar, voz e animações
- capturar entrada do usuário
- mediar acesso a recursos locais
- reportar estado ambiental não sensível

## Papel do Core

O Core continua sendo a autoridade sobre:

- identidade da Jay
- sessão
- memória
- conhecimento
- planejamento
- decisão

## Tipos de interação

O contrato entre Core e frontend deve cobrir quatro fluxos principais:

### 1. Conexão

O frontend inicia conexão, identifica-se e declara capacidades disponíveis.

Exemplos de capacidades:

- áudio de saída
- captura de microfone
- renderização de avatar
- notificações
- compartilhamento de tela mediante permissão

### 2. Eventos emitidos pelo Core

O Core pode enviar eventos como:

- fala
- expressão
- animação
- pedido de notificação
- pedido de captura de recurso
- consulta de estado do ambiente

### 3. Eventos emitidos pelo frontend

O frontend pode enviar:

- entrada textual do usuário
- transcrição de fala
- confirmação de permissão
- estado do ambiente
- erro de execução local
- atualização de capacidades

### 4. Respostas e resultados

Toda ação que exigir retorno deve permitir confirmação, falha e metadados de resultado.

## Capacidades esperadas

O frontend pode:

- conectar ao Core
- receber eventos
- enviar entrada do usuário
- reportar estado do ambiente
- solicitar permissões ao sistema

## Estado do ambiente

O frontend pode expor apenas metadados necessários para a Jay operar melhor, sem vazar conteúdo por padrão.

Exemplos aceitáveis:

- janela ativa
- tempo de inatividade
- modo fullscreen
- número de monitores

Exemplos sensíveis que exigem fluxo explícito:

- conteúdo da tela
- áudio do microfone
- câmera
- automação de mouse e teclado

## Restrições

- o frontend não acessa memória diretamente
- o frontend não toma decisões arquiteturais
- o frontend não define a identidade da Jay

## Regras de segurança

- permissões do sistema devem ser solicitadas pelo frontend, não pelo Core diretamente
- o frontend não deve ganhar acesso irrestrito apenas por estar conectado
- recursos sensíveis precisam de consentimento explícito do usuário

## Tolerância a falhas

O frontend deve poder cair e reconectar sem reinicializar a Jay.

O Core deve continuar funcional mesmo sem frontend conectado.

## Consequência arquitetural

Qualquer implementação de frontend que respeite este contrato deve ser capaz de representar a Jay sem exigir mudança estrutural no Core.
