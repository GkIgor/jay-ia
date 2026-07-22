package daemon

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/GkIgor/jay-ia/core/internal/api"
	"github.com/GkIgor/jay-ia/core/internal/ipc"
	"github.com/GkIgor/jay-ia/core/internal/llm"
	"github.com/GkIgor/jay-ia/core/internal/permission"
	"github.com/GkIgor/jay-ia/core/internal/service"
	"github.com/GkIgor/jay-ia/core/internal/storage"
)

// Repositories agrupa a suíte de repositórios de persistência do Jay Core.
type Repositories struct {
	RegRepo  *storage.RegistrationRepository
	RuleRepo *storage.SharedRuleRepository
	ChatRepo *storage.ChatRepository
	MsgRepo  *storage.MessageRepository
	ToolRepo *storage.ToolRepository
}

// Services agrupa a suíte de serviços de aplicação do Jay Core.
type Services struct {
	RegSvc       *service.RegistrationService
	ChatSvc      *service.ChatService
	MsgSvc       *service.MessageService
	ToolSvc      *service.ToolService
	ProcessorSvc *service.ProcessorService
	Evaluator    *permission.Evaluator
}

// Daemon atua como o Composition Root da aplicação Jay Core, construindo o grafo completo de dependências sem conter regras de negócio.
type Daemon struct {
	engine    *storage.StorageEngine
	router    *api.Router
	server    *ipc.Server
	ctx       context.Context
	cancelCtx context.CancelFunc
}

// New cria e inicializa o Composition Root do Daemon.
func New() (*Daemon, error) {
	dbPath := resolveDatabasePath()
	return NewDaemon(dbPath)
}

// NewDaemon constrói a instância do Daemon com um caminho customizado de banco de dados.
func NewDaemon(dbPath string) (*Daemon, error) {
	ctx, cancel := context.WithCancel(context.Background())

	// 1. Constrói a Camada de Armazenamento SQLite
	engine, err := buildStorage(dbPath)
	if err != nil {
		cancel()
		return nil, fmt.Errorf("daemon: falha ao inicializar banco de dados: %w", err)
	}

	// 2. Constrói os Repositórios
	repos, err := buildRepositories(engine.DB())
	if err != nil {
		cancel()
		_ = engine.Close()
		return nil, fmt.Errorf("daemon: falha ao inicializar repositórios: %w", err)
	}

	// 3. Constrói o Cliente LLM
	llmClient, err := buildLLMClient()
	if err != nil {
		cancel()
		_ = engine.Close()
		return nil, fmt.Errorf("daemon: falha ao inicializar provedor LLM: %w", err)
	}

	// 4. Constrói os Serviços de Aplicação
	services, err := buildServices(repos, llmClient)
	if err != nil {
		cancel()
		_ = engine.Close()
		return nil, fmt.Errorf("daemon: falha ao inicializar serviços: %w", err)
	}

	// 5. Constrói o Roteador RPC e Cadastra os Handlers
	router, err := buildHandlersAndRouter(services)
	if err != nil {
		cancel()
		_ = engine.Close()
		return nil, fmt.Errorf("daemon: falha ao registrar rotas RPC: %w", err)
	}

	// 6. Constrói o Servidor de Socket Unix IPC
	server, err := buildServer(router)
	if err != nil {
		cancel()
		_ = engine.Close()
		return nil, fmt.Errorf("daemon: falha ao iniciar servidor IPC: %w", err)
	}

	return &Daemon{
		engine:    engine,
		router:    router,
		server:    server,
		ctx:       ctx,
		cancelCtx: cancel,
	}, nil
}

// Start inicia o servidor Unix Socket e bloqueia até o recebimento de sinais de encerramento do SO (SIGINT/SIGTERM).
func (d *Daemon) Start() error {
	log.Println("Iniciando Jay Core Daemon (Composition Root)...")

	if err := d.server.Start(); err != nil {
		return fmt.Errorf("daemon: falha ao iniciar servidor de socket IPC: %w", err)
	}

	log.Println("Jay Core Daemon pronto e aguardando requisições IPC.")

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	sig := <-sigChan
	log.Printf("Sinal de término recebido (%v). Encerrando Daemon graciosamente...", sig)

	return d.Stop()
}

// Stop encerra ordenadamente o Daemon (cancelCtx -> server.Stop -> engine.Close).
func (d *Daemon) Stop() error {
	log.Println("Encerrando Jay Core Daemon...")

	if d.cancelCtx != nil {
		d.cancelCtx()
	}

	if d.server != nil {
		d.server.Stop()
	}

	if d.engine != nil {
		if err := d.engine.Close(); err != nil {
			log.Printf("Erro ao fechar conexões com SQLite: %v", err)
			return err
		}
	}

	log.Println("Jay Core Daemon encerrado com sucesso.")
	return nil
}

// Router retorna o Roteador RPC instanciado.
func (d *Daemon) Router() *api.Router {
	return d.router
}

// --- Builders Privados ---

