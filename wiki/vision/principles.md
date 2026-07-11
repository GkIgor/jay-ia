# Princípios

## Finalidade

Esta seção registra os princípios inegociáveis da Jay. Eles orientam decisões de arquitetura, produto e implementação.

## Princípios Fundamentais

### 1. Jay é um agente persistente

Jay não deve ser tratada como uma sessão temporária de chat.

Ela possui continuidade, estado, memória e identidade próprias.

### 2. A personalidade da Jay é separada da LLM

Modelos de linguagem são ferramentas de raciocínio e comunicação.

Eles não definem quem a Jay é.

### 3. Conhecimento autoritativo pertence à própria Jay

Em domínios administrados por ela, a fonte principal de conhecimento deve ser sua própria base, e não o conhecimento treinado da LLM.

Se não souber algo, Jay deve admitir isso, aprender e registrar.

### 4. O isolamento é parte da identidade do projeto

Jay vive em um ambiente próprio e limitado.

Essa fronteira existe para preservar segurança, previsibilidade e independência arquitetural.

### 5. O Core deve ser independente de interface

Frontend é corpo.

Core é identidade, coordenação e decisão.

Ambos devem permanecer desacoplados.

### 6. A Wiki é a fonte de verdade do projeto

Decisões importantes não devem existir apenas em código, conversas ou memória informal.

O projeto evolui por documentação curada e rastreável.
