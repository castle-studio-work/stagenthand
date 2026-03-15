package pipeline

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/baochen10luo/stagenthand/internal/audio"
	"github.com/baochen10luo/stagenthand/internal/domain"
)

// MultiSpeakerAudioBatcher adapts an audio.MultiSpeakerClient into the AudioBatcher interface.
// For panels with DialogueLines, it generates one audio file per line and concatenates them.
// For panels with only the legacy Dialogue field, it falls back to a single-line call.
type MultiSpeakerAudioBatcher struct {
	client  audio.MultiSpeakerClient
	rootDir string
}

// NewMultiSpeakerAudioBatcher wraps a MultiSpeakerClient as an AudioBatcher.
func NewMultiSpeakerAudioBatcher(client audio.MultiSpeakerClient, rootDir string) AudioBatcher {
	return &MultiSpeakerAudioBatcher{client: client, rootDir: rootDir}
}

// BatchGenerateAudio generates audio for all panels.
// - If panel.DialogueLines is non-empty, generates one clip per line and concatenates bytes.
// - If panel.DialogueLines is empty but panel.Dialogue != "", falls back to single-line call.
// - Panels with no dialogue are skipped.
// Smart Resume: skips generation if the target mp3 already exists and is non-empty.
func (b *MultiSpeakerAudioBatcher) BatchGenerateAudio(ctx context.Context, panels []domain.Panel, targetDir string) ([]domain.Panel, error) {
	fullDir := filepath.Join(b.rootDir, targetDir)
	if err := os.MkdirAll(fullDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create audio dir %s: %w", fullDir, err)
	}

	result := make([]domain.Panel, len(panels))
	for i, p := range panels {
		result[i] = p

		// Determine whether there is any dialogue to synthesize
		hasLines := len(p.DialogueLines) > 0
		hasFallback := !hasLines && p.Dialogue != ""

		if !hasLines && !hasFallback {
			continue // no dialogue — skip
		}

		filename := fmt.Sprintf("scene_%d_panel_%d.mp3", p.SceneNumber, p.PanelNumber)
		absPath := filepath.Join(fullDir, filename)

		// Smart Resume
		if info, err := os.Stat(absPath); err == nil && info.Size() > 0 {
			result[i].AudioURL = absPath
			continue
		}

		var audioBytes []byte
		var err error

		if hasLines {
			audioBytes, err = b.generateForLines(ctx, p.DialogueLines)
		} else {
			// Fallback: treat the legacy Dialogue field as a single neutral narrator line
			fallbackLine := domain.DialogueLine{
				Speaker: "",
				Text:    p.Dialogue,
				Emotion: "neutral",
			}
			audioBytes, err = b.client.GenerateSpeechForLine(ctx, fallbackLine)
		}
		if err != nil {
			return nil, fmt.Errorf("panel %d-%d audio gen failed: %w", p.SceneNumber, p.PanelNumber, err)
		}

		if len(audioBytes) == 0 {
			continue // nothing to save
		}

		if err := os.WriteFile(absPath, audioBytes, 0644); err != nil {
			return nil, fmt.Errorf("failed to save audio %s: %w", absPath, err)
		}

		result[i].AudioURL = absPath
	}
	return result, nil
}

// generateForLines generates audio for each DialogueLine and concatenates the raw bytes.
func (b *MultiSpeakerAudioBatcher) generateForLines(ctx context.Context, lines []domain.DialogueLine) ([]byte, error) {
	var combined []byte
	for _, line := range lines {
		if line.Text == "" {
			continue
		}
		chunk, err := b.client.GenerateSpeechForLine(ctx, line)
		if err != nil {
			return nil, fmt.Errorf("speech for speaker %q: %w", line.Speaker, err)
		}
		combined = append(combined, chunk...)
	}
	return combined, nil
}
