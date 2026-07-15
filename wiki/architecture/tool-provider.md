# Tool Provider

## Visão Geral

A Jay possui a capacidade de executar ações no sistema operacional, acessar serviços, consultar modelos de linguagem e interagir com diferentes recursos externos.

Entretanto, essas capacidades não pertencem ao Core.
O Core apenas toma decisões. A execução dessas decisões é responsabilidade de um **Tool Provider**.

O objetivo desta camada é desacoplar a lógica da Jay da tecnologia utilizada para executar suas ações.

## Filosofia

A Jay é o agente. O Tool Provider é apenas um executor.
A inteligência, personalidade, memória e planejamento pertencem exclusivamente ao Core.
O Provider apenas transforma uma intenção em uma ação concreta.

## Arquitetura

O Core nunca interage diretamente com o sistema operacional. Toda comunicação passa pelo Tool Provider.

```text
                Jay Core
                    │
            Planejamento
                    │
                    ▼
             Tool Provider
                    │
      ┌─────────────┼─────────────┐
      │             │             │
    Shell         HTTP          Git
      │             │             │
      └─────────────┼─────────────┘
                    │
              Sistema Operacional
```

## Native Provider

A implementação inicial e padrão desta interface é chamada **Native Provider**.

Seu objetivo é fornecer apenas as capacidades necessárias para que o projeto possa nascer e evoluir de forma incremental na Fase 1. Inicialmente ele pode executar:
- Comandos locais
- Leitura e escrita de arquivos
- Operações HTTP e chamadas a modelos de linguagem
- Operações básicas com Git
- Abertura de aplicações externas

> **Nota Arquitetural**: O Native Provider não representa a implementação definitiva da camada de execução. Ele existe para validar a arquitetura do Core e reduzir a complexidade da Fase 1. A adoção de um executor mais sofisticado (OpenClaw, MCP, Remote, Cloud) permanece uma possibilidade futura sólida e transparente, visto que a troca de implementação não deverá exigir alterações na lógica da Jay.
