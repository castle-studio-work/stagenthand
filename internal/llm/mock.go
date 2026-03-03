package llm

import "context"

// MockClient is a manually crafted mock for TDD.
type MockClient struct {
	GenerateFunc func(ctx context.Context, systemPrompt string, inputData []byte) ([]byte, error)
	CallCount    int
}

// GenerateTransformation implements the Client interface.
func (m *MockClient) GenerateTransformation(ctx context.Context, systemPrompt string, inputData []byte) ([]byte, error) {
	m.CallCount++
	if m.GenerateFunc != nil {
		return m.GenerateFunc(ctx, systemPrompt, inputData)
	}
	return nil, nil
}
