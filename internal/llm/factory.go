package llm

import (
	"context"
	"fmt"
)

// NewClient returns a new LLM client based on the provider name.
// If dryRun is true, always returns a MockClient regardless of provider.
func NewClient(provider string, dryRun bool) (Client, error) {
	if dryRun {
		return &MockClient{
			GenerateFunc: func(ctx context.Context, systemPrompt string, inputData []byte) ([]byte, error) {
				return []byte(`{"status": "dry-run-ok"}`), nil
			},
		}, nil
	}

	switch provider {
	case "nova", "amazon-nova":
		// API key injected via config; populated by caller before NewClient.
		// For now returns a stub that fails loudly until Phase 3 wires Bedrock SDK.
		return NewNovaClient("", "", ""), nil
	case "mock":
		return &MockClient{}, nil
	default:
		return nil, fmt.Errorf("provider %q not implemented — use --dry-run or set provider to \"nova\"", provider)
	}
}
