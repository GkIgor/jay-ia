# ADR 0002: Tipagem Forte e Estrutura Extensível para Balões de Chat em C++

## Status
Aprovado

## Contexto
O chat da aplicação é um painel de observabilidade e depuração da Jay. O fluxo inicial de mensagens no C++ utilizava formatos em string JSON desestruturados na renderização, o que gerava riscos de erros de digitação e acoplamento desnecessário à camada de IPC. O crescimento planejado do chat (logs de execução, blocos de raciocínio, toasts de agendamento) exige uma arquitetura extensível.

## Decisão
Fica acordado que:
1. O JSON bruto recebido do IPC será parseado e restrito unicamente à camada de `EventDispatcher`.
2. As mensagens no feed interno do frontend C++ serão representadas usando tipos fortes e enums (`ChatKind`) e payloads específicos por tipo em um variant tipado (`ChatPayload` via `std::variant`).
3. O cálculo de alturas e o despacho de desenho de cada tipo de balão são desacoplados do renderizador principal e centralizados em um componente de registro (`BubbleRegistry`), isolando o `ChatRenderer`.
4. Os estados de expansão de blocos recolhíveis serão controlados e indexados usando o `id` estável e imutável da mensagem, evitando que a inserção ou exclusão de novos balões interfira no layout.
5. Comentários redundantes que apenas descrevem o que o código C++ faz ou organizam seções visuais foram integralmente removidos para manter o código limpo e idiomático.

## Consequências
- Segurança de tipos em tempo de compilação na manipulação das mensagens ricas.
- Facilidade para adicionar novos tipos de balões (basta criar o renderer de desenho e registrá-lo no `BubbleRegistry`).
- Estabilidade visual durante o scroll e redimensionamento dinâmico de balões expansíveis.
