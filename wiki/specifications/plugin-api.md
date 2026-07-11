# API de Plugins

## Objetivo

Definir como extensões e providers interagem com a Jay sem acoplamento indevido ao Core.

## Direção atual

Plugins devem conversar com interfaces estáveis, não com detalhes internos de implementação.

## Papel dos plugins

Plugins existem para estender capacidades da Jay sem reescrever o Core.

Eles podem representar, por exemplo:

- providers de ferramentas
- providers de LLM
- providers de voz
- integrações externas
- mecanismos de conhecimento
- automações específicas

## Regras arquiteturais

- plugins dependem de contratos, não de componentes concretos
- plugins não redefinem identidade, memória ou filosofia da Jay
- plugins devem ter responsabilidade clara
- falha de plugin não deve comprometer o Core inteiro

## Capacidades mínimas esperadas

Um plugin deve poder declarar, em nível conceitual:

- nome
- tipo
- versão
- capacidades expostas
- requisitos
- limitações

## Tipos iniciais de extensão

### Tool Provider

Executa ferramentas, automações ou chamadas externas.

### LLM Provider

Expõe modelos de linguagem para comunicação e raciocínio.

### Voice Provider

Lida com TTS, STT ou fluxo de áudio relacionado.

### Knowledge Provider

Expõe fontes ou mecanismos de recuperação de conhecimento.

## Expectativas

- contrato claro
- isolamento de responsabilidade
- versionamento futuro
- capacidade de adicionar novos providers de ferramenta, LLM, voz ou conhecimento

## Ciclo de vida esperado

Em termos arquiteturais, o Core deve conseguir:

- descobrir plugins disponíveis
- carregar capacidades
- usar interfaces expostas
- isolar falhas
- desabilitar plugins problemáticos

## Limites

Esta especificação ainda não define:

- formato definitivo de empacotamento
- ABI ou mecanismo de carregamento
- política de assinatura
- versionamento semântico obrigatório

## Observação

O desenho detalhado da API de plugins ainda está em aberto.
