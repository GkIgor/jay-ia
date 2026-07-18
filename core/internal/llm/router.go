package llm

import (
	"fmt"
)

// NewClient cria um provedor de LLM baseado nas configurações fornecidas
func NewClient(cfg Config) (Client, error) {
	switch cfg.Provider {
	case "gemini":
		if cfg.APIKey == "" {
			return nil, fmt.Errorf("api key is required for provider gemini")
		}
		return NewGeminiClient(cfg.APIKey)
	case "openrouter":
		if cfg.APIKey == "" {
			return nil, fmt.Errorf("api key is required for provider openrouter")
		}
		model := cfg.Model
		if model == "" {
			model = "google/gemini-2.5-flash"
		}
		return NewOpenRouterClient(cfg.APIKey, model)
	case "mock":
		return NewMockClient(), nil
	default:
		return nil, fmt.Errorf("unknown llm provider: %s", cfg.Provider)
	}
}
