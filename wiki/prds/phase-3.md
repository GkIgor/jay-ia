# PRD Fase 3 - Execução de Ferramentas

## Objetivo

Permitir que a Jay execute ações concretas em seu ambiente através de um sistema de ferramentas desacoplado, seguro e extensível, iniciando a execução a partir de decisões do LLM e retornando os resultados para compor a resposta ao usuário.

---

# Escopo

Esta fase introduz a primeira infraestrutura completa de execução de ferramentas da Jay.

Ao final desta fase, a inteligência deixa de apenas gerar respostas e passa a interagir com o sistema operacional de forma controlada, sempre respeitando o fluxo de permissões definido pela aplicação.

---

# Requisitos Funcionais

## Arquitetura de Providers

Implementar a abstração definida na ADR 0008.

Todo provider deve implementar a interface comum `ToolProvider`, permitindo que novos mecanismos de execução sejam adicionados sem alterar a inteligência da Jay.

Nesta fase será implementado apenas o `NativeToolProvider`.

---

## Tool Bus

Implementar o Tool Bus como ponto único de comunicação entre a inteligência e os providers.

O Tool Bus deve ser responsável por:

- localizar a ferramenta solicitada;
- selecionar o provider responsável;
- encaminhar a execução;
- retornar o resultado padronizado.

O Tool Bus não deve conhecer detalhes internos de nenhum provider.

---

## Ferramentas iniciais

Disponibilizar as primeiras ferramentas nativas.

Escopo mínimo:

- leitura de arquivos;
- escrita de arquivos;
- execução de comandos;
- listagem de diretórios.

Todas as ferramentas devem possuir:

- identificação única;
- descrição;
- parâmetros definidos;
- validação de entrada;
- retorno padronizado.

---

## Fluxo de execução

Toda execução deverá seguir obrigatoriamente o fluxo abaixo:

```text
Usuário
    ↓
LLM
    ↓
Planner / Orquestrador
    ↓
Tool Bus
    ↓
Tool Provider
    ↓
Consentimento (quando necessário)
    ↓
Ferramenta
    ↓
Resultado
    ↓
LLM
    ↓
Resposta ao usuário
```

Não deve existir nenhuma forma da inteligência acessar diretamente o sistema operacional.

Toda interação obrigatoriamente passa pelo Tool Bus.

---

## Consentimento do usuário

Antes da execução de qualquer ferramenta classificada como sensível, a Jay deverá solicitar autorização explícita.

O modal de consentimento será implementado no cliente C++.

O usuário poderá:

- permitir;
- negar;
- cancelar.

Enquanto não houver resposta do usuário, a execução permanece bloqueada.

---

## Integração com o LLM

O LLM deverá ser capaz de:

- solicitar a execução de ferramentas;
- aguardar sua conclusão;
- receber o resultado estruturado;
- utilizar esse resultado para produzir a resposta final.

A infraestrutura somente será considerada concluída quando este ciclo estiver funcional de ponta a ponta.

---

# Requisitos Não Funcionais

- Inteligência e execução devem permanecer completamente desacopladas.
- Providers devem ser intercambiáveis.
- Ferramentas devem possuir contratos estáveis.
- Toda execução deve produzir sucesso ou erro padronizado.
- Falhas de uma ferramenta não devem comprometer o restante da aplicação.
- A arquitetura deve permitir a adição futura de novos providers (MCP, Remote Provider, Plugins etc.) sem alterações na inteligência.

---

# Fora de Escopo

Esta fase não contempla:

- MCP;
- Providers remotos;
- Marketplace de ferramentas;
- Execução paralela;
- Permissões persistentes;
- Instalação dinâmica de ferramentas;
- Agentes múltiplos.

---

# Critérios de Conclusão

A Fase 3 será considerada concluída somente quando todos os itens abaixo forem verdadeiros:

- A abstração `ToolProvider` estiver implementada.
- O `NativeToolProvider` estiver operacional.
- O Tool Bus estiver intermediando todas as execuções.
- As ferramentas básicas (arquivos e comandos) estiverem funcionando.
- O modal de consentimento estiver integrado ao cliente C++.
- O LLM conseguir solicitar a execução de uma ferramenta.
- A execução aguardar o consentimento do usuário quando necessário.
- O resultado retornar ao LLM após a execução.
- O LLM utilizar esse resultado para responder ao usuário.
- Não existir acesso direto da inteligência ao sistema operacional.
