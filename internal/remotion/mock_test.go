package remotion_test

import (
	"context"
	"errors"
	"testing"

	"github.com/baochen10luo/stagenthand/internal/remotion"
)

func TestMockExecutor(t *testing.T) {
	errMock := errors.New("mock error")
	m := &remotion.MockExecutor{MockErr: errMock}

	err := m.Render(context.Background(), "", "", "", "")
	if err != errMock {
		t.Errorf("Expected mock error, got %v", err)
	}
	if m.RenderCalls != 1 {
		t.Errorf("Expected 1 render call, got %d", m.RenderCalls)
	}

	err = m.Preview(context.Background(), "", "", "")
	if err != errMock {
		t.Errorf("Expected mock error, got %v", err)
	}
	if m.PreviewCalls != 1 {
		t.Errorf("Expected 1 preview call, got %d", m.PreviewCalls)
	}
}
