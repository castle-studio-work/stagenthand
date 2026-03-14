package llm_test

import (
	"context"
	"errors"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime"
	brtypes "github.com/aws/aws-sdk-go-v2/service/bedrockruntime/types"
	"github.com/baochen10luo/stagenthand/internal/llm"
	"github.com/stretchr/testify/assert"
)

// mockBedrockAPI is a test double that satisfies llm.BedrockAPI.
type mockBedrockAPI struct {
	ConverseFunc func(ctx context.Context, params *bedrockruntime.ConverseInput, optFns ...func(*bedrockruntime.Options)) (*bedrockruntime.ConverseOutput, error)
}

func (m *mockBedrockAPI) Converse(ctx context.Context, params *bedrockruntime.ConverseInput, optFns ...func(*bedrockruntime.Options)) (*bedrockruntime.ConverseOutput, error) {
	return m.ConverseFunc(ctx, params, optFns...)
}

func TestBedrockClient_GenerateTransformation(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		mock := &mockBedrockAPI{
			ConverseFunc: func(ctx context.Context, params *bedrockruntime.ConverseInput, optFns ...func(*bedrockruntime.Options)) (*bedrockruntime.ConverseOutput, error) {
				// Verify model ID is passed through
				assert.Equal(t, "amazon.nova-pro-v1:0", *params.ModelId)
				// Verify system prompt is set
				assert.Len(t, params.System, 1)
				// Verify user message is set
				assert.Len(t, params.Messages, 1)

				responseText := `{"outline": {"title": "test"}}`
				return &bedrockruntime.ConverseOutput{
					Output: &brtypes.ConverseOutputMemberMessage{
						Value: brtypes.Message{
							Role: brtypes.ConversationRoleAssistant,
							Content: []brtypes.ContentBlock{
								&brtypes.ContentBlockMemberText{
									Value: responseText,
								},
							},
						},
					},
				}, nil
			},
		}

		client := llm.NewBedrockClientWithAPI(mock, "amazon.nova-pro-v1:0")
		res, err := client.GenerateTransformation(context.Background(), "You are a director.", []byte("test input"))

		assert.NoError(t, err)
		assert.JSONEq(t, `{"outline": {"title": "test"}}`, string(res))
	})

	t.Run("api error", func(t *testing.T) {
		mock := &mockBedrockAPI{
			ConverseFunc: func(ctx context.Context, params *bedrockruntime.ConverseInput, optFns ...func(*bedrockruntime.Options)) (*bedrockruntime.ConverseOutput, error) {
				return nil, errors.New("ThrottlingException: rate exceeded")
			},
		}

		client := llm.NewBedrockClientWithAPI(mock, "amazon.nova-pro-v1:0")
		_, err := client.GenerateTransformation(context.Background(), "sys", []byte("in"))

		assert.ErrorContains(t, err, "bedrock converse failed")
		assert.ErrorContains(t, err, "ThrottlingException")
	})

	t.Run("empty response", func(t *testing.T) {
		mock := &mockBedrockAPI{
			ConverseFunc: func(ctx context.Context, params *bedrockruntime.ConverseInput, optFns ...func(*bedrockruntime.Options)) (*bedrockruntime.ConverseOutput, error) {
				return &bedrockruntime.ConverseOutput{
					Output: &brtypes.ConverseOutputMemberMessage{
						Value: brtypes.Message{
							Role:    brtypes.ConversationRoleAssistant,
							Content: []brtypes.ContentBlock{},
						},
					},
				}, nil
			},
		}

		client := llm.NewBedrockClientWithAPI(mock, "amazon.nova-pro-v1:0")
		_, err := client.GenerateTransformation(context.Background(), "sys", []byte("in"))

		assert.ErrorContains(t, err, "empty response")
	})

	t.Run("markdown strip", func(t *testing.T) {
		mock := &mockBedrockAPI{
			ConverseFunc: func(ctx context.Context, params *bedrockruntime.ConverseInput, optFns ...func(*bedrockruntime.Options)) (*bedrockruntime.ConverseOutput, error) {
				wrapped := "```json\n{\"clean\": true}\n```"
				return &bedrockruntime.ConverseOutput{
					Output: &brtypes.ConverseOutputMemberMessage{
						Value: brtypes.Message{
							Role: brtypes.ConversationRoleAssistant,
							Content: []brtypes.ContentBlock{
								&brtypes.ContentBlockMemberText{Value: wrapped},
							},
						},
					},
				}, nil
			},
		}

		client := llm.NewBedrockClientWithAPI(mock, "amazon.nova-pro-v1:0")
		res, err := client.GenerateTransformation(context.Background(), "sys", []byte("in"))

		assert.NoError(t, err)
		assert.JSONEq(t, `{"clean": true}`, string(res))
	})

	t.Run("nil output", func(t *testing.T) {
		mock := &mockBedrockAPI{
			ConverseFunc: func(ctx context.Context, params *bedrockruntime.ConverseInput, optFns ...func(*bedrockruntime.Options)) (*bedrockruntime.ConverseOutput, error) {
				return &bedrockruntime.ConverseOutput{Output: nil}, nil
			},
		}

		client := llm.NewBedrockClientWithAPI(mock, "amazon.nova-pro-v1:0")
		_, err := client.GenerateTransformation(context.Background(), "sys", []byte("in"))

		assert.ErrorContains(t, err, "unexpected output type")
	})
}

func TestNewBedrockClient_InvalidCreds(t *testing.T) {
	t.Parallel()

	_, err := llm.NewBedrockClient("", "secret", "us-east-1", "model")
	assert.ErrorContains(t, err, "aws_access_key_id is required")

	_, err = llm.NewBedrockClient("key", "", "us-east-1", "model")
	assert.ErrorContains(t, err, "aws_secret_access_key is required")
}
