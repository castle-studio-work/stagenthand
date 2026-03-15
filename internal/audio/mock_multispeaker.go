package audio

import (
	"context"

	"github.com/baochen10luo/stagenthand/internal/domain"
)

// MockMultiSpeakerClient is a test double for MultiSpeakerClient.
type MockMultiSpeakerClient struct {
	Calls []MultiSpeakerCall
	Data  []byte
	Err   error
}

// MultiSpeakerCall records a single call to GenerateSpeechForLine.
type MultiSpeakerCall struct {
	Line domain.DialogueLine
}

// GenerateSpeechForLine records the call and returns mock data.
func (m *MockMultiSpeakerClient) GenerateSpeechForLine(_ context.Context, line domain.DialogueLine) ([]byte, error) {
	m.Calls = append(m.Calls, MultiSpeakerCall{Line: line})
	if m.Err != nil {
		return nil, m.Err
	}
	if m.Data != nil {
		return m.Data, nil
	}
	return []byte("mock-audio"), nil
}
