package video_test

import (
	"context"
	"testing"

	"github.com/baochen10luo/stagenthand/internal/video"
)

// mockCriticClient implements llm.VideoCriticClient
type mockCriticClient struct {
	output []byte
	err    error
}

func (m *mockCriticClient) ReviewVideo(ctx context.Context, systemPrompt string, propsJSONData []byte, mediaType string, mediaData []byte) ([]byte, error) {
	return m.output, m.err
}

func TestCheckApproval(t *testing.T) {
	tests := []struct {
		name     string
		eval     video.Evaluation
		approved bool
	}{
		{"All perfect", video.Evaluation{VisualScore: 10, AudioSyncScore: 10, AdherenceScore: 10, ToneScore: 10}, true},
		{"Low Visual FATAL", video.Evaluation{VisualScore: 7, AudioSyncScore: 10, AdherenceScore: 10, ToneScore: 10}, false},
		{"Low Audio FATAL", video.Evaluation{VisualScore: 10, AudioSyncScore: 7, AdherenceScore: 10, ToneScore: 10}, false},
		{"Low Total Score (<32)", video.Evaluation{VisualScore: 8, AudioSyncScore: 8, AdherenceScore: 7, ToneScore: 8}, false},
		{"Barely passes (32)", video.Evaluation{VisualScore: 8, AudioSyncScore: 8, AdherenceScore: 8, ToneScore: 8}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.eval.CheckApproval(); got != tt.approved {
				t.Errorf("CheckApproval() = %v, want %v", got, tt.approved)
			}
		})
	}
}

func TestCritic_Evaluate(t *testing.T) {
	client := &mockCriticClient{
		output: []byte(`{"visual_score": 9, "audio_sync_score": 9, "adherence_score": 9, "tone_score": 9, "action": "APPROVE", "feedback": "Good"}`),
	}
	critic := video.NewCritic(client)

	eval, err := critic.Evaluate(context.Background(), "../../test_storyboard.json", []byte("{}"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if eval.Action != "APPROVE" {
		t.Errorf("expected APPROVE, got %s", eval.Action)
	}
}
