# Estilo de Desenvolvimento

## Princípios

- clareza antes de esperteza
- desacoplamento antes de conveniência local
- contratos antes de integração improvisada
- documentação antes de decisão implícita

## Estilo arquitetural esperado

Como regra geral:

- o Core deve depender de interfaces, não de implementações concretas
- frontend deve ser cliente, não parte do Core
- providers devem ser substituíveis
- conhecimento, memória e execução não devem ser misturados

## Linguagem e documentação

## Regras atuais

- documentação em português
- nomes técnicos podem permanecer em inglês
- mudanças arquiteturais relevantes devem refletir a wiki
- detalhes tecnológicos não decididos não devem ser congelados na documentação

## Convenções de nomes

- arquivos Markdown em kebab-case
- ADRs numerados sequencialmente
- componentes técnicos podem manter nomes em inglês quando forem mais naturais no domínio

## Regra para exemplos

Exemplos devem ilustrar contrato e responsabilidade, não introduzir decisões escondidas que contradigam a arquitetura documentada.
