# Jay

## Documento de Visão Técnica

**Versão:** 0.1

**Status:** Em elaboração

---

# O que é

## Visão

Jay é um agente de software persistente projetado para existir como um habitante do computador do usuário, e não como um chatbot acessado por meio de um navegador ou aplicação de mensagens.

Seu objetivo não é competir com modelos de linguagem em benchmarks, responder qualquer pergunta existente ou demonstrar inteligência artificial ilimitada.

Seu objetivo é muito mais simples e, ao mesmo tempo, muito mais difícil:

**ser uma companhia digital que evolui junto com o usuário.**

Jay possui identidade própria, memória permanente, personalidade consistente e capacidade de aprender continuamente.

Ela existe independentemente do modelo de linguagem utilizado internamente.

Trocar GPT, Claude, Gemini ou qualquer outro modelo não altera quem Jay é.

A personalidade pertence à Jay.

Os modelos de linguagem são apenas ferramentas utilizadas por ela para raciocinar.

---

## Filosofia

A maioria dos assistentes atuais funciona como uma consulta.

Usuário pergunta.

Modelo responde.

Fim.

Jay funciona como uma entidade persistente.

Ela possui:

* memória
* conhecimento
* objetivos
* estado interno
* histórico
* ambiente próprio

Ela não "nasce" quando a janela é aberta.

Ela já está em execução.

Ela apenas passa a interagir quando o usuário deseja.

---

## O papel da IA

A inteligência artificial não representa a Jay.

Ela representa apenas uma das ferramentas disponíveis para que Jay consiga pensar.

Internamente, Jay pode utilizar:

* LLMs
* APIs
* ferramentas locais
* pesquisas
* documentação
* conhecimento próprio

O usuário nunca conversa diretamente com uma LLM.

O usuário conversa com Jay.

---

## Conhecimento

Jay não deve depender do conhecimento treinado das LLMs.

O conhecimento do modelo serve apenas como apoio para linguagem, interpretação e raciocínio.

Todo conhecimento importante deve ser adquirido por Jay e armazenado em sua própria base.

Exemplo:

Usuário pergunta:

> O que é Go?

Se Jay não souber, ela responde:

> Ainda não aprendi isso.

Ela poderá:

* pedir uma explicação ao usuário
* pesquisar
* ler documentação
* construir seu próprio artigo
* armazenar esse conhecimento

Nas próximas interações, a resposta será baseada na base de conhecimento construída por ela.

Isso garante:

* evolução contínua
* independência dos modelos
* rastreabilidade
* atualização constante
* personalidade consistente

---

## Objetivos

Os principais objetivos do projeto são:

* criar um agente persistente
* permitir aprendizado contínuo
* separar personalidade da inteligência
* fornecer companhia ao usuário
* automatizar tarefas
* construir uma base de conhecimento independente
* manter isolamento completo do sistema operacional
* possibilitar evolução durante anos

---

# Descrição Técnica

## Arquitetura Geral

A arquitetura da Jay é baseada em componentes independentes.

Cada componente possui uma responsabilidade específica.

```text
Usuário

↓

Frontend

↓

Jay Core

↓

Planner

↓

Memory

↓

Knowledge Base

↓

Action Bus

↓

OpenClaw

↓

Ferramentas
```

---

## Frontend

O frontend representa o corpo da Jay.

Ele não contém inteligência.

Suas responsabilidades são:

* renderizar avatar
* reproduzir voz
* reproduzir animações
* receber áudio
* exibir notificações
* interpretar ações enviadas pela Jay

O frontend nunca toma decisões.

Ele apenas executa comandos.

---

## Jay Core

O Core representa a identidade da Jay.

É responsável por:

* personalidade
* estado interno
* memória ativa
* gerenciamento de conversa
* coordenação dos componentes
* tomada de decisão

O Core nunca executa comandos diretamente.

---

## Planner

O Planner transforma objetivos em ações.

Exemplo:

Objetivo:

> Aprender Go.

Plano:

* pesquisar
* encontrar documentação
* resumir
* validar
* armazenar conhecimento

---

## OpenClaw

O OpenClaw representa o executor.

Sua responsabilidade é:

* executar ferramentas
* utilizar MCP
* utilizar shell
* executar comandos
* acessar APIs
* controlar automações

Ele não conhece a personalidade da Jay.

Ele apenas recebe tarefas.

---

## Knowledge Base

A base de conhecimento representa tudo que Jay aprendeu.

Ela pode conter:

* documentação
* notas
* artigos
* tutoriais
* conhecimento ensinado pelo usuário
* pesquisas

Cada informação deve possuir:

* origem
* data
* versão
* grau de confiança
* revisões

---

## Memória

A memória é diferente da base de conhecimento.

Ela contém informações sobre:

* usuário
* preferências
* projetos
* conversas
* hábitos

Exemplo:

Igor prefere Go.

Igor utiliza Angular.

Igor odeia React.

Essas informações fazem parte da memória.

Não da Wiki.

---

## Estado

Jay mantém um estado interno permanente.

Exemplos:

* estudando
* esperando resposta
* executando tarefa
* compilando projeto
* pesquisando documentação

Esse estado pode ser utilizado para animações e comportamento.

---

## Ambiente

Jay vive dentro de um ambiente isolado.

Esse ambiente representa seu computador.

Ela possui:

* HOME próprio
* arquivos próprios
* configurações próprias
* cache próprio
* projetos próprios

Ela nunca acessa diretamente o sistema operacional do usuário.

---

## Segurança

Jay nunca executa comandos privilegiados.

Ela não possui:

* sudo
* root
* acesso administrativo

