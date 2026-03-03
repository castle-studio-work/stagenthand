package llm_test

import (
	"testing"

	"github.com/baochen10luo/stagenthand/internal/llm"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewClient(t *testing.T) {
	t.Parallel()

	t.Run("dry-run returns MockClient", func(t *testing.T) {
		client, err := llm.NewClient("anything", true)
		require.NoError(t, err)
		require.NotNil(t, client)

		// dry-run mock must return dry-run-ok payload
		out, err := client.GenerateTransformation(t.Context(), "prompt", []byte("input"))
		assert.NoError(t, err)
		assert.Contains(t, string(out), "dry-run-ok")
	})

	t.Run("provider=mock returns MockClient", func(t *testing.T) {
		client, err := llm.NewClient("mock", false)
		require.NoError(t, err)
		require.NotNil(t, client)
	})

	t.Run("provider=nova returns NovaClient stub", func(t *testing.T) {
		client, err := llm.NewClient("nova", false)
		require.NoError(t, err)
		require.NotNil(t, client)
	})

	t.Run("provider=amazon-nova alias works", func(t *testing.T) {
		client, err := llm.NewClient("amazon-nova", false)
		require.NoError(t, err)
		require.NotNil(t, client)
	})

	t.Run("unknown provider returns error", func(t *testing.T) {
		client, err := llm.NewClient("unknown-provider", false)
		assert.Error(t, err)
		assert.Nil(t, client)
		assert.Contains(t, err.Error(), "not implemented")
	})
}
