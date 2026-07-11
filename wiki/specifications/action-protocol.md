# Protocolo de Ações

## Objetivo

Definir como a Jay expressa intenções para componentes externos.

## Modelo conceitual

A Jay não executa interface diretamente.

Ela emite ações como:

- `avatar.wave`
- `avatar.smile`
- `speech.start`
- `notify`
- `request.microphone`

## Conceito central

O protocolo descreve **o que** a Jay quer que aconteça.

Ele não descreve **como** cada frontend deve renderizar ou executar isso.

## Estrutura conceitual da mensagem

Toda ação deve poder carregar ao menos:

- tipo da ação
- identificador único
- origem
- instante
- carga útil
- correlação opcional com resposta

## Categorias de ação

### Ações de avatar

Representam expressões, gestos, postura e foco visual.

Exemplos:

- `avatar.wave`
- `avatar.smile`
- `avatar.look_left`
- `avatar.idle`

### Ações de fala

Representam intenção de comunicar algo ao usuário.

Exemplos:

- `speech.start`
- `speech.chunk`
- `speech.end`

### Ações de interface

Representam pedidos de apresentação local.

Exemplos:

- `notify`
- `window.show`
- `window.hide`

### Ações de permissão

Representam pedidos para acesso a recursos sensíveis.

Exemplos:

- `request.microphone`
- `request.screen`
- `request.camera`

### Ações de consulta

Representam pedidos de estado ao frontend.

Exemplos:

- `system.state`
- `frontend.capabilities`

## Propriedades desejadas

- formato serializável
- legível por humanos
- independente de frontend específico
- extensível

## Respostas

Quando aplicável, uma ação deve admitir retorno em três estados principais:

- aceita
- recusada
- falhou

Também deve ser possível devolver:

- motivo
- metadados
- resultado parcial

## Composição

O protocolo deve permitir sequências curtas de ações relacionadas.

Exemplo conceitual:

```text
avatar.look_left
↓
avatar.smile
↓
speech.start
↓
speech.end
```

Isso permite que o Core descreva roteiros sem assumir detalhes de engine, toolkit ou plataforma.

## Regra

O protocolo descreve intenção, não implementação visual.
