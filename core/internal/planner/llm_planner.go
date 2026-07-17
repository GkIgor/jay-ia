package planner

import (
	"context"
	"strings"

	"github.com/GkIgor/jay-ia/core/internal/llm"
	"github.com/GkIgor/jay-ia/core/internal/tools"
)

// LLMPlanner atua como orquestrador que coordena o fluxo de consulta à LLM agnóstica
type LLMPlanner struct {
	llmClient     llm.Client
	toolBus       *tools.ToolBus
	simplePlanner *SimplePlanner
}

func NewLLMPlanner(client llm.Client, toolBus *tools.ToolBus) *LLMPlanner {
	return &LLMPlanner{
		llmClient:     client,
		toolBus:       toolBus,
		simplePlanner: NewSimplePlanner(),
	}
}

func (p *LLMPlanner) Plan(ctx context.Context, input string, planCtx PlanningContext) (*Plan, error) {
	trimmed := strings.TrimSpace(input)

	// Se o comando for "/chat <prompt>", limpamos o prefixo e enviamos para a LLM
	if strings.HasPrefix(trimmed, "/chat ") {
		input = strings.TrimPrefix(trimmed, "/chat ")
	} else if strings.HasPrefix(trimmed, "/") {
		// Se começa com outra barra (ex: /ping, /write), é um comando CLI direto
		return p.simplePlanner.Plan(ctx, input, planCtx)
	}

	// Recupera definições de ferramentas disponíveis catalogadas no barramento
	availableTools := p.toolBus.ListAvailableTools()

	// Invoca o adaptador da LLM de forma agnóstica
	llmResp, err := p.llmClient.GenerateContent(ctx, planCtx.History, availableTools)
	if err != nil {
		return nil, err
	}

	// Transforma a resposta agnóstica no Plan final via Plan Builder
	return BuildPlan(llmResp), nil
}

// BuildPlan transforma o retorno agnóstico da LLM em um plano operacional de passos
func BuildPlan(resp *llm.Response) *Plan {
	if resp == nil {
		return &Plan{}
	}

	var steps []Step
	// Se a LLM solicitou chamada de ferramenta
	for _, fc := range resp.FunctionCalls {
		steps = append(steps, Step{
			Type: StepToolExecute,
			Params: map[string]any{
				"tool": fc.Name,
				"args": fc.Args,
			},
		})
	}

	// Se a LLM retornou texto final
	if resp.Text != "" {
		steps = append(steps, Step{
			Type: StepRespond,
			Params: map[string]any{
				"text": resp.Text,
			},
		})
	}

	return &Plan{Steps: steps}
}
