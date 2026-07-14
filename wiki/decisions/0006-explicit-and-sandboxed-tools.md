# ADR 0006: Ferramentas Explícitas e Autocontidas

## Status

Aceita

## Contexto

A Fase 3 (Ações) introduz a capacidade da Jay interagir com o ambiente e executar tarefas no sistema operacional hospedeiro. Em ambientes de agentes comuns, é frequente expor um interpretador de terminal arbitrário (como um wrapper de bash/shell genérico) para que a LLM execute comandos. 

No entanto, essa abordagem traz riscos graves de segurança a longo prazo:
1. **Falta de controle de permissões**: Impossível auditar ou barrar ações maliciosas granulares se todas forem empacotadas no mesmo processo bash.
2. **Fragilidade do Planner**: O modelo de linguagem gera código arbitrário que pode quebrar dependendo de sutilezas do ambiente.
3. **Falta de isolamento semântico**: O Core perde a introspecção das capacidades reais do agente.

## Decisão

As ferramentas da Jay serão estritamente **explícitas e autocontidas**.

- O Core **não** oferecerá um terminal arbitrário (como um shell genérico executando comandos aleatórios de texto livre) como mecanismo primário de execução.
- Toda ação do sistema (ex: leitura de arquivos, escrita, listagem, diff, requisições de rede) deve ser de grão fino e modelada como uma ferramenta independente e autodescritiva (por exemplo, `fs.read_file`, `git.status`), que expõe assinaturas e regras de validação explícitas.
- Essas ferramentas rodarão sob controle de um sandbox e serão expostas de forma granular através de providers no `ToolBus`.

## Consequências

- **Vantagens**: 
  - Segurança granular: Conseguimos injetar verificações de permissões e controle humano (human-in-the-loop) para ações específicas (ex: permitir `fs.read_file` por padrão, mas exigir consentimento expresso em `fs.write_file`).
  - Robusteza: Menos erros de parser por parte da inteligência, já que as ferramentas expõem parâmetros estruturados claros.
  - Audibilidade: Logs precisos de cada ação executada pelo agente, sem a opacidade de sessões bash.
- **Limitações**: Adicionar novas capacidades requer criar e registrar estruturas de ferramentas específicas em Go, porém isso é facilitado por um barramento extensível baseado em provedores.
