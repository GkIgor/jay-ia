# ADR 0001: Regra de Responsabilidade Única (SOLID) e Limite de Complexidade

## Status
Aprovado

## Contexto
Durante a implementação de novos provedores de IA (como o OpenRouter), identificou-se uma tendência de funções de orquestração acumularem responsabilidade de conversão de dados, chamadas de rede e mapeamento de respostas. Métodos com muitas linhas e múltiplas responsabilidades dificultam a leitura, testes unitários isolados e violam o Princípio da Responsabilidade Única (SRP) do SOLID.

## Decisão
Fica estabelecido que:
1. Nenhuma função ou método nos repositórios do ecossistema Jay deve acumular múltiplas responsabilidades.
2. Camadas de transporte de dados ou comunicação externa (como chamadas HTTP) devem ter suas etapas de:
   - Tradução de entrada (Request payload mapping)
   - Execução de E/S (I/O, chamadas de rede)
   - Tradução de saída (Response payload mapping)
   isoladas em funções auxiliares pequenas, testáveis e focadas.
3. Comentários redundantes e óbvios que apenas narram o código ou dividem seções visuais ("bolo de comentários") devem ser evitados. O código deve ser autoexplicativo por meio de nomenclatura explícita de variáveis e funções de responsabilidade única.

## Consequências
- Aumento da testabilidade: funções de mapeamento podem ser testadas unitariamente de forma síncrona sem precisar mockar requisições HTTP reais.
- Facilidade de leitura e manutenção da base de código.
- Refatoração imediata da função `GenerateContent` em `openrouter.go`.
