package llm

import (
	"context"
	"fmt"
)

// NovaClient is a stub for Amazon Nova Act LLM integration.
// Nova Act is Amazon's UI automation / agent SDK (hackathon target).
// Phase 3+ will replace the stub with actual Nova Act API calls.
//
// Architecture note: Nova Act operates differently from standard LLM providers —
// it uses an "act" instruction paradigm rather than chat completion.
// This wrapper normalizes it to the llm.Client interface so the pipeline
// remains provider-agnostic.
type NovaClient struct {
	APIKey  string
	BaseURL string
	Model   string
}

// NewNovaClient initializes a Nova Act client.
// baseURL defaults to the Amazon Nova Act endpoint when empty.
func NewNovaClient(apiKey, baseURL, model string) *NovaClient {
	if baseURL == "" {
		baseURL = "https://nova.aws.amazon.com/v1"
	}
	if model == "" {
		model = "amazon.nova-pro-v1:0"
	}
	return &NovaClient{
		APIKey:  apiKey,
		BaseURL: baseURL,
		Model:   model,
	}
}

// GenerateTransformation implements llm.Client.
// TODO (Phase 3+): Replace stub with actual Nova Act SDK call.
// Nova Act uses a browser/UI context for automation; for text-only pipeline steps,
// we can route to Nova Lite/Pro via Bedrock's converse API.
func (c *NovaClient) GenerateTransformation(ctx context.Context, systemPrompt string, inputData []byte) ([]byte, error) {
	if c.APIKey == "" {
		return nil, fmt.Errorf("nova: API key not configured")
	}
	// Stub: returns placeholder until Bedrock SDK integration is wired up.
	return nil, fmt.Errorf("nova: provider not yet implemented — API key is set, wire up Bedrock SDK in Phase 3")
}
