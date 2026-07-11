# Jay

## Documento de Visão Técnica

**Versão:** 0.2

**Status:** Em elaboração

---

# O que é

Jay é um agente de software persistente projetado para existir como um habitante do computador do usuário, e não como um chatbot acessado apenas por navegador, mensageiro ou aba temporária.

Seu objetivo não é competir em benchmarks nem fingir conhecimento ilimitado.

Seu objetivo é:

**ser uma companhia digital que evolui junto com o usuário.**

Jay possui:

- identidade própria
- memória permanente
- personalidade consistente
- capacidade de aprender continuamente
- ambiente próprio

Ela continua sendo Jay independentemente do modelo de linguagem utilizado internamente.

Trocar GPT, Claude, Gemini, Ollama ou qualquer outro provider não altera sua identidade.

---

# Filosofia

Assistentes comuns funcionam como consulta:

```text
usuário pergunta
↓
modelo responde
↓
fim
```

Jay funciona como entidade persistente:

```text
Jay já está viva
↓
mantém estado, memória e objetivos
↓
interage quando o usuário deseja
↓
continua existindo depois da conversa
```

Ela não nasce quando a janela é aberta.

Ela já está em execução.

---

# Papel da LLM

A LLM não representa a Jay.

Ela é apenas um componente do sistema, com dois papéis principais:

- comunicação
- raciocínio

A LLM pode ajudar a interpretar, planejar e conversar.

Ela **não** é a fonte oficial de conhecimento da Jay para domínios administrados por ela.

---

# Conhecimento

Jay não deve depender do conhecimento treinado da LLM como autoridade para assuntos que ela administra.

Isso inclui, por exemplo:

- programação
- Linux
- projetos do usuário
- documentação
- workspace
- preferências do usuário

Quando não souber algo, o comportamento esperado é:

```text
Jay sabe?
↓
não
↓
admite que ainda não sabe
↓
pergunta ao usuário ou pesquisa
↓
cura o conteúdo
↓
registra
↓
reutiliza no futuro
```

Exemplo:

> "O que é Go?"

Se Jay ainda não souber, ela deve responder algo como:

> "Ainda não aprendi isso. Você quer me explicar ou prefere que eu pesquise?"

Essa abordagem garante:

- evolução contínua
- rastreabilidade
- atualização progressiva
- independência parcial dos modelos
- personalidade mais consistente

---

# Objetivos do Projeto

- criar um agente persistente
- separar personalidade, memória, conhecimento e execução
- permitir aprendizado contínuo
- fornecer companhia digital sem se tornar intrusiva
- automatizar tarefas dentro de limites claros
- operar em ambiente isolado
- sustentar evolução por muitos anos

---

# Fonte de Verdade

A wiki deste repositório é a fonte oficial de verdade do projeto.

Ela documenta:

- princípios
- responsabilidades
- decisões arquiteturais
- especificações
- planejamento

Se algo importante não estiver documentado na wiki, o agente deve pedir orientação a um humano e registrar a decisão no lugar correto.

## Importante

A wiki do projeto não deve ser confundida com a futura base interna de conhecimento da própria Jay.

Esta wiki descreve o projeto.

A wiki interna da Jay continua em aberto como tema de arquitetura e evolução.

---

# Arquitetura Geral

Jay é organizada como um conjunto de componentes independentes com responsabilidades claras.

```text
Usuário
   ↓
Frontend
   ↓
Jay Core
   ├── Personality
   ├── Conversation
   ├── Planner
   ├── Memory
   ├── Knowledge
   ├── Learning
   ├── LLM Router
   └── Tool Bus
         ├── OpenClaw Provider
         ├── MCP Provider
         ├── Native Provider
         └── HTTP/API Provider
```

## Frontend

O frontend representa o corpo da Jay.

Responsabilidades:

- avatar
- voz
- animações
- renderização
- microfone
- notificações
- interpretação das ações enviadas pelo Core

O frontend não pensa.

Ele executa.

## Core

O Core representa a identidade operacional da Jay.

Responsabilidades:

