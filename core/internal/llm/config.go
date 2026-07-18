package llm

// Config armazena dados explícitos para roteamento e credenciais da LLM
type Config struct {
	Provider string `json:"provider" yaml:"provider"` // ex: "gemini", "openrouter", "mock"
	APIKey   string `json:"api_key" yaml:"api_key"`
	Model    string `json:"model" yaml:"model"`
}
