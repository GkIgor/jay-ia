# ADR 0003: Abstração Pura do Provedor de LLM com IDs e Metadados Agnósticos

## Status
Aprovado

## Contexto
Durante o desenvolvimento do provedor OpenRouter, a dependência temporária de structs específicas de SDK (como `*genai.Part` ou estruturas do OpenRouter) sob o campo genérico `RawSDKPart` quebrou a pureza e o desacoplamento do pacote `llm`. Além disso, a associação entre chamadas de ferramentas e suas respostas correspondentes no OpenRouter dependia de heurísticas de busca retrógrada no histórico em vez de associações explícitas.

## Decisão
Fica estabelecido que:
1. O campo `RawSDKPart` de tipo `any` será permanentemente removido das estruturas públicas `llm.Part` e `llm.FunctionCall`.
2. Provedores que precisam persistir estruturas internas específicas (como o `thought_signature` do Gemini) devem serializar esse dado como string JSON ou mapear para pares de strings genéricas e salvá-los no campo agnóstico `Metadata map[string]string` de `llm.Part` ou `llm.FunctionCall`. O provedor é responsável por desserializar esse dado de volta para o tipo nativo do seu SDK.
3. Adiciona-se o identificador de chamada (`ID string`) diretamente às estruturas `llm.FunctionCall` e `llm.FunctionResponse`. Esse ID é gerado pelo provedor de LLM e devolvido na resposta de tool call, sendo posteriormente reenviado associado à resposta da ferramenta de forma consistente, eliminando heurísticas retroativas no histórico.
4. Conexões de rede HTTP feitas por provedores de LLM personalizados (como OpenRouter) devem configurar e utilizar um `http.Client` com timeout de rede explícito de 30 segundos para evitar bloqueios indefinidos da CPU ou vazamentos de recursos.
5. Políticas de modelos padrão ou fallbacks específicos (como associar o modelo de chat do OpenRouter ao Gemini caso o modelo não seja fornecido) devem ser tratadas pelo `router.go` ou camada de configuração e injetadas no cliente, mantendo a camada do provedor pura e agnóstica às políticas de negócio.

## Consequências
- Desacoplamento total de bibliotecas de terceiros das estruturas internas de conversação do Core.
- Robustez contra conexões HTTP travadas.
- Rastreamento confiável de tool calls em múltiplos turnos de conversa.
