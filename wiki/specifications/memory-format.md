# Formato de Memória

## Objetivo

Descrever como informações de memória devem ser organizadas conceitualmente.

## Categorias mínimas

- memória durável
- contexto de trabalho
- estado transitório

## Definições

### Memória durável

Registra fatos persistentes relevantes para continuidade da Jay.

Exemplos:

- preferências do usuário
- relações importantes
- hábitos recorrentes
- contexto profissional duradouro

### Contexto de trabalho

Registra o que está sendo tratado agora ou em janela de tempo curta.

Exemplos:

- tarefa atual
- objetivo em andamento
- conversa recente
- foco temporário

### Estado transitório

Registra condições operacionais momentâneas.

Exemplos:

- compilando
- aguardando resposta
- executando ação
- baixando arquivos

## Metadados desejáveis

- origem
- data
- relevância
- confiança
- possibilidade de revisão

## Distinções obrigatórias

Memória não é:

- base de conhecimento
- log bruto
- histórico integral de chat
- estado operacional efêmero sem curadoria

## Relação com conhecimento

Conhecimento responde a algo como:

> "o que a Jay considera verdade sobre um domínio"

Memória responde a algo como:

> "o que a Jay lembra sobre a continuidade do usuário e de sua própria experiência"

## Requisitos arquiteturais

- memória deve ser consultável pelo Planner
- memória deve influenciar continuidade da personalidade
- memória deve poder ser revisada e atualizada
- memória deve evitar crescimento descontrolado de ruído irrelevante

## Regra

Formato de memória não deve ser confundido com base de conhecimento nem com histórico bruto de conversa.
