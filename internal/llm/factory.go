package llm

import (
	"context"
	"fmt"

	"github.com/baochen10luo/stagenthand/config"
)

// NewClient returns a new LLM client.
// If dryRun is true, it returns a MockClient that responds with a dummy JSON payload.
func NewClient(provider string, dryRun bool, cfg *config.Config) (Client, error) {
	if dryRun || provider == "mock" {
		return &MockClient{
			GenerateFunc: func(ctx context.Context, systemPrompt string, inputData []byte) ([]byte, error) {
				return []byte(`{"status": "dry-run-ok"}`), nil
			},
		}, nil
	}

	switch provider {
	case "openai", "gemini":
		model := ""
		baseURL := ""
		apiKey := ""
		if cfg != nil {
			model = cfg.LLM.Model
			baseURL = cfg.LLM.BaseURL
			apiKey = cfg.LLM.APIKey
		}
		if model == "" {
			if provider == "gemini" {
				model = "gemini-2.5-pro"
			} else {
				model = "gpt-4o"
			}
		}
		return NewOpenAICompatibleClient(baseURL, apiKey, model), nil
	case "bedrock":
		if cfg == nil {
			return nil, fmt.Errorf("bedrock provider requires config")
		}
		modelID := cfg.LLM.Model
		if modelID == "" {
			modelID = "amazon.nova-pro-v1:0"
		}
		return NewBedrockClient(
			cfg.LLM.AWSAccessKeyID,
			cfg.LLM.AWSSecretAccessKey,
			cfg.LLM.AWSRegion,
			modelID,
		)
	default:
		return nil, fmt.Errorf("provider %s not implemented yet. Use --dry-run for testing", provider)
	}
}
