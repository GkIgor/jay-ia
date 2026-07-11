# Planner

## Responsabilidade

O Planner transforma intenção em plano de ação.

Ele decide como a Jay deve:

- responder
- consultar memória
- consultar conhecimento
- aprender algo novo
- executar ferramentas
- pedir ajuda ao usuário

## Papel na arquitetura

O Planner é o componente que conecta percepção, memória, conhecimento e execução.

Ele não é apenas um gerador de passos.

Ele é responsável por decidir:

- se a Jay já sabe o suficiente para responder
- se precisa consultar memória
- se precisa consultar conhecimento
- se precisa aprender algo antes de agir
- se deve executar ferramentas
- se deve pedir confirmação humana

## Entradas conceituais

O Planner pode operar com base em:

- entrada do usuário
- estado atual do sistema
- memória durável
- contexto de trabalho
- conhecimento recuperado
- capacidades disponíveis
- limites e permissões

## Saídas conceituais

O resultado do Planner pode incluir:

- resposta direta
- pergunta de esclarecimento
- plano de múltiplas etapas
- pedido de permissão
- delegação para provider de ferramenta
- acionamento do pipeline de aprendizado

## Loop conceitual

```text
Perceber
↓
Planejar
↓
Consultar memória
↓
Consultar conhecimento
↓
Executar
↓
Avaliar
↓
Aprender
```

## Modos de operação

### Resposta imediata

Quando a Jay já possui contexto e conhecimento suficientes.

### Consulta

Quando precisa recuperar informação antes de responder.

### Aprendizado

Quando o conhecimento necessário ainda não existe.

### Execução

Quando a tarefa exige agir no ambiente.

### Escalada humana

Quando existe ambiguidade, risco ou ausência de decisão documentada.

## Relação com a LLM

A LLM pode auxiliar em:

- decomposição de tarefas
- interpretação de linguagem natural
- síntese de alternativas
- planejamento textual

A LLM não deve, sozinha, definir o comportamento estrutural do Planner.

## Critérios de qualidade

Um bom plano deve ser:

- rastreável
- revisável
- compatível com permissões disponíveis
- coerente com memória e conhecimento existentes
- interrompível quando houver risco ou ambiguidade

## Observações

A LLM pode auxiliar o Planner em raciocínio e interpretação, mas o ciclo do agente não deve depender apenas de prompting.
