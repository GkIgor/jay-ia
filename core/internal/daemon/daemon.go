package daemon

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/GkIgor/jay-ia/core/internal/bus"
	"github.com/GkIgor/jay-ia/core/internal/conversation"
	"github.com/GkIgor/jay-ia/core/internal/ipc"
	"github.com/GkIgor/jay-ia/core/internal/llm"
	"github.com/GkIgor/jay-ia/core/internal/memory"
	"github.com/GkIgor/jay-ia/core/internal/planner"
	"github.com/GkIgor/jay-ia/core/internal/state"
	"github.com/GkIgor/jay-ia/core/internal/tools"
	sdkipc "github.com/GkIgor/jay-ia/sdk/ipc"
)

// PermissionTimeout define o tempo limite de aprovação da interface
const PermissionTimeout = 30 * time.Second

// MaxPlanningIterations define o limite de segurança contra loops infinitos de IA
const MaxPlanningIterations = 5

// Daemon coordinates all core components
type Daemon struct {
	currentState   state.State
	memoryStore    memory.MemoryStore
	ipcServer      *ipc.Server
	planner        planner.Planner
	bus            *bus.InternalBus
	toolBus        *tools.ToolBus
	convManager    *conversation.Manager
	pendingPermsMu sync.Mutex
	pendingPerms   map[string]chan bool
	nextRequestID  uint64
}

// New creates a new Jay Daemon
func New() (*Daemon, error) {
	b := bus.NewInternalBus()
	tb := tools.NewToolBus()

	// Registra provedor nativo e ferramentas explíticas de arquivos
	np := tools.NewNativeProvider()
	np.RegisterTool(tools.ReadFileTool{})
	np.RegisterTool(tools.WriteFileTool{})
	np.RegisterTool(tools.ListDirTool{})
	tb.RegisterProvider(np)

	apiKey := os.Getenv("GEMINI_API_KEY")
	var llmClient llm.Client
	var err error
	if apiKey != "" {
		llmClient, err = llm.NewClient(llm.Config{
			Provider: "gemini",
			APIKey:   apiKey,
		})
	} else {
		log.Println("WARNING: GEMINI_API_KEY not found in environment. Initializing Daemon with 'mock' LLM Client.")
		llmClient, err = llm.NewClient(llm.Config{
			Provider: "mock",
		})
	}
	if err != nil {
		return nil, fmt.Errorf("failed to initialize LLM client: %w", err)
	}

	d := &Daemon{
		currentState: state.Idle,
		memoryStore:  memory.NewInMemoryStore(),
		planner:      planner.NewLLMPlanner(llmClient, tb),
		bus:          b,
		toolBus:      tb,
		convManager:  conversation.NewManager(),
		pendingPerms: make(map[string]chan bool),
	}

	ipcServer, err := ipc.NewServer(d.handleIPCMessage)
	if err != nil {
		return nil, err
	}
	d.ipcServer = ipcServer

	// Subscribe IPC Server to the InternalBus
	ch := d.bus.Subscribe(100)
	go func() {
		for ev := range ch {
			switch e := ev.(type) {
			case bus.StateChangedEvent:
				d.ipcServer.Broadcast(ipc.IPCEvent{
					Type:    e.EventName(),
					Payload: map[string]string{"state": strings.ToLower(e.NewState)},
				})
			case bus.AnimationPlayEvent:
				d.ipcServer.Broadcast(ipc.IPCEvent{
					Type:    e.EventName(),
					Payload: map[string]string{"animation": e.Animation},
				})
			case bus.ToolProgressEvent:
				d.ipcServer.Broadcast(ipc.IPCEvent{
					Type: e.EventName(),
					Payload: map[string]any{
						"tool":    e.ToolName,
						"state":   e.State,
						"percent": e.Percent,
						"message": e.Message,
					},
				})
			case bus.ToolCompletedEvent:
				d.ipcServer.Broadcast(ipc.IPCEvent{
					Type: e.EventName(),
					Payload: map[string]any{
						"tool":    e.ToolName,
						"success": e.Success,
						"output":  e.Output,
						"error":   e.Error,
					},
				})
			case bus.PermissionRequestedEvent:
				d.ipcServer.Broadcast(ipc.IPCEvent{
					Type: "request.permission",
					Payload: map[string]any{
						"ref_id":     e.RequestID,
						"permission": e.Permission,
						"prompt":     e.Prompt,
					},
				})
			}
		}
	}()

	return d, nil
}

// Start starts the daemon and blocks until stopped
func (d *Daemon) Start() error {
	log.Printf("Starting Jay Daemon... Current State: %s", d.currentState)

	if err := d.ipcServer.Start(); err != nil {
		return err
	}

	d.setState(state.Idle)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	<-sigChan
	log.Println("Termination signal received. Shutting down...")

	d.Stop()
	return nil
}

// Stop safely shuts down the daemon
func (d *Daemon) Stop() {
	if d.ipcServer != nil {
		d.ipcServer.Stop()
	}
	log.Println("Jay Daemon stopped.")
}

