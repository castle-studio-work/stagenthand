package video

import (
	"context"
)

// MockClient implements Client interface for testing.
type MockClient struct {
	CapturedURL    string
	CapturedPrompt string
	MockVideoBytes []byte
	MockErr        error
}

func (m *MockClient) GenerateVideo(ctx context.Context, imageURL string, prompt string) ([]byte, error) {
	m.CapturedURL = imageURL
	m.CapturedPrompt = prompt
	return m.MockVideoBytes, m.MockErr
}
