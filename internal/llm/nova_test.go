package llm_test

import (
	"context"
	"testing"

	"github.com/baochen10luo/stagenthand/internal/llm"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewNovaClient_Defaults(t *testing.T) {
	t.Parallel()

	client := llm.NewNovaClient("", "", "")
	require.NotNil(t, client)
	assert.Equal(t, "https://nova.aws.amazon.com/v1", client.BaseURL)
	assert.Equal(t, "amazon.nova-pro-v1:0", client.Model)
}

func TestNewNovaClient_Custom(t *testing.T) {
	t.Parallel()

	client := llm.NewNovaClient("my-key", "https://bedrock.us-east-1.amazonaws.com", "amazon.nova-lite-v1:0")
	require.NotNil(t, client)
	assert.Equal(t, "my-key", client.APIKey)
	assert.Equal(t, "https://bedrock.us-east-1.amazonaws.com", client.BaseURL)
	assert.Equal(t, "amazon.nova-lite-v1:0", client.Model)
}

func TestNovaClient_GenerateTransformation_NoKey(t *testing.T) {
	t.Parallel()

	client := llm.NewNovaClient("", "", "")
	_, err := client.GenerateTransformation(context.Background(), "system", []byte("input"))
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "API key not configured")
}

func TestNovaClient_GenerateTransformation_StubWithKey(t *testing.T) {
	t.Parallel()

	// With a key set, it should fail with "not yet implemented" (stub), not "no key"
	client := llm.NewNovaClient("fake-key", "", "")
	_, err := client.GenerateTransformation(context.Background(), "system", []byte("input"))
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not yet implemented")
}
