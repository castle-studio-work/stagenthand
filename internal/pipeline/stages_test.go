package pipeline_test

import (
	"context"
	"errors"
	"testing"

	"github.com/baochen10luo/stagenthand/internal/llm"
	"github.com/baochen10luo/stagenthand/internal/pipeline"
	"github.com/stretchr/testify/assert"
)

func TestRunTransformationStage(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		client := &llm.MockClient{
			GenerateFunc: func(ctx context.Context, systemPrompt string, inputData []byte) ([]byte, error) {
				return []byte(`{"ok": true}`), nil
			},
		}

		res, err := pipeline.RunTransformationStage(context.Background(), client, pipeline.PromptStoryToOutline, []byte("input story"))
		assert.NoError(t, err)
		assert.Equal(t, []byte(`{"ok": true}`), res)
		assert.Equal(t, 1, client.CallCount)
	})

	t.Run("empty input", func(t *testing.T) {
		client := &llm.MockClient{}
		_, err := pipeline.RunTransformationStage(context.Background(), client, pipeline.PromptStoryToOutline, nil)
		assert.ErrorContains(t, err, "input data cannot be empty")
		assert.Equal(t, 0, client.CallCount)
	})

	t.Run("empty sysprompt", func(t *testing.T) {
		client := &llm.MockClient{}
		_, err := pipeline.RunTransformationStage(context.Background(), client, "", []byte("123"))
		assert.ErrorContains(t, err, "system prompt cannot be empty")
		assert.Equal(t, 0, client.CallCount)
	})

	t.Run("llm failure", func(t *testing.T) {
		myErr := errors.New("API limit")
		client := &llm.MockClient{
			GenerateFunc: func(ctx context.Context, systemPrompt string, inputData []byte) ([]byte, error) {
				return nil, myErr
			},
		}
		_, err := pipeline.RunTransformationStage(context.Background(), client, pipeline.PromptStoryToOutline, []byte("req"))
		assert.ErrorIs(t, err, myErr)
		assert.ErrorContains(t, err, "transformer failed")
	})

	t.Run("llm returns empty", func(t *testing.T) {
		client := &llm.MockClient{
			GenerateFunc: func(ctx context.Context, systemPrompt string, inputData []byte) ([]byte, error) {
				return nil, nil
			},
		}
		_, err := pipeline.RunTransformationStage(context.Background(), client, pipeline.PromptStoryToOutline, []byte("req"))
		assert.ErrorContains(t, err, "transformer returned empty output")
	})
}