- personalidade
- estado interno
- coordenação da conversa
- planejamento
- memória
- conhecimento
- aprendizado
- orquestração de providers

O Core deve ser completamente independente de qualquer frontend.

## Planner

O Planner transforma intenção em plano de ação.

Exemplo:

```text
objetivo: aprender Go
↓
pesquisar
↓
encontrar fontes
↓
curar
↓
registrar
↓
reutilizar
```

## OpenClaw

OpenClaw é executor de ferramentas.

Ele pode ser usado para shell, skills, MCP, automação e integrações.

Ele não representa:

- personalidade
- memória
- conhecimento
- cérebro da Jay

## Memória

Memória registra continuidade pessoal e operacional.

Exemplos:

- preferências do usuário
- hábitos
- projetos recorrentes
- relações e contexto durável

## Conhecimento

Conhecimento representa aquilo que a Jay considera verdade nos domínios que ela administra.

Cada item idealmente deve permitir:

- origem
- data
- confiança
- revisão futura

## Estado

Estado é transitório.

Exemplos:

- compilando
- pesquisando
- esperando resposta
- executando tarefa

---

# Isolamento

Jay deve viver em um ambiente isolado e persistente.

Esse ambiente é sua casa operacional.

Requisitos arquiteturais já decididos:

- usuário dedicado
- ausência de privilégios administrativos
- filesystem próprio
- persistência de dados
- isolamento do host

## O que ainda está em aberto

A tecnologia concreta de isolamento ainda não foi escolhida.

O projeto ainda não fixa Podman, LXC/LXD, systemd-nspawn ou outra alternativa.

A arquitetura não deve depender dessa escolha.

---

# Segurança

Jay não deve operar com acesso administrativo como comportamento normal.

Ela não deve depender de:

- `sudo`
- `root`
- acesso irrestrito ao host

Recursos sensíveis como tela, microfone e webcam devem ser acessados por mediação do frontend e por mecanismos explícitos de permissão.

O objetivo é limitar danos, preservar previsibilidade e manter fronteiras claras entre agente e sistema operacional.

---

# Comunicação entre Core e Frontend

Jay não deve desenhar interface diretamente.

Ela emite intenções por protocolo.

Exemplos conceituais:

- `avatar.wave`
- `avatar.smile`
- `speech.start`
- `notify`
- `request.microphone`

O frontend interpreta essas ações e decide como renderizá-las na plataforma local.

Essa separação permite:

- Core headless
- reinício isolado do frontend
- múltiplos frontends
- portabilidade futura

---

# Fases do Projeto

## Fase 1: Vida

Estabelece a existência persistente da Jay como processo contínuo.

Foco:

- daemon
- identidade básica
- memória mínima
- operação headless

## Fase 2: Corpo

Dá presença visual e vocal à Jay sem acoplar o Core à interface.

Foco:

- frontend inicial
- avatar
- voz
- animações básicas

## Fase 3: Ações

Permite que a Jay utilize ferramentas e aja em seu ambiente.

Foco:

- Tool Bus
- OpenClaw como executor
- shell, arquivos, automação

## Fase 4: Autonomia Assistida

Introduz iniciativa limitada, útil e não intrusiva.

Foco:

- lembretes
- sugestões
- agenda
- ciclos próprios de verificação

## Fase 5: Aprendizado Contínuo

Consolida a capacidade de aprender, registrar, revisar e reutilizar conhecimento.

Foco:

- aquisição de conhecimento
- curadoria
- confiança
- revisão

## Fase 6: Companhia Digital

Aprofunda presença, continuidade e convivência com o usuário.

Foco:

- coerência de personalidade
- presença persistente
- sensação de companhia
- múltiplos clientes e superfícies

Partes desta fase ainda representam direção atual, não decisão final.

---

# Direção de Longo Prazo

No longo prazo, Jay deve evoluir como plataforma de agente pessoal persistente.

Isso inclui:

- novos frontends
- novos providers
- maior capacidade de aprendizado
- presença mais refinada
- evolução sem perda de identidade

O objetivo não é construir mais um chat com voz.

O objetivo é construir um habitante digital do computador.
