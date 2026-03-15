package pipeline_test

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/baochen10luo/stagenthand/internal/audio"
	"github.com/baochen10luo/stagenthand/internal/domain"
	"github.com/baochen10luo/stagenthand/internal/pipeline"
)

func TestMultiSpeakerBatcher_UsesDialogueLines(t *testing.T) {
	tmpDir := t.TempDir()
	mock := &audio.MockMultiSpeakerClient{Data: []byte("mp3-data")}
	batcher := pipeline.NewMultiSpeakerAudioBatcher(mock, tmpDir)

	panels := []domain.Panel{
		{
			SceneNumber: 1,
			PanelNumber: 1,
			DialogueLines: []domain.DialogueLine{
				{Speaker: "Alice", Text: "Hello", Emotion: "happy"},
				{Speaker: "Bob", Text: "World", Emotion: "neutral"},
			},
		},
	}

	result, err := batcher.BatchGenerateAudio(context.Background(), panels, "proj/audio")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result[0].AudioURL == "" {
		t.Error("expected AudioURL to be set")
	}
	// Should have been called for each dialogue line
	if len(mock.Calls) != 2 {
		t.Errorf("expected 2 calls (one per dialogue line), got %d", len(mock.Calls))
	}
	if mock.Calls[0].Line.Speaker != "Alice" {
		t.Errorf("first call speaker = %q, want Alice", mock.Calls[0].Line.Speaker)
	}
	if mock.Calls[1].Line.Speaker != "Bob" {
		t.Errorf("second call speaker = %q, want Bob", mock.Calls[1].Line.Speaker)
	}
}

func TestMultiSpeakerBatcher_FallbackToDialogue(t *testing.T) {
	tmpDir := t.TempDir()
	mock := &audio.MockMultiSpeakerClient{Data: []byte("mp3-data")}
	batcher := pipeline.NewMultiSpeakerAudioBatcher(mock, tmpDir)

	panels := []domain.Panel{
		{
			SceneNumber:   1,
			PanelNumber:   1,
			Dialogue:      "Some fallback text",
			DialogueLines: nil, // empty
		},
	}

	result, err := batcher.BatchGenerateAudio(context.Background(), panels, "proj/audio")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result[0].AudioURL == "" {
		t.Error("expected AudioURL to be set via fallback")
	}
	if len(mock.Calls) != 1 {
		t.Errorf("expected 1 fallback call, got %d", len(mock.Calls))
	}
	if mock.Calls[0].Line.Text != "Some fallback text" {
		t.Errorf("fallback text = %q, want %q", mock.Calls[0].Line.Text, "Some fallback text")
	}
	if mock.Calls[0].Line.Emotion != "neutral" {
		t.Errorf("fallback emotion = %q, want neutral", mock.Calls[0].Line.Emotion)
	}
}

func TestMultiSpeakerBatcher_SmartResume(t *testing.T) {
	tmpDir := t.TempDir()
	mock := &audio.MockMultiSpeakerClient{Err: errors.New("SHOULD NOT BE CALLED")}
	batcher := pipeline.NewMultiSpeakerAudioBatcher(mock, tmpDir)

	// Pre-create the mp3 file
	audioDir := filepath.Join(tmpDir, "proj/audio")
	os.MkdirAll(audioDir, 0755)
	existingPath := filepath.Join(audioDir, "scene_1_panel_1.mp3")
	os.WriteFile(existingPath, []byte("existing-audio"), 0644)

	panels := []domain.Panel{
		{
			SceneNumber: 1,
			PanelNumber: 1,
			DialogueLines: []domain.DialogueLine{
				{Speaker: "Alice", Text: "Hello"},
			},
		},
	}

	result, err := batcher.BatchGenerateAudio(context.Background(), panels, "proj/audio")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result[0].AudioURL != existingPath {
		t.Errorf("expected AudioURL = %q, got %q", existingPath, result[0].AudioURL)
	}
	if len(mock.Calls) != 0 {
		t.Errorf("expected 0 calls (smart resume), got %d", len(mock.Calls))
	}
}

func TestMultiSpeakerBatcher_EmptyDialogue(t *testing.T) {
	tmpDir := t.TempDir()
	mock := &audio.MockMultiSpeakerClient{}
	batcher := pipeline.NewMultiSpeakerAudioBatcher(mock, tmpDir)

	panels := []domain.Panel{
		{
			SceneNumber:   1,
			PanelNumber:   1,
			Dialogue:      "",
			DialogueLines: nil,
		},
	}

	result, err := batcher.BatchGenerateAudio(context.Background(), panels, "proj/audio")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result[0].AudioURL != "" {
		t.Errorf("expected empty AudioURL for panel with no dialogue, got %q", result[0].AudioURL)
	}
	if len(mock.Calls) != 0 {
		t.Errorf("expected 0 calls for empty dialogue panel, got %d", len(mock.Calls))
	}
}

func TestMultiSpeakerBatcher_ErrorPropagates(t *testing.T) {
	tmpDir := t.TempDir()
	mock := &audio.MockMultiSpeakerClient{Err: errors.New("TTS service down")}
	batcher := pipeline.NewMultiSpeakerAudioBatcher(mock, tmpDir)

	panels := []domain.Panel{
		{
			SceneNumber: 1,
			PanelNumber: 1,
			DialogueLines: []domain.DialogueLine{
				{Speaker: "Alice", Text: "Hello"},
			},
		},
	}

	_, err := batcher.BatchGenerateAudio(context.Background(), panels, "proj/audio")
	if err == nil {
		t.Error("expected error to propagate from MultiSpeakerClient, got nil")
	}
}

func TestMultiSpeakerBatcher_MultiplePanels(t *testing.T) {
	tmpDir := t.TempDir()
	mock := &audio.MockMultiSpeakerClient{Data: []byte("audio")}
	batcher := pipeline.NewMultiSpeakerAudioBatcher(mock, tmpDir)

	panels := []domain.Panel{
		{
			SceneNumber: 1,
			PanelNumber: 1,
			DialogueLines: []domain.DialogueLine{
				{Speaker: "Alice", Text: "Hello"},
			},
		},
		{
			SceneNumber:   1,
			PanelNumber:   2,
			Dialogue:      "",
			DialogueLines: nil,
		},
		{
			SceneNumber: 2,
			PanelNumber: 1,
			Dialogue:    "Fallback text",
		},
	}

	result, err := batcher.BatchGenerateAudio(context.Background(), panels, "proj/audio")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result[0].AudioURL == "" {
		t.Error("panel 1-1: expected AudioURL to be set")
	}
	if result[1].AudioURL != "" {
		t.Errorf("panel 1-2 (no dialogue): expected empty AudioURL, got %q", result[1].AudioURL)
	}
	if result[2].AudioURL == "" {
		t.Error("panel 2-1 (fallback): expected AudioURL to be set")
	}
}
