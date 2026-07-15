# ADR 0008: Adoção de um Native Tool Provider para a Fase 1

## Status

Aceita

## Contexto

A Jay possui a capacidade de executar ações no sistema operacional, acessar serviços, consultar modelos de linguagem e interagir com diferentes recursos externos. Originalmente, o OpenClaw (ADR 0003) foi selecionado para ser o motor principal dessas execuções.

Contudo, utilizar uma plataforma externa de execução desde o primeiro dia introduz complexidade excessiva para a Fase 1 (Gateway, ciclo de vida de Providers, gerenciamento de Skills e comunicação adicional), o que atrasa o nascimento e a validação do próprio Core.

## Decisão

Institui-se uma camada de abstração genérica chamada **Tool Provider**.
O Core apenas toma decisões; a execução concreta das ferramentas é sempre repassada a um Provider. 

Para a Fase 1, implementaremos o **Native Provider**, a implementação padrão da Jay embutida no próprio Core, focada apenas nas capacidades fundamentais:
* Comandos locais shell
* Leitura/escrita de arquivos
* Operações HTTP e chamadas a LLMs
* Operações básicas com Git e abertura de aplicações

## Consequências

- **Positivo**: O Core pode nascer, ser testado e validado imediatamente sem depender estruturalmente de projetos externos. Baixo acoplamento e redução dramática de complexidade inicial.
- **Positivo**: A arquitetura através de uma interface genérica de Tool Provider garante que a troca futura por executores mais sofisticados (MCP Provider, OpenClaw Provider, Cloud Provider) ocorra sem afetar o raciocínio da Jay.
- **Negativo**: Algum esforço inicial para recriar funcionalidades básicas (shell, arquivos) de forma segura em vez de já ganhar as proteções completas do OpenClaw.
