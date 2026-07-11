# OpenClaw

## Papel no sistema

OpenClaw é tratado como executor de ferramentas.

Ele não representa:

- identidade da Jay
- memória
- conhecimento
- planejamento central

## Uso esperado

OpenClaw pode ser um provider do barramento de ferramentas da Jay para:

- shell
- skills
- MCP
- automação
- integrações externas

## Regra de desacoplamento

Jay não deve depender estruturalmente de OpenClaw.

O sistema deve permitir a troca futura por outros executores sem reescrever o Core.
