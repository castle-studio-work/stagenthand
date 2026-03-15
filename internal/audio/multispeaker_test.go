package audio

import (
	"context"
	"os"
	"os/exec"
	"strings"
	"testing"

	"github.com/baochen10luo/stagenthand/internal/character"
	"github.com/baochen10luo/stagenthand/internal/domain"
)

// mockCommandFactory captures the SSML passed to the fake aws command for assertion.
type captureSSMLFactory struct {
	ssml    string
	outData []byte
}

func (f *captureSSMLFactory) factory(ctx context.Context, name string, args ...string) *exec.Cmd {
	// Find --text arg (it follows "--text" flag)
	for i, a := range args {
		if a == "--text" && i+1 < len(args) {
			f.ssml = args[i+1]
		}
	}
	// Find the output path (last arg) and write fake data
	outPath := args[len(args)-1]
	data := f.outData
	if data == nil {
		data = []byte("fake-mp3")
	}
	os.WriteFile(outPath, data, 0644) //nolint:errcheck
	return exec.Command("true")
}

func TestPollyMultiSpeakerClient_FallbackToDefault(t *testing.T) {
	reg := character.NewMockRegistry()
	// "Unknown" speaker is not registered

	client := NewPollyMultiSpeakerClient("us-east-1", "ak", "sk", "en-US", reg)

	cap := &captureSSMLFactory{}
	// Inject the mock factory into the internal default client
	client.defaultClient.commandFactory = cap.factory

	line := domain.DialogueLine{Speaker: "Unknown", Text: "Hello world", Emotion: "neutral"}
	got, err := client.GenerateSpeechForLine(context.Background(), line)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) == 0 {
		t.Error("expected audio bytes, got empty")
	}
}

func TestPollyMultiSpeakerClient_EmotionSSML_Angry(t *testing.T) {
	reg := character.NewMockRegistry()

	client := NewPollyMultiSpeakerClient("us-east-1", "ak", "sk", "en-US", reg)

	cap := &captureSSMLFactory{}
	client.defaultClient.commandFactory = cap.factory

	line := domain.DialogueLine{Speaker: "", Text: "I am angry", Emotion: "angry"}
	_, err := client.GenerateSpeechForLine(context.Background(), line)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(cap.ssml, `rate="fast"`) {
		t.Errorf("angry emotion SSML should contain rate=\"fast\", got: %s", cap.ssml)
	}
}

func TestPollyMultiSpeakerClient_EmotionSSML_Sad(t *testing.T) {
	reg := character.NewMockRegistry()

	client := NewPollyMultiSpeakerClient("us-east-1", "ak", "sk", "en-US", reg)

	cap := &captureSSMLFactory{}
	client.defaultClient.commandFactory = cap.factory

	line := domain.DialogueLine{Speaker: "", Text: "I am sad", Emotion: "sad"}
	_, err := client.GenerateSpeechForLine(context.Background(), line)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(cap.ssml, `rate="slow"`) {
		t.Errorf("sad emotion SSML should contain rate=\"slow\", got: %s", cap.ssml)
	}
}

func TestPollyMultiSpeakerClient_EmotionSSML_Happy(t *testing.T) {
	reg := character.NewMockRegistry()

	client := NewPollyMultiSpeakerClient("us-east-1", "ak", "sk", "en-US", reg)

	cap := &captureSSMLFactory{}
	client.defaultClient.commandFactory = cap.factory

	line := domain.DialogueLine{Speaker: "", Text: "I am happy", Emotion: "happy"}
	_, err := client.GenerateSpeechForLine(context.Background(), line)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(cap.ssml, `rate="medium"`) {
		t.Errorf("happy emotion SSML should contain rate=\"medium\", got: %s", cap.ssml)
	}
}

func TestPollyMultiSpeakerClient_EmotionSSML_Whisper(t *testing.T) {
	reg := character.NewMockRegistry()

	client := NewPollyMultiSpeakerClient("us-east-1", "ak", "sk", "en-US", reg)

	cap := &captureSSMLFactory{}
	client.defaultClient.commandFactory = cap.factory

	line := domain.DialogueLine{Speaker: "", Text: "I am whispering", Emotion: "whisper"}
	_, err := client.GenerateSpeechForLine(context.Background(), line)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(cap.ssml, `whispered`) {
		t.Errorf("whisper emotion SSML should use whispered effect, got: %s", cap.ssml)
	}
}

func TestPollyMultiSpeakerClient_WithRegisteredVoice(t *testing.T) {
	reg := character.NewMockRegistry()
	_, _ = reg.Register(context.Background(), "Alice", []byte("img"))
	reg.SetMeta("Alice", &character.CharacterMeta{
		Name:    "Alice",
		VoiceID: "Joanna",
	})

	client := NewPollyMultiSpeakerClient("us-east-1", "ak", "sk", "zh-TW", reg)

	cap := &captureSSMLFactory{}
	// The client should create a dedicated PollyCLIClient for Alice with Joanna voice
	// We need to intercept after the per-speaker client is created.
	// Call once so the client is lazily built, then inject factory.
	line := domain.DialogueLine{Speaker: "Alice", Text: "Hello", Emotion: "neutral"}

	// Set up a pre-injected factory so we can capture before the call
	client.commandFactoryOverride = cap.factory

	got, err := client.GenerateSpeechForLine(context.Background(), line)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) == 0 {
		t.Error("expected audio bytes")
	}
}
