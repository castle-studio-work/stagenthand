package remotion

import "context"

type MockExecutor struct {
	RenderCalls  int
	PreviewCalls int
	MockErr      error
}

func (m *MockExecutor) Render(ctx context.Context, templatePath string, composition string, propsPath string, outputPath string) error {
	m.RenderCalls++
	return m.MockErr
}

func (m *MockExecutor) Preview(ctx context.Context, templatePath string, composition string, propsPath string) error {
	m.PreviewCalls++
	return m.MockErr
}