// setState changes the internal state and publishes the event to the bus
func (d *Daemon) setState(s state.State) {
	d.currentState = s
	d.bus.Publish(bus.StateChangedEvent{NewState: s.String()})
}

// RequestPermission publica a solicitação no bus e bloqueia aguardando a resposta do Frontend
func (d *Daemon) RequestPermission(ctx context.Context, toolName, permission string) (bool, error) {
	d.pendingPermsMu.Lock()
	reqID := fmt.Sprintf("req_%d", atomic.AddUint64(&d.nextRequestID, 1))
	ch := make(chan bool, 1)
	d.pendingPerms[reqID] = ch
	d.pendingPermsMu.Unlock()

	prompt := fmt.Sprintf("Jay quer executar a ferramenta '%s' que exige a permissão '%s'. Permitir?", toolName, permission)

	// Publica a intenção conceitualmente no barramento
	d.bus.Publish(bus.PermissionRequestedEvent{
		RequestID:  reqID,
		Permission: permission,
		Prompt:     prompt,
	})

	log.Printf("Permission '%s' (ref: %s) requested for tool '%s'. Waiting approval...", permission, reqID, toolName)

	var allowed bool
	select {
	case allowed = <-ch:
	case <-time.After(PermissionTimeout):
		log.Printf("Permission '%s' (ref: %s) timed out after %v", permission, reqID, PermissionTimeout)
		allowed = false
	}

	d.pendingPermsMu.Lock()
	delete(d.pendingPerms, reqID)
	d.pendingPermsMu.Unlock()

	return allowed, nil
}

