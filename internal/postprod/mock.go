package postprod

import (
	"context"
	"time"

	"github.com/baochen10luo/stagenthand/internal/domain"
)

// MockLLMClient is a test double for the llm.Client interface.
type MockLLMClient struct {
	Response []byte
	Err      error
}

func (m *MockLLMClient) GenerateTransformation(_ context.Context, _ string, _ []byte) ([]byte, error) {
	return m.Response, m.Err
}

// MockVideoEvaluator cycles through a predefined list of EvaluationResults.
type MockVideoEvaluator struct {
	Results []*EvaluationResult
	index   int
}

func (m *MockVideoEvaluator) Evaluate(_ context.Context, _ string, _ []byte) (*EvaluationResult, error) {
	if m.index >= len(m.Results) {
		// Return last result repeatedly if we run out
		return m.Results[len(m.Results)-1], nil
	}
	r := m.Results[m.index]
	m.index++
	return r, nil
}

// MockEditPlanner returns a minimal EditPlan without calling an LLM.
type MockEditPlanner struct {
	CallCount    int
	FixedPlan    *domain.EditPlan
}

func (m *MockEditPlanner) Plan(_ context.Context, _ *EvaluationResult, _ domain.RemotionProps) (*domain.EditPlan, error) {
	m.CallCount++
	if m.FixedPlan != nil {
		return m.FixedPlan, nil
	}
	return &domain.EditPlan{
		Version:     "mock-v1",
		GeneratedAt: time.Now(),
		Operations:  []domain.EditOperation{},
		Rationale:   "mock plan",
	}, nil
}

// MockEditApplier returns the input props unchanged and records call count.
type MockEditApplier struct {
	CallCount int
}

func (m *MockEditApplier) Apply(_ context.Context, plan *domain.EditPlan, props domain.RemotionProps) (*domain.EditResult, error) {
	m.CallCount++
	return &domain.EditResult{
		PlanVersion:       plan.Version,
		OperationsApplied: len(plan.Operations),
		OperationsFailed:  0,
		UpdatedProps:      props,
		Success:           true,
	}, nil
}

// MockVideoRenderer records call count without rendering anything.
type MockVideoRenderer struct {
	CallCount int
}

func (m *MockVideoRenderer) Render(_ context.Context, _ []byte, _ string) error {
	m.CallCount++
	return nil
}

// MockPropsEvaluator is a test double for the PropsEvaluator interface.
type MockPropsEvaluator struct {
	Result    *PropsEvaluation
	Err       error
	CallCount int
}

func (m *MockPropsEvaluator) Evaluate(_ context.Context, _ []byte) (*PropsEvaluation, error) {
	m.CallCount++
	return m.Result, m.Err
}
