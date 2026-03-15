package pipeline_test

import (
	"context"
	"errors"
	"testing"

	"github.com/baochen10luo/stagenthand/internal/domain"
	"github.com/baochen10luo/stagenthand/internal/pipeline"
)

// --- Mock implementations ---

type mockTransformer struct {
	output       []byte
	err          error
	GenerateFunc func(ctx context.Context, systemPrompt string, inputData []byte) ([]byte, error)
}

func (m *mockTransformer) GenerateTransformation(ctx context.Context, systemPrompt string, inputData []byte) ([]byte, error) {
	if m.GenerateFunc != nil {
		return m.GenerateFunc(ctx, systemPrompt, inputData)
	}
	return m.output, m.err
}

type mockImageBatcher struct {
	called bool
	err    error
}

func (m *mockImageBatcher) BatchGenerateImages(_ context.Context, panels []domain.Panel, _ string) ([]domain.Panel, error) {
	m.called = true
	if m.err != nil {
		return nil, m.err
	}
	// mark each panel as having an image
	result := make([]domain.Panel, len(panels))
	for i, p := range panels {
		p.ImageURL = "https://example.com/generated_" + p.Description + ".png"
		result[i] = p
	}
	return result, nil
}

type mockCheckpointStore struct {
	approved bool
}

func (m *mockCheckpointStore) CreateAndWait(_ context.Context, _ string, _ domain.CheckpointStage) error {
	if !m.approved {
		return errors.New("checkpoint rejected")
	}
	return nil
}

// --- Tests ---

func TestOrchestrator_RunStagesInOrder(t *testing.T) {
	outlineJSON := []byte(`{"project_id":"p1","episodes":[{"number":1,"title":"Ep1","synopsis":"s","hook":"h","cliffhanger":"c"}]}`)
	storyboardJSON := []byte(`{"project_id":"p1","episode":1,"scenes":[{"number":1,"description":"scene"}]}`)
	panelsJSON := []byte(`{"panels":[{"scene_number":1,"panel_number":1,"description":"hero","dialogue":"Hello","character_refs":[],"duration_sec":3.0}]}`)

	orch := pipeline.NewOrchestrator(pipeline.OrchestratorDeps{
		LLM: &mockTransformer{
			GenerateFunc: func(_ context.Context, systemPrompt string, _ []byte) ([]byte, error) {
				if systemPrompt == pipeline.PromptOutlineToStoryboard {
					return storyboardJSON, nil
				}
				if systemPrompt == pipeline.PromptStoryboardToPanels {
					return panelsJSON, nil
				}
				return nil, errors.New("unexpected prompt")
			},
		},
		Images:      &mockImageBatcher{},
		Checkpoints: &mockCheckpointStore{approved: true},
		DryRun:      false,
		SkipHITL:    true,
	})

	ctx := context.Background()
	result, err := orch.Run(ctx, outlineJSON)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
}

func TestOrchestrator_DryRunSkipsImages(t *testing.T) {
	storyboardJSON := []byte(`{"project_id":"dry","episode":1,"scenes":[{"number":1,"description":"s"}]}`)
	panelsJSON := []byte(`{"panels":[{"scene_number":1,"panel_number":1,"description":"p"}]}`)

	imgBatcher := &mockImageBatcher{}
	orch := pipeline.NewOrchestrator(pipeline.OrchestratorDeps{
		LLM:         &mockTransformer{output: panelsJSON},
		Images:      imgBatcher,
		Checkpoints: &mockCheckpointStore{approved: true},
		DryRun:      true,
		SkipHITL:    true,
	})

	_, err := orch.Run(context.Background(), storyboardJSON)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if imgBatcher.called {
		t.Error("dry-run: image generator should NOT be called")
	}
}

func TestOrchestrator_HITLRejectionAborts(t *testing.T) {
	storyboardJSON := []byte(`{"project_id":"hitl","episode":1,"scenes":[{"number":1,"description":"s"}]}`)
	panelsJSON := []byte(`{"panels":[]}`)

	orch := pipeline.NewOrchestrator(pipeline.OrchestratorDeps{
		LLM:         &mockTransformer{output: panelsJSON},
		Images:      &mockImageBatcher{},
		Checkpoints: &mockCheckpointStore{approved: false}, // rejects
		DryRun:      false,
		SkipHITL:    false,
	})

	_, err := orch.Run(context.Background(), storyboardJSON)
	if err == nil {
		t.Error("expected error when checkpoint is rejected, got nil")
	}
}

