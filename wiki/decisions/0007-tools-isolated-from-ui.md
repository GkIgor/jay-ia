# ADR 0007: Isolamento entre Ferramentas e Interface (Tool-User Isolation)

## Status

Aceita

## Contexto

Na Fase 3 (Ações), as ferramentas integradas ao Core precisam de validações, interações ou autorizações pontuais por parte do usuário (como consentimento de segurança para ler ou escrever arquivos restritos). 

Se permitirmos que as ferramentas individuais (ex: `fs.write_file`) gerenciem a comunicação com o usuário, decodifiquem protocolos IPC ou acessem rotinas de interface gráfica, introduziremos sérios problemas arquiteturais:
1. **Acoplamento**: As ferramentas deixariam de ser autocontidas e dependeriam de detalhes visuais ou de rede.
2. **Inconsistência**: Cada ferramenta precisaria reimplementar ou conhecer rotinas de consentimento humano e controle de erros.
3. **Falta de controle central**: O Core (Daemon) perderia a capacidade de auditar, gerenciar políticas globais ou barrar chamadas antes que ocorram.

## Decisão

**A ferramenta (Tool) nunca conversa com o usuário.**

- O fluxo de consentimento ou exibição de dados deve obrigatoriamente respeitar a hierarquia:
  ```text
  Tool (Requer permissão na definição)
     ↓
  Daemon (Intercepta e gerencia consentimento)
     ↓
  InternalBus (Dispara evento conceitual)
     ↓
  IPC Server (Transmite o JSON conceitual de request)
     ↓
  Frontend (Desenha o modal e coleta a resposta do Usuário)
  ```
- Nenhuma ferramenta conterá rotinas que solicitem ativamente dados, desenhem caixas de diálogo ou bloqueiem threads de forma independente do Daemon. Elas apenas declaram em seus metadados (`Describe()`) quais permissões exigem, delegando a responsabilidade de obter essa validação ao `Daemon`.

## Consequências

- **Vantagens**:
  - **Autocontenção**: As ferramentas mantêm-se como funções puras de sistema operacional, simplificando testes de unidade isolados.
  - **Segurança Centralizada**: O `Daemon` é o único portão de segurança (gatekeeper), garantindo auditoria de logs e validações padronizadas de consentimento antes de chamar qualquer execução.
  - **Flexibilidade**: O frontend pode renderizar o pedido de permissão de diferentes formas (modal gráfico, prompt de voz, confirmação via CLI) sem requerer alterações em uma única linha do código das ferramentas.
