package image

import "context"

// MockClient is meant for testing and --dry-run setups.
type MockClient struct {
	GenerateImageFunc func(ctx context.Context, prompt string, characterRefs []string) ([]byte, error)
	CallCount         int
}

func (m *MockClient) GenerateImage(ctx context.Context, prompt string, characterRefs []string) ([]byte, error) {
	m.CallCount++
	if m.GenerateImageFunc != nil {
		return m.GenerateImageFunc(ctx, prompt, characterRefs)
	}
	// Return a tiny 1x1 black GIF as a reasonable dummy image if needed
	return []byte("GIF89a\x01\x00\x01\x00\x80\x00\x00\x00\x00\x00\xff\xff\xff!\xf9\x04\x01\x00\x00\x00\x00,\x00\x00\x00\x00\x01\x00\x01\x00\x00\x02\x01D\x00;"), nil
}
