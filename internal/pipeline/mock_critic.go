package pipeline

import (
	"context"
	"fmt"
)

// MockVideoCriticEvaluator is a test double for VideoCriticEvaluator.
// It returns Results in order, cycling through them per call.
type MockVideoCriticEvaluator struct {
	Results []*CriticResult
	callIdx int
}

// Evaluate returns the next result in the Results slice.
// Returns an error if Results is empty.
func (m *MockVideoCriticEvaluator) Evaluate(_ context.Context, _ string, _ []byte) (*CriticResult, error) {
	if len(m.Results) == 0 {
		return nil, fmt.Errorf("MockVideoCriticEvaluator: no results configured")
	}
	idx := m.callIdx
	if idx >= len(m.Results) {
		idx = len(m.Results) - 1
	}
	m.callIdx++
	return m.Results[idx], nil
}
