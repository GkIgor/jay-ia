# Ambiente

## Estado atual

O repositório ainda está em fase de definição arquitetural e documental.

## Pré-requisito atual

Para o desenvolvimento do `core`, `cli` e `sdk`, a linguagem base escolhida foi Go (ver ADR 0005). O colaborador precisará ter o compilador do Go (1.21+) instalado no seu ambiente.

Antes de preparar qualquer ambiente de desenvolvimento, o colaborador precisa entender a arquitetura pretendida.

## Leitura obrigatória

Antes de criar setup definitivo, qualquer colaborador deve consultar:

- `wiki/README.md`
- `wiki/index.md`
- `wiki/vision/`
- arquitetura e ADRs relevantes

## Fluxo de início recomendado

1. Ler a wiki base.
2. Identificar a fase e o PRD relacionados ao trabalho.
3. Confirmar se já existe ADR cobrindo a decisão necessária.
4. Só então iniciar estrutura, código ou automação.

## Regra operacional

Se uma tarefa exigir uma decisão ainda não formalizada, o trabalho deve parar no ponto de ambiguidade, solicitar orientação humana e registrar o resultado na wiki antes de seguir.

## Observação

Quando o ambiente real de desenvolvimento for formalizado, este documento deverá ser atualizado com comandos e pré-requisitos concretos.