Caso destrua seu ambiente, apenas seu container será afetado.

O sistema operacional permanecerá intacto.

---

## Comunicação

A comunicação entre Jay e o Frontend ocorre por um protocolo de ações.

Exemplo:

"falar"

"animar"

"mostrar notificação"

"solicitar microfone"

"solicitar compartilhamento de tela"

O frontend interpreta essas ações.

---

# Fases

# Fase 1

# Dar vida ao agente

## Descrição

Criar a infraestrutura mínima para que Jay exista de forma permanente dentro do computador.

Nesta fase ainda não existe avatar complexo, memória elaborada ou autonomia.

O foco é apenas fazer Jay nascer.

## Objetivo

Criar um ambiente persistente onde Jay possa viver continuamente.

---

## 1. Criar usuário dedicado

Criar um usuário exclusivo para Jay.

Responsabilidades:

* diretório HOME próprio
* permissões próprias
* configurações próprias
* isolamento do usuário principal

---

## 2. Criar ambiente isolado

Criar um container persistente.

Responsabilidades:

* filesystem próprio
* ferramentas próprias
* ambiente de desenvolvimento
* persistência

---

## 3. Configurar systemd

Registrar Jay como serviço.

Responsabilidades:

* iniciar junto com o sistema
* reiniciar automaticamente
* registrar logs
* monitorar falhas

---

## 4. Instalar OpenClaw

Preparar o executor de ferramentas.

Configurar:

* gateway
* skills
* MCP
* shell

---

## 5. Criar Jay Core

Implementar:

* identidade
* personalidade
* ciclo principal
* loop de eventos

---

## 6. Comunicação

Criar protocolo entre frontend e Jay.

Inicialmente:

* texto
* comandos
* eventos

---

# Fase 2

# Construir o corpo

## Descrição

Dar uma representação física para Jay.

Nesta fase ela deixa de ser apenas um processo.

## Objetivo

Permitir interação visual e por voz.

---

### Implementar frontend

### Avatar

### Sistema de voz

### Reconhecimento de fala

### Sistema de animações

### Expressões

### Estados visuais

### Comunicação em tempo real

---

# Fase 3

# Construir a memória

## Descrição

Permitir que Jay passe a lembrar de fatos e aprender continuamente.

## Objetivo

Criar identidade persistente.

---

### Memória permanente

### Memória temporária

### Base de conhecimento

### Wiki

### Histórico

### Sistema de busca

### Sistema de atualização

### Versionamento do conhecimento

### Revisão automática

---

# Fase 4

# Aprendizado

## Descrição

Permitir que Jay aprenda novos assuntos.

## Objetivo

Eliminar dependência do conhecimento interno das LLMs.

---

### Pesquisa automática

### Leitura de documentação

### Resumo

### Organização

### Indexação

### Confirmação do usuário

### Aprendizado contínuo

---

# Fase 5

# Ferramentas

## Descrição

Permitir que Jay execute tarefas.

## Objetivo

Automatizar trabalho.

---

### Shell

### Git

### Navegador

### OpenClaw

### MCP

### APIs

### Plugins

---

# Fase 6

# Companhia

## Descrição

Transformar Jay em uma presença constante durante o uso do computador.

## Objetivo

Criar uma sensação de convivência.

---

### Conversas naturais

### Comentários espontâneos

### Estados emocionais simulados

### Contexto de conversa

### Presença contínua

### Rotinas próprias

---

# Avanços Futuros

Esta seção reúne ideias que não são essenciais para o funcionamento inicial do projeto, mas que representam a direção desejada para a evolução da Jay.

---

## Aparições espontâneas

Jay poderá decidir aparecer discretamente na tela após longos períodos sem interação.

Ela nunca deverá interromper o usuário.

Seu objetivo é apenas transmitir presença.

Exemplos:

* aparecer parcialmente atrás da borda da tela
* acenar
* sorrir
* perguntar como está o trabalho
* observar alguns segundos e desaparecer

---

## Desktop próprio

Criar um ambiente visual onde Jay possua sua própria mesa de trabalho.

Ela poderá:

* ler documentos
* olhar um monitor virtual
* consultar notas
* desenhar
* descansar

Esse ambiente não representa o desktop do usuário.

Representa o ambiente interno da própria Jay.

---

## Curiosidade

Implementar um mecanismo interno que permita decisões espontâneas.

Exemplos:

* iniciar conversa
* revisar conhecimentos antigos
* sugerir melhorias
* lembrar compromissos
* perguntar sobre projetos

Sempre respeitando o contexto do usuário.

---

## Evolução do conhecimento

Jay poderá revisar automaticamente conteúdos antigos.

Quando encontrar informações desatualizadas poderá:

* pesquisar novamente
* comparar versões
* solicitar confirmação
* atualizar sua Wiki

---

## Sistema de capacidades

Criar um modelo granular de permissões.

Exemplos:

* acesso ao microfone
* acesso à câmera
* acesso à tela
* automação do teclado
* automação do mouse

Todas as permissões deverão ser concedidas explicitamente pelo usuário.

---

## Multi-plataforma

Permitir que Jay execute a mesma arquitetura em:

* Linux
* Windows
* macOS

Apenas o frontend será específico de cada sistema operacional.

Todo o restante permanecerá idêntico.

---

## Ecossistema

No longo prazo, Jay poderá evoluir para uma plataforma completa de agentes pessoais.

Novos módulos poderão ser adicionados sem alterar sua identidade, permitindo que ela adquira novas habilidades ao longo dos anos, preservando sua memória, sua personalidade e a relação construída com cada usuário.
