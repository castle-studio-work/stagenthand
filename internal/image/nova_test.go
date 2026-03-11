package image

import (
	"context"
	"errors"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime"
)

type mockBedrockInvoker struct {
	output *bedrockruntime.InvokeModelOutput
	err    error
}

func (m *mockBedrockInvoker) InvokeModel(ctx context.Context, params *bedrockruntime.InvokeModelInput, optFns ...func(*bedrockruntime.Options)) (*bedrockruntime.InvokeModelOutput, error) {
	return m.output, m.err
}

func TestNovaCanvasClient_GenerateImage(t *testing.T) {
	tests := []struct {
		name       string
		prompt     string
		mockOutput *bedrockruntime.InvokeModelOutput
		mockErr    error
		wantErr    bool
	}{
		{
			name:   "success",
			prompt: "test prompt",
			mockOutput: &bedrockruntime.InvokeModelOutput{
				Body: []byte(`{"images":["YmFzZTY0aW1hZ2VkYXRh"]}`),
			},
			wantErr: false,
		},
		{
			name:    "invoke error",
			prompt:  "test prompt",
			mockErr: errors.New("invoke error"),
			wantErr: true,
		},
		{
			name:   "invalid json response",
			prompt: "test prompt",
			mockOutput: &bedrockruntime.InvokeModelOutput{
				Body: []byte(`{"images": invalid json`),
			},
			wantErr: true,
		},
		{
			name:   "empty images",
			prompt: "test prompt",
			mockOutput: &bedrockruntime.InvokeModelOutput{
				Body: []byte(`{"images":[]}`),
			},
			wantErr: true,
		},
		{
			name:   "invalid base64",
			prompt: "test prompt",
			mockOutput: &bedrockruntime.InvokeModelOutput{
				Body: []byte(`{"images":["invalid base64!!!"]}`),
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := &NovaCanvasClient{
				client: &mockBedrockInvoker{
					output: tt.mockOutput,
					err:    tt.mockErr,
				},
				width:  1024,
				height: 576,
			}

			img, err := client.GenerateImage(context.Background(), tt.prompt, nil)
			if (err != nil) != tt.wantErr {
				t.Errorf("GenerateImage() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && len(img) == 0 {
				t.Errorf("GenerateImage() returned empty image")
			}
		})
	}
}

func TestNewNovaCanvasClient(t *testing.T) {
	client, err := NewNovaCanvasClient("test_ak", "test_sk", "", "", 0, 0, "")
	if err != nil {
		t.Fatalf("NewNovaCanvasClient() error = %v", err)
	}
	if client.model != "amazon.nova-canvas-v1:0" {
		t.Errorf("expected default model, got %s", client.model)
	}
	if client.width != 1024 {
		t.Errorf("expected default width 1024, got %d", client.width)
	}
}
