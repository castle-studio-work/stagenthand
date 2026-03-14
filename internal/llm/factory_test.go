package llm_test

import (
	"testing"

	"github.com/baochen10luo/stagenthand/config"
	"github.com/baochen10luo/stagenthand/internal/llm"
	"github.com/stretchr/testify/assert"
)

func TestNewClient(t *testing.T) {
	t.Parallel()

	cfg := &config.Config{
		LLM: config.LLMConfig{
			APIKey: "test",
		},
	}

	t.Run("dry run", func(t *testing.T) {
		client, err := llm.NewClient("gemini", true, cfg)
		assert.NoError(t, err)
		_, ok := client.(*llm.MockClient)
		assert.True(t, ok)
	})

	t.Run("mock provider", func(t *testing.T) {
		client, err := llm.NewClient("mock", false, cfg)
		assert.NoError(t, err)
		_, ok := client.(*llm.MockClient)
		assert.True(t, ok)
	})

	t.Run("gemini provider", func(t *testing.T) {
		client, err := llm.NewClient("gemini", false, cfg)
		assert.NoError(t, err)
		_, ok := client.(*llm.OpenAICompatibleClient)
		assert.True(t, ok)
	})

	t.Run("openai provider", func(t *testing.T) {
		client, err := llm.NewClient("openai", false, cfg)
		assert.NoError(t, err)
		_, ok := client.(*llm.OpenAICompatibleClient) // maps to OpenAICompatible internally
		assert.True(t, ok)
	})

	t.Run("bedrock provider", func(t *testing.T) {
		bedrockCfg := &config.Config{
			LLM: config.LLMConfig{
				AWSAccessKeyID:     "AKIATEST",
				AWSSecretAccessKey: "secret",
				AWSRegion:          "us-east-1",
			},
		}
		client, err := llm.NewClient("bedrock", false, bedrockCfg)
		assert.NoError(t, err)
		_, ok := client.(*llm.BedrockClient)
		assert.True(t, ok)
	})

	t.Run("bedrock without config", func(t *testing.T) {
		client, err := llm.NewClient("bedrock", false, nil)
		assert.ErrorContains(t, err, "requires config")
		assert.Nil(t, client)
	})

	t.Run("unknown provider", func(t *testing.T) {
		client, err := llm.NewClient("unknown", false, nil)
		assert.ErrorContains(t, err, "not implemented")
		assert.Nil(t, client)
	})
}
