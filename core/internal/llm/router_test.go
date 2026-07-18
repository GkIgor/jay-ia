package llm

import (
	"testing"
)

func TestRouter(t *testing.T) {
	// Test mock client
	c, err := NewClient(Config{Provider: "mock"})
	if err != nil {
		t.Fatalf("failed to create mock client: %v", err)
	}
	if _, ok := c.(*MockClient); !ok {
		t.Errorf("expected MockClient instance")
	}

	// Test invalid provider
	_, err = NewClient(Config{Provider: "invalid"})
	if err == nil {
		t.Errorf("expected error for invalid provider")
	}

	// Test missing API key for Gemini
	_, err = NewClient(Config{Provider: "gemini", APIKey: ""})
	if err == nil {
		t.Errorf("expected error for Gemini with missing API Key")
	}

	// Test OpenRouter with valid/missing keys
	_, err = NewClient(Config{Provider: "openrouter", APIKey: ""})
	if err == nil {
		t.Errorf("expected error for OpenRouter with missing API Key")
	}

	orc, err := NewClient(Config{Provider: "openrouter", APIKey: "test-key", Model: "test-model"})
	if err != nil {
		t.Fatalf("failed to create OpenRouter client: %v", err)
	}
	if _, ok := orc.(*OpenRouterClient); !ok {
		t.Errorf("expected OpenRouterClient instance")
	}
}