func TestOrchestrator_LLMFailurePropagates(t *testing.T) {
	orch := pipeline.NewOrchestrator(pipeline.OrchestratorDeps{
		LLM:         &mockTransformer{err: errors.New("LLM quota exceeded")},
		Images:      &mockImageBatcher{},
		Checkpoints: &mockCheckpointStore{approved: true},
		DryRun:      false,
		SkipHITL:    true,
	})

	_, err := orch.Run(context.Background(), []byte(`{"story":"test"}`))
	if err == nil {
		t.Error("expected LLM error to propagate, got nil")
	}
}

// --- Critic retry tests ---

func makePanelsInput() []byte {
	return []byte(`{"panels":[{"scene_number":1,"panel_number":1,"description":"hero","dialogue":"Hi","duration_sec":3.0}]}`)
}

func makeApproveResult() *pipeline.CriticResult {
	return &pipeline.CriticResult{VisualScore: 9, AudioSyncScore: 9, AdherenceScore: 8, ToneScore: 8, Action: "APPROVE"}
}

func makeRejectResult(visualScore int) *pipeline.CriticResult {
	return &pipeline.CriticResult{VisualScore: visualScore, AudioSyncScore: 9, AdherenceScore: 7, ToneScore: 7, Action: "REJECT", Feedback: "needs work"}
}

func TestCriticRetry_approve_on_first(t *testing.T) {
	mockCritic := &pipeline.MockVideoCriticEvaluator{
		Results: []*pipeline.CriticResult{makeApproveResult()},
	}

	orch := pipeline.NewOrchestrator(pipeline.OrchestratorDeps{
		LLM:         &mockTransformer{output: []byte(`{"panels":[]}`)},
		Images:      &mockImageBatcher{},
		Checkpoints: &mockCheckpointStore{approved: true},
		DryRun:      true,
		SkipHITL:    true,
		Critic:      mockCritic,
		MaxRetries:  2,
		VideoPath:   "/tmp/fake.mp4",
	})

	result, err := orch.Run(context.Background(), makePanelsInput())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.CriticAttempts != 1 {
		t.Errorf("CriticAttempts = %d, want 1", result.CriticAttempts)
	}
	if !result.CriticApproved {
		t.Error("expected CriticApproved = true")
	}
}

func TestCriticRetry_retry_then_approve(t *testing.T) {
	mockCritic := &pipeline.MockVideoCriticEvaluator{
		Results: []*pipeline.CriticResult{
			makeRejectResult(5),   // first call: REJECT (visual=5)
			makeApproveResult(),   // second call: APPROVE
		},
	}

	orch := pipeline.NewOrchestrator(pipeline.OrchestratorDeps{
		LLM:         &mockTransformer{output: []byte(`{"panels":[]}`)},
		Images:      &mockImageBatcher{},
		Checkpoints: &mockCheckpointStore{approved: true},
		DryRun:      true,
		SkipHITL:    true,
		Critic:      mockCritic,
		MaxRetries:  2,
		VideoPath:   "/tmp/fake.mp4",
	})

	result, err := orch.Run(context.Background(), makePanelsInput())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.CriticAttempts != 2 {
		t.Errorf("CriticAttempts = %d, want 2", result.CriticAttempts)
	}
	if !result.CriticApproved {
		t.Error("expected CriticApproved = true")
	}
	// StylePrompt should have been prepended after visual score < 8
	directives := result.Storyboard.Directives
	if directives == nil || directives.StylePrompt == "" {
		t.Error("expected StylePrompt to be modified after visual score < 8")
	}
}

func TestCriticRetry_all_fail(t *testing.T) {
	mockCritic := &pipeline.MockVideoCriticEvaluator{
		Results: []*pipeline.CriticResult{
			makeRejectResult(5),
			makeRejectResult(6),
			makeRejectResult(7),
		},
	}

	orch := pipeline.NewOrchestrator(pipeline.OrchestratorDeps{
		LLM:         &mockTransformer{output: []byte(`{"panels":[]}`)},
		Images:      &mockImageBatcher{},
		Checkpoints: &mockCheckpointStore{approved: true},
		DryRun:      true,
		SkipHITL:    true,
		Critic:      mockCritic,
		MaxRetries:  2,
		VideoPath:   "/tmp/fake.mp4",
	})

	result, err := orch.Run(context.Background(), makePanelsInput())
	if err != nil {
		t.Fatalf("unexpected error (should return result even when all rejected): %v", err)
	}
	if result.CriticApproved {
		t.Error("expected CriticApproved = false when all attempts fail")
	}
	if result.CriticAttempts != 3 {
		t.Errorf("CriticAttempts = %d, want 3", result.CriticAttempts)
	}
}
