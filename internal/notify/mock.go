package notify

import "context"

// MockNotifier provides a test double for the Notifier interface.
type MockNotifier struct {
	CapturedCalls []struct {
		Title   string
		Message string
		Color   int
	}
	MockErr error
}

// Notify implementations records the call and returns MockErr.
func (m *MockNotifier) Notify(ctx context.Context, title string, message string, color int) error {
	m.CapturedCalls = append(m.CapturedCalls, struct {
		Title   string
		Message string
		Color   int
	}{
		Title:   title,
		Message: message,
		Color:   color,
	})
	return m.MockErr
}
