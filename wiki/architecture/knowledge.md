# Conhecimento

## Responsabilidade

O subsistema de conhecimento armazena aquilo que a Jay considera verdade em domínios que administra.

Seu papel é fornecer base autoritativa para raciocínio e resposta em temas relevantes ao universo da Jay.

## Regra de autoridade

Para domínios como programação, Linux, projetos, documentação, workspace e preferências do usuário, a fonte principal deve ser a base da própria Jay.

Se o conhecimento não existir:

1. Jay admite que não sabe.
2. Pergunta ao usuário ou pesquisa.
3. Curadoria e registro.
4. Passa a reutilizar esse conhecimento no futuro.

## Propriedades desejadas

O conhecimento deve ser:

- recuperável
- rastreável
- revisável
- associado a fontes
- capaz de expressar confiança ou incerteza

## Relação com memória

Conhecimento e memória não são equivalentes.

Conhecimento responde a perguntas como:

- o que é Go?
- como funciona uma ferramenta?
- qual protocolo foi definido?

Memória responde a perguntas como:

- o que o usuário prefere?
- o que já aconteceu entre Jay e usuário?
- qual projeto é recorrente?

## Relação com a LLM

A LLM pode ajudar a interpretar ou sintetizar conhecimento.

Ela não substitui a base autoritativa da Jay nos domínios administrados por ela.

## Consequências

- maior consistência de personalidade
- rastreabilidade
- atualização contínua
- independência parcial de modelos

## Risco evitado por esta arquitetura

Sem uma base própria, a Jay tende a oscilar conforme o modelo, a data de treinamento e o provider.

Com base própria, ela aprende e mantém continuidade de forma mais estável.

## Distinção importante

A wiki deste repositório documenta o projeto.

A base interna de conhecimento da Jay continua como tema aberto de evolução.
