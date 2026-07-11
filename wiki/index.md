# Wiki Index

Bem-vindo à Wiki da **Jay**.

Esta Wiki representa a **fonte oficial de conhecimento** do projeto.

Toda decisão arquitetural, funcional ou técnica deve ser documentada aqui antes ou durante sua implementação. O código deve refletir a Wiki, e não o contrário.

Este documento serve como ponto de entrada para desenvolvedores, colaboradores e agentes de IA.

---

# Objetivo da Wiki

A Wiki possui quatro objetivos principais:

* Centralizar toda a documentação do projeto.
* Explicar a arquitetura da Jay de forma clara e independente da implementação.
* Servir como fonte de contexto para agentes de IA durante o desenvolvimento.
* Registrar decisões importantes ao longo da evolução do projeto.

A Wiki não substitui o código.

Ela explica **por que** as coisas existem, **qual problema resolvem** e **como os componentes se relacionam**.

---

# Ordem de leitura recomendada

Ao iniciar o projeto ou antes de implementar qualquer funcionalidade, recomenda-se seguir a seguinte ordem de leitura.

## 1. README

**Arquivo**

`README.md`

Apresenta a visão geral da Jay.

Responde perguntas como:

* O que é a Jay?
* Qual seu propósito?
* Qual a filosofia do projeto?
* Como a arquitetura está organizada?
* Quais são as fases de desenvolvimento?

---

## 2. Vision

`vision/`

Define os princípios fundamentais do projeto.

Esta seção responde perguntas como:

* O que a Jay deve ser?
* O que a Jay nunca deve se tornar?
* Quais decisões são inegociáveis?

Os documentos desta pasta representam a identidade do projeto e tendem a mudar muito pouco ao longo do tempo.

---

## 3. Architecture

`architecture/`

Explica cada componente da arquitetura.

Cada documento descreve:

* responsabilidades
* limitações
* relação com outros componentes
* objetivos

A implementação pode mudar.

A responsabilidade de cada componente deve permanecer consistente.

---

## 4. Specifications

`specifications/`

Contém as especificações técnicas.

Esta seção descreve protocolos, contratos e formatos utilizados pelos componentes.

Exemplos:

* protocolo IPC
* mensagens
* Action Bus
* APIs internas
* formatos de armazenamento

---

## 5. PRDs

`prds/`

Os PRDs descrevem como cada fase será implementada.

Cada PRD representa um plano de desenvolvimento.

Após a conclusão de uma fase, o PRD torna-se um documento histórico.

---

## 6. Decisions

`decisions/`

Armazena todas as decisões arquiteturais do projeto.

Cada documento responde:

* Qual problema existia?
* Quais alternativas foram avaliadas?
* Qual decisão foi tomada?
* Quais consequências essa decisão possui?

Nenhuma decisão importante deve existir apenas em conversas.

---

## 7. Development

`development/`

Documentação destinada aos desenvolvedores.

Inclui:

* preparação do ambiente
* convenções
* estrutura do projeto
* testes
* releases

---

## 8. Future

`future/`

Área destinada para pesquisas, ideias e funcionalidades futuras.

Nada presente nesta seção representa um compromisso de implementação.

Seu objetivo é registrar possibilidades para evolução do projeto.

---

# Estrutura da Wiki

```text
wiki/
│
├── README.md
├── index.md
│
├── vision/
├── architecture/
├── specifications/
├── phases/
├── prds/
├── decisions/
├── knowledge/
├── development/
└── future/
```

Cada diretório possui uma responsabilidade única.

Evite duplicar informações entre documentos.

Quando necessário, faça referências cruzadas.

---

# Organização da documentação

A documentação deve seguir uma hierarquia simples.

## Visão

Explica o propósito do projeto.

## Arquitetura

Explica os componentes.

## Especificação

Explica como os componentes funcionam.

## Implementação

Explica como desenvolver, validar e evoluir o projeto sem violar a arquitetura definida.

Essa separação reduz inconsistências e facilita futuras alterações.

---

# Convenções

Todos os documentos devem utilizar linguagem simples, objetiva e técnica.

Sempre que possível:

* explicar responsabilidades antes da implementação;
* utilizar exemplos;
* evitar ambiguidades;
* justificar decisões importantes;
* manter consistência entre documentos.

Diagramas em ASCII são recomendados quando facilitarem a compreensão.

---

# Fonte de Verdade

Esta Wiki representa a principal fonte de conhecimento do projeto.

Ao existir divergência entre:

* código;
* documentação externa;
* conversas;
* decisões antigas;

a Wiki deve ser considerada a referência principal.

Caso a implementação deixe de refletir a documentação, a divergência deve ser corrigida.

---

# Utilização por Agentes de IA

Esta documentação foi organizada para servir também como contexto para agentes de IA.

Antes de implementar qualquer funcionalidade, recomenda-se que o agente siga o seguinte fluxo:

1. Ler `index.md`.
2. Ler `README.md`.
3. Ler os documentos da pasta `vision/`.
4. Ler a arquitetura relacionada à funcionalidade.
5. Ler o PRD correspondente.
6. Consultar decisões arquiteturais existentes.
7. Somente então iniciar a implementação.

Nenhum agente deve assumir comportamentos ou arquiteturas que não estejam documentados.

Quando identificar inconsistências, o agente deve priorizar a atualização da documentação antes da implementação.

---

# Evolução da Wiki

A Wiki evolui junto com o projeto.

Novos documentos podem ser adicionados sempre que necessário.

Entretanto, deve-se preservar uma organização simples e consistente.

Cada documento deve possuir uma responsabilidade clara.

Quando um documento crescer excessivamente, ele deve ser dividido em novos documentos especializados.

---

# Filosofia

A documentação da Jay não existe apenas para registrar decisões técnicas.

Ela existe para preservar a identidade do projeto.

A implementação pode mudar.

As tecnologias podem ser substituídas.

Os modelos de linguagem podem evoluir.

Entretanto, a filosofia, os princípios e a arquitetura da Jay devem permanecer compreensíveis para qualquer pessoa ou agente de IA que participe do projeto, independentemente do momento em que ingressar em seu desenvolvimento.

---

# Diretriz Geral

Esta wiki documenta responsabilidades, princípios e decisões arquiteturais do projeto.

Ela não deve detalhar bibliotecas, frameworks ou tecnologias específicas, exceto quando essas escolhas já tiverem sido formalizadas por ADR.

A wiki do projeto não deve ser confundida com a futura wiki interna de conhecimento da própria Jay.