// handleIPCMessage acts as the main orchestrator for commands received from IPC.
func (d *Daemon) handleIPCMessage(msg sdkipc.Message) sdkipc.Message {
	// Trata a resposta de consentimento do usuário vinda da interface
	if msg.Type == "permission.response" {
		payloadBytes, err := json.Marshal(msg.Payload)
		if err != nil {
			return d.errorResponse("internal_error", "failed to process permission response")
		}
		var resp struct {
			RefID    string `json:"ref_id"`
			Allowed  bool   `json:"allowed"`
			Modality string `json:"modality"`
		}
		if err := json.Unmarshal(payloadBytes, &resp); err == nil {
			log.Printf("IPC Permission Response received. RefID: %s, Allowed: %v, Modality: %s", resp.RefID, resp.Allowed, resp.Modality)
			d.pendingPermsMu.Lock()
			ch, exists := d.pendingPerms[resp.RefID]
			d.pendingPermsMu.Unlock()
			if exists {
				select {
				case ch <- resp.Allowed:
				default:
				}
			}
		}
		return sdkipc.Message{
			Type: "response",
			Payload: sdkipc.Response{
				Status: "ok",
				Data:   "permission response processed",
			},
		}
	}

	if msg.Type != "command" {
		return sdkipc.Message{
			Type: "error",
			Payload: sdkipc.Error{
				Code:    400,
				Message: fmt.Sprintf("invalid message type: %s", msg.Type),
			},
		}
	}

	payloadBytes, err := json.Marshal(msg.Payload)
	if err != nil {
		return d.errorResponse("internal_error", "failed to process command payload")
	}

	var cmd sdkipc.Command
	if err := json.Unmarshal(payloadBytes, &cmd); err != nil {
		return d.errorResponse("invalid_command", "invalid command structure")
	}

	// 1. Resolve/Prepare PlanningContext (Perception phase)
	d.setState(state.Thinking)
	time.Sleep(2 * time.Second)

	input := ""
	if cmd.Action != "" {
		input = "/" + cmd.Action
		if cmd.Data != nil {
			if s, ok := cmd.Data.(string); ok && s != "" {
				input += " " + strings.TrimSpace(s)
			}
		}
	} else if cmd.Data != nil {
		if s, ok := cmd.Data.(string); ok {
			input = strings.TrimSpace(s)
		}
	}
	input = strings.TrimSpace(input)

	// Adiciona a entrada do usuário à sessão do Conversation Manager
	d.convManager.AddUserMessage(input)

	var responseText string
	iteration := 0

	for iteration < MaxPlanningIterations {
		iteration++

		planCtx := planner.PlanningContext{
			WorkingMemory: make(map[string]string),
			History:       d.convManager.GetHistory(),
		}

		if cmd.Action == "recall" || strings.HasPrefix(input, "/recall") {
			var key string
			if strings.HasPrefix(input, "/recall") {
				parts := strings.SplitN(input, " ", 2)
				if len(parts) >= 2 {
					key = strings.TrimSpace(parts[1])
				}
			} else {
				key = input
			}

			if key != "" {
				val, err := d.memoryStore.Get(key)
				if err == nil {
					if valStr, ok := val.(string); ok {
						planCtx.WorkingMemory[key] = valStr
					}
				}
			}
		}

		// 2. Call Planner
		plan, err := d.planner.Plan(context.Background(), input, planCtx)
		if err != nil {
			d.setState(state.Idle)
			return d.errorResponse(cmd.ID, fmt.Sprintf("planning error: %v", err))
		}

		hasToolExecution := false
		var currentToolName string
		var currentArgs map[string]any
		var currentRawSDKPart any

		for _, step := range plan.Steps {
			if step.Type == planner.StepToolExecute {
				hasToolExecution = true
				if t, ok := step.Params["tool"].(string); ok {
					currentToolName = t
				}
				if a, ok := step.Params["args"].(map[string]any); ok {
					currentArgs = a
				} else if m, ok := step.Params["args"].(map[string]interface{}); ok {
					currentArgs = m
				}
				// Preserva o Part original do SDK (com thought_signature) para reenvio correto ao Gemini
				currentRawSDKPart = step.Params["raw_sdk_part"]
				break
			}
		}

		// Se o plano não solicita ferramentas, apenas retorna a resposta textual final
		if !hasToolExecution {
			for _, step := range plan.Steps {
				switch step.Type {
				case planner.StepRespond:
					if text, ok := step.Params["text"].(string); ok {
						responseText = text
						d.convManager.AddModelMessage(text)
					}
				case planner.StepMemoryPut:
					key, kOk := step.Params["key"].(string)
					val, vOk := step.Params["value"].(string)
					if kOk && vOk {
						if err := d.memoryStore.Put(key, val); err != nil {
							log.Printf("Failed to write to memory: %v", err)
						}
					}
				case planner.StepHumanEscalate:
					responseText = "Escalated to human."
					d.convManager.AddModelMessage(responseText)
				}
			}
			break
		}

		// 3. Execute Steps (Side-effects execution phase)
		d.setState(state.Executing)
		time.Sleep(1 * time.Second)

		// Obtém metadados da ferramenta para checar permissões
		tool, exists := d.toolBus.GetTool(currentToolName)
		if !exists {
			errStr := fmt.Sprintf("Tool %s not found", currentToolName)
			d.convManager.AddFunctionCall(currentToolName, currentArgs, currentRawSDKPart)
			d.convManager.AddFunctionResponse(currentToolName, map[string]any{"error": errStr})
			responseText = fmt.Sprintf("[Erro na Ferramenta: %s não encontrada]", currentToolName)
			break
		}

		// Intercepta e valida consentimento de segurança de forma centralizada no Daemon
		permissionsAllowed := true
		for _, perm := range tool.Describe().Permissions {
			allowed, err := d.RequestPermission(context.Background(), currentToolName, perm)
			if err != nil || !allowed {
				permissionsAllowed = false
				break
			}
		}

		if !permissionsAllowed {
			d.bus.Publish(bus.ToolCompletedEvent{
				ToolName: currentToolName,
				Success:  false,
				Error:    "permission denied by user",
			})
			d.convManager.AddFunctionCall(currentToolName, currentArgs, currentRawSDKPart)
			d.convManager.AddFunctionResponse(currentToolName, map[string]any{"error": "permission denied by user"})
			responseText = "[Erro na Ferramenta: permissão negada pelo usuário]"
			break
		}

		progressChan := make(chan tools.ProgressUpdate, 10)

		// Consome canal de progresso concorrentemente
		go func() {
			for up := range progressChan {
				d.bus.Publish(bus.ToolProgressEvent{
					ToolName: currentToolName,
					State:    string(up.State),
					Percent:  up.Percent,
					Message:  up.Message,
				})
			}
		}()

		res, err := d.toolBus.Execute(context.Background(), currentToolName, tools.Request{
			Args:     currentArgs,
			Progress: progressChan,
		})
		close(progressChan)

		success := err == nil && res.Success
		var out any
		var errStr string
		if err != nil {
			errStr = err.Error()
		} else {
			out = res.Output
			errStr = res.Error
		}

		// Emite conclusão da ferramenta no bus
		d.bus.Publish(bus.ToolCompletedEvent{
			ToolName: currentToolName,
			Success:  success,
			Output:   out,
			Error:    errStr,
		})

		// Registra no histórico da conversa a chamada e a resposta estruturada
		d.convManager.AddFunctionCall(currentToolName, currentArgs, currentRawSDKPart)
		if success {
			d.convManager.AddFunctionResponse(currentToolName, out)
		} else {
			d.convManager.AddFunctionResponse(currentToolName, map[string]any{"error": errStr})
		}

		// Limpa o input inicial para a próxima iteração não cair em regras de comandos CLI
		input = ""
		d.setState(state.Thinking)
		time.Sleep(1 * time.Second)
	}

	// Emitir uma animação teste para o C++
	d.bus.Publish(bus.AnimationPlayEvent{Animation: "smile"})
	time.Sleep(1 * time.Second)

	d.setState(state.Idle)

	return sdkipc.Message{
		Type: "response",
		Payload: sdkipc.Response{
			RefID:  cmd.ID,
			Status: "ok",
			Data:   responseText,
		},
	}
}

func (d *Daemon) errorResponse(refID string, message string) sdkipc.Message {
	return sdkipc.Message{
		Type: "response",
		Payload: sdkipc.Response{
			RefID:  refID,
			Status: "error",
			Data:   message,
		},
	}
}
