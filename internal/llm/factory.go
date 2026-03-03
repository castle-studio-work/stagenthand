package llm

import (
	"context"
	"fmt"
)

// NewClient returns a new LLM client.
// If dryRun is true, it returns a MockClient that responds with a dummy JSON payload.
func NewClient(provider string, dryRun bool) (Client, error) {
	if dryRun {
		return &MockClient{
			GenerateFunc: func(ctx context.Context, systemPrompt string, inputData []byte) ([]byte, error) {
				// Returning a simple valid JSON string to simulate success across all three stages.
				return []byte(`{"status": "dry-run-ok"}`), nil
			},
		}, nil
	}

	// For Phase 2 actual provider implementations (OpenAI/Gemini) will be here later.
	return nil, fmt.Errorf("provider %s not implemented yet. Use --dry-run for testing", provider)
}
