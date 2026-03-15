package audio

import (
	"context"
	"os"
	"os/exec"
	"testing"
)

func TestFormatSSML(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"Basic text", "Hello", "<speak><prosody rate=\"90%\">Hello</prosody></speak>"},
		{"Whisper tag", "Whisper: quiet please", "<speak><prosody rate=\"90%\"><amazon:effect name=\"whispered\">quiet please</amazon:effect></prosody></speak>"},
		{"Character prefix", "Narrator: Once upon a time", "<speak><prosody rate=\"90%\">Once upon a time</prosody></speak>"},
		{"Stage directions", "Go away [shouting]", "<speak><prosody rate=\"90%\">Go away</prosody></speak>"},
		{"Scrub quotes", "\"Yes\", he said", "<speak><prosody rate=\"90%\">Yes, he said</prosody></speak>"},
		{"XML escape", "Tom & Jerry", "<speak><prosody rate=\"90%\">Tom &amp; Jerry</prosody></speak>"},
		{"Empty string", "", "<speak></speak>"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatSSML(tt.input)
			if got != tt.expected {
				t.Errorf("formatSSML(%q) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}

func TestPollyCLIClient_GenerateSpeech_Basic(t *testing.T) {
	fakeOutput := []byte("fake-mp3-content")
	c := NewPollyCLIClient("us-east-1", "ak", "sk")
	
	// Mock commandFactory to succeed and produce a fake file
	c.commandFactory = func(ctx context.Context, name string, args ...string) *exec.Cmd {
		// The last arg is the output file path in GenerateSpeech
		outPath := args[len(args)-1]
		os.WriteFile(outPath, fakeOutput, 0644)
		// Return a command that does nothing (true)
		return exec.Command("true")
	}

	got, err := c.GenerateSpeech(context.Background(), "Hello")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(got) != string(fakeOutput) {
		t.Errorf("GenerateSpeech() = %q, want %q", string(got), string(fakeOutput))
	}
}

func TestPollyCLIClient_GenerateSpeech_Empty(t *testing.T) {
	c := NewPollyCLIClient("us-east-1", "ak", "sk")
	got, err := c.GenerateSpeech(context.Background(), "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != nil {
		t.Errorf("expected nil for empty text, got %v", got)
	}
}

func TestPollyCLIClient_GenerateSpeech_Error(t *testing.T) {
	c := NewPollyCLIClient("us-east-1", "ak", "sk")
	c.commandFactory = func(ctx context.Context, name string, args ...string) *exec.Cmd {
		// Return a command that fails
		return exec.Command("false")
	}

	_, err := c.GenerateSpeech(context.Background(), "Fail")
	if err == nil {
		t.Fatal("expected error from failed command, got nil")
	}
}
