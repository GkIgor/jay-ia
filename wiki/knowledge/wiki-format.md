# Formato da Wiki

## Escopo

Este diretório descreve como a documentação do projeto deve ser organizada.

Ele não define ainda o formato final da wiki interna de conhecimento da própria Jay.

## Distinção obrigatória

Existem dois conceitos diferentes:

### 1. Wiki do projeto

É a documentação deste repositório.

Ela registra:

- visão
- princípios
- arquitetura
- ADRs
- especificações
- fases
- PRDs

### 2. Base de conhecimento futura da Jay

É o sistema pelo qual a própria Jay poderá armazenar e consultar conhecimento operacional ao longo do tempo.

Ela poderá conter, por exemplo:

- conceitos aprendidos
- documentação curada
- fatos sobre ferramentas
- informações validadas sobre projetos

## Regra central

A wiki do projeto define como o sistema deve ser construído.

A base de conhecimento da Jay define o que a Jay sabe.

Esses dois espaços não devem ser misturados.

## Regras

- cada documento possui uma responsabilidade clara
- decisões formais devem virar ADR
- especificações não devem depender de tecnologia não decidida
- o índice deve permanecer atualizado
- textos devem priorizar princípios e contratos

## Organização esperada da wiki do projeto

Documentos devem responder, preferencialmente, a uma destas perguntas:

- o que a Jay deve ser?
- como a arquitetura está organizada?
- quais decisões foram tomadas?
- como uma fase será executada?
- que contrato técnico um componente precisa respeitar?

## O que evitar

- detalhes de implementação ainda não decididos
- duplicação entre páginas
- decisões implícitas escondidas em exemplos
- conteúdo temporário sem destino claro

## Linguagem

A documentação deve ser escrita em português, preservando termos técnicos em inglês quando fizer sentido.