func buildStorage(dbPath string) (*storage.StorageEngine, error) {
	cfg := storage.Config{
		DatabasePath: dbPath,
	}
	engine, err := storage.NewStorageEngine(cfg)
	if err != nil {
		return nil, err
	}
	if err := engine.Open(); err != nil {
		return nil, err
	}

	migrator, err := storage.NewMigrationEngine(engine.DB())
	if err != nil {
		_ = engine.Close()
		return nil, err
	}
	if err := migrator.Run(); err != nil {
		_ = engine.Close()
		return nil, err
	}

	return engine, nil
}

func buildRepositories(db *sql.DB) (*Repositories, error) {
	regRepo, err := storage.NewRegistrationRepository(db)
	if err != nil {
		return nil, err
	}
	ruleRepo, err := storage.NewSharedRuleRepository(db)
	if err != nil {
		return nil, err
	}
	chatRepo, err := storage.NewChatRepository(db)
	if err != nil {
		return nil, err
	}
	msgRepo, err := storage.NewMessageRepository(db)
	if err != nil {
		return nil, err
	}
	toolRepo, err := storage.NewToolRepository(db)
	if err != nil {
		return nil, err
	}

	return &Repositories{
		RegRepo:  regRepo,
		RuleRepo: ruleRepo,
		ChatRepo: chatRepo,
		MsgRepo:  msgRepo,
		ToolRepo: toolRepo,
	}, nil
}

func buildLLMClient() (llm.Client, error) {
	provider := strings.ToLower(strings.TrimSpace(os.Getenv("LLM_PROVIDER")))
	apiKey := ""
	model := ""

	if provider == "" {
		if os.Getenv("OPENROUTER_API_KEY") != "" {
			provider = "openrouter"
		} else if os.Getenv("GEMINI_API_KEY") != "" {
			provider = "gemini"
		} else {
			provider = "mock"
		}
	}

	switch provider {
	case "openrouter":
		apiKey = os.Getenv("OPENROUTER_API_KEY")
		model = os.Getenv("OPENROUTER_MODEL")
	case "gemini":
		apiKey = os.Getenv("GEMINI_API_KEY")
	case "mock":
		log.Println("AVISO: Provedor LLM configurado para 'mock'.")
	default:
		return nil, fmt.Errorf("provedor LLM desconhecido ou inválido '%s'", provider)
	}

	return llm.NewClient(llm.Config{
		Provider: provider,
		APIKey:   apiKey,
		Model:    model,
	})
}

func buildServices(repos *Repositories, llmClient llm.Client) (*Services, error) {
	evaluator := permission.NewEvaluator()

	regSvc, err := service.NewRegistrationService(repos.RegRepo, repos.RuleRepo, evaluator)
	if err != nil {
		return nil, err
	}

	chatSvc, err := service.NewChatService(repos.ChatRepo, repos.RegRepo, repos.RuleRepo, evaluator)
	if err != nil {
		return nil, err
	}

	msgSvc, err := service.NewMessageService(repos.MsgRepo, repos.ChatRepo, repos.RuleRepo, evaluator)
	if err != nil {
		return nil, err
	}

	toolSvc, err := service.NewToolService(repos.ToolRepo, repos.RuleRepo, evaluator)
	if err != nil {
		return nil, err
	}

	procSvc, err := service.NewProcessorService(repos.MsgRepo, repos.ChatRepo, repos.ToolRepo, repos.RuleRepo, evaluator, llmClient)
	if err != nil {
		return nil, err
	}

	return &Services{
		RegSvc:       regSvc,
		ChatSvc:      chatSvc,
		MsgSvc:       msgSvc,
		ToolSvc:      toolSvc,
		ProcessorSvc: procSvc,
		Evaluator:    evaluator,
	}, nil
}

func buildHandlersAndRouter(services *Services) (*api.Router, error) {
	router := api.NewRouter()

	regHandler, err := api.NewRegistrationHandler(services.RegSvc)
	if err != nil {
		return nil, err
	}
	regHandler.RegisterRoutes(router)

	chatHandler, err := api.NewChatHandler(services.ChatSvc)
	if err != nil {
		return nil, err
	}
	chatHandler.RegisterRoutes(router)

	msgHandler, err := api.NewMessageHandler(services.MsgSvc)
	if err != nil {
		return nil, err
	}
	msgHandler.RegisterRoutes(router)

	toolHandler, err := api.NewToolHandler(services.ToolSvc)
	if err != nil {
		return nil, err
	}
	toolHandler.RegisterRoutes(router)

	procHandler, err := api.NewProcessorHandler(services.ProcessorSvc)
	if err != nil {
		return nil, err
	}
	procHandler.RegisterRoutes(router)

	return router, nil
}

func buildServer(router *api.Router) (*ipc.Server, error) {
	return ipc.NewRawServer(router.Dispatch)
}

func resolveDatabasePath() string {
	if dbEnv := os.Getenv("JAY_DB_PATH"); dbEnv != "" {
		return dbEnv
	}

	if xdgData := os.Getenv("XDG_DATA_HOME"); xdgData != "" {
		return filepath.Join(xdgData, "jay", "jay.db")
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return "./jay.db"
	}
	return filepath.Join(home, ".jay", "jay.db")
}
