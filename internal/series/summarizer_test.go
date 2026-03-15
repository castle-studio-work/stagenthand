package series_test

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/baochen10luo/stagenthand/internal/series"
)

// mockLLMClient implements series.LLMClient for testing.
type mockLLMClient struct {
	output []byte
	err    error
}

func (m *mockLLMClient) GenerateTransformation(_ context.Context, _ string, _ []byte) ([]byte, error) {
	return m.output, m.err
}

func TestLLMSummarizer_Summarize_ParsesJSON(t *testing.T) {
	expected := series.EpisodeMemory{
		Episode:   1,
		KeyEvents: []string{"hero battles dragon", "mentor sacrificed"},
		Characters: []series.CharacterSnapshot{
			{Name: "Hero", Description: "brave warrior", Motivation: "avenge family", State: "victorious but sad"},
		},
		WorldFacts: []string{"dragons can speak"},
	}

	responseJSON, _ := json.Marshal(expected)

	client := &mockLLMClient{output: responseJSON}
	summarizer := series.NewLLMSummarizer(client)

	result, err := summarizer.Summarize(context.Background(), 1, []byte(`{"scenes":[]}`))
	if err != nil {
		t.Fatalf("Summarize failed: %v", err)
	}

	if result.Episode != 1 {
		t.Errorf("Episode: got %d, want 1", result.Episode)
	}
	if len(result.KeyEvents) != 2 {
		t.Errorf("KeyEvents length: got %d, want 2", len(result.KeyEvents))
	}
	if result.KeyEvents[0] != "hero battles dragon" {
		t.Errorf("KeyEvents[0]: got %q, want %q", result.KeyEvents[0], "hero battles dragon")
	}
	if len(result.Characters) != 1 {
		t.Errorf("Characters length: got %d, want 1", len(result.Characters))
	}
	if result.Characters[0].Name != "Hero" {
		t.Errorf("Characters[0].Name: got %q, want %q", result.Characters[0].Name, "Hero")
	}
	if len(result.WorldFacts) != 1 {
		t.Errorf("WorldFacts length: got %d, want 1", len(result.WorldFacts))
	}
}

func TestLLMSummarizer_Summarize_EnsuresEpisodeNum(t *testing.T) {
	// LLM might return wrong episode number; we override it
	ep := series.EpisodeMemory{Episode: 99, KeyEvents: []string{"event"}}
	responseJSON, _ := json.Marshal(ep)

	client := &mockLLMClient{output: responseJSON}
	summarizer := series.NewLLMSummarizer(client)

	result, err := summarizer.Summarize(context.Background(), 5, []byte(`{}`))
	if err != nil {
		t.Fatalf("Summarize failed: %v", err)
	}
	if result.Episode != 5 {
		t.Errorf("Episode should be overridden to 5, got %d", result.Episode)
	}
}

func TestLLMSummarizer_CompressGlobal_ReturnsString(t *testing.T) {
	expected := "A brave hero fought many battles and grew stronger with each challenge."
	client := &mockLLMClient{output: []byte(expected)}
	summarizer := series.NewLLMSummarizer(client)

	m := &series.SeriesMemory{
		SeriesTitle: "Dragon Wars",
		Episodes: []series.EpisodeMemory{
			{Episode: 1, KeyEvents: []string{"fight 1"}},
		},
	}

	result, err := summarizer.CompressGlobal(context.Background(), m)
	if err != nil {
		t.Fatalf("CompressGlobal failed: %v", err)
	}
	if result != expected {
		t.Errorf("CompressGlobal result: got %q, want %q", result, expected)
	}
}

func TestLLMSummarizer_Summarize_InvalidJSON(t *testing.T) {
	client := &mockLLMClient{output: []byte("not valid json")}
	summarizer := series.NewLLMSummarizer(client)

	_, err := summarizer.Summarize(context.Background(), 1, []byte(`{}`))
	if err == nil {
		t.Error("expected error for invalid JSON, got nil")
	}
}
