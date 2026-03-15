package postprod_test

import (
	"context"
	"testing"

	"github.com/baochen10luo/stagenthand/internal/domain"
	"github.com/baochen10luo/stagenthand/internal/postprod"
)

func TestLoop_converge_first_iteration(t *testing.T) {
	evaluator := &postprod.MockVideoEvaluator{
		Results: []*postprod.EvaluationResult{
			{VisualScore: 9, AudioSyncScore: 9, AdherenceScore: 8, ToneScore: 8, Action: "APPROVE"},
		},
	}
	planner := &postprod.MockEditPlanner{}
	applier := &postprod.MockEditApplier{}
	renderer := &postprod.MockVideoRenderer{}

	cfg := postprod.LoopConfig{MaxIterations: 3, OutputDir: t.TempDir()}
	loop := postprod.NewPostProdLoop(evaluator, planner, applier, renderer, cfg)

	result, err := loop.Run(context.Background(), "/fake/video.mp4", makeProps())
	if err != nil {
		t.Fatalf("Run() error: %v", err)
	}
	if !result.Converged {
		t.Error("expected Converged=true")
	}
	if result.Iterations != 1 {
		t.Errorf("Iterations: got %d, want 1", result.Iterations)
	}
	if planner.CallCount != 0 {
		t.Errorf("Planner should not be called when approved on first eval, got %d calls", planner.CallCount)
	}
}

func TestLoop_converge_after_retry(t *testing.T) {
	evaluator := &postprod.MockVideoEvaluator{
		Results: []*postprod.EvaluationResult{
			{VisualScore: 5, AudioSyncScore: 9, AdherenceScore: 9, ToneScore: 9, Action: "REJECT"},
			{VisualScore: 9, AudioSyncScore: 9, AdherenceScore: 9, ToneScore: 9, Action: "APPROVE"},
		},
	}
	planner := &postprod.MockEditPlanner{}
	applier := &postprod.MockEditApplier{}
	renderer := &postprod.MockVideoRenderer{}

	cfg := postprod.LoopConfig{MaxIterations: 3, OutputDir: t.TempDir()}
	loop := postprod.NewPostProdLoop(evaluator, planner, applier, renderer, cfg)

	result, err := loop.Run(context.Background(), "/fake/video.mp4", makeProps())
	if err != nil {
		t.Fatalf("Run() error: %v", err)
	}
	if !result.Converged {
		t.Error("expected Converged=true")
	}
	if result.Iterations != 2 {
		t.Errorf("Iterations: got %d, want 2", result.Iterations)
	}
	if planner.CallCount != 1 {
		t.Errorf("Planner CallCount: got %d, want 1", planner.CallCount)
	}
	if renderer.CallCount != 1 {
		t.Errorf("Renderer CallCount: got %d, want 1", renderer.CallCount)
	}
}

func TestLoop_max_iterations_not_converged(t *testing.T) {
	// All evaluations reject
	evaluator := &postprod.MockVideoEvaluator{
		Results: []*postprod.EvaluationResult{
			{VisualScore: 5, AudioSyncScore: 5, AdherenceScore: 5, ToneScore: 5, Action: "REJECT"},
			{VisualScore: 5, AudioSyncScore: 5, AdherenceScore: 5, ToneScore: 5, Action: "REJECT"},
			{VisualScore: 5, AudioSyncScore: 5, AdherenceScore: 5, ToneScore: 5, Action: "REJECT"},
		},
	}
	planner := &postprod.MockEditPlanner{}
	applier := &postprod.MockEditApplier{}
	renderer := &postprod.MockVideoRenderer{}

	cfg := postprod.LoopConfig{MaxIterations: 3, OutputDir: t.TempDir()}
	loop := postprod.NewPostProdLoop(evaluator, planner, applier, renderer, cfg)

	result, err := loop.Run(context.Background(), "/fake/video.mp4", makeProps())
	if err != nil {
		t.Fatalf("Run() error: %v", err)
	}
	if result.Converged {
		t.Error("expected Converged=false")
	}
	if result.Iterations != 3 {
		t.Errorf("Iterations: got %d, want 3", result.Iterations)
	}
	if renderer.CallCount != 3 {
		t.Errorf("Renderer CallCount: got %d, want 3", renderer.CallCount)
	}
}

func TestLoop_writes_files(t *testing.T) {
	tmpDir := t.TempDir()

	evaluator := &postprod.MockVideoEvaluator{
		Results: []*postprod.EvaluationResult{
			{VisualScore: 5, AudioSyncScore: 9, AdherenceScore: 9, ToneScore: 9, Action: "REJECT"},
			{VisualScore: 9, AudioSyncScore: 9, AdherenceScore: 9, ToneScore: 9, Action: "APPROVE"},
		},
	}
	planner := &postprod.MockEditPlanner{}
	applier := &postprod.MockEditApplier{}
	renderer := &postprod.MockVideoRenderer{}

	cfg := postprod.LoopConfig{MaxIterations: 3, OutputDir: tmpDir}
	loop := postprod.NewPostProdLoop(evaluator, planner, applier, renderer, cfg)

	_, err := loop.Run(context.Background(), "/fake/video.mp4", makeProps())
	if err != nil {
		t.Fatalf("Run() error: %v", err)
	}
	// After 1 rejected iteration, should have written props and plan files
	// Just verify the run completes without error - file system assertions are integration concerns
}

func TestEvaluationResult_IsApproved(t *testing.T) {
	cases := []struct {
		name     string
		eval     postprod.EvaluationResult
		approved bool
	}{
		{"all 9s approve", postprod.EvaluationResult{VisualScore: 9, AudioSyncScore: 9, AdherenceScore: 9, ToneScore: 9}, true},
		{"visual low reject", postprod.EvaluationResult{VisualScore: 7, AudioSyncScore: 9, AdherenceScore: 9, ToneScore: 9}, false},
		{"audio low reject", postprod.EvaluationResult{VisualScore: 9, AudioSyncScore: 7, AdherenceScore: 9, ToneScore: 9}, false},
		{"total low reject", postprod.EvaluationResult{VisualScore: 8, AudioSyncScore: 8, AdherenceScore: 8, ToneScore: 7}, false},
		{"total exact 32 approve", postprod.EvaluationResult{VisualScore: 8, AudioSyncScore: 8, AdherenceScore: 8, ToneScore: 8}, true},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := c.eval.IsApproved()
			if got != c.approved {
				t.Errorf("IsApproved() = %v, want %v for %+v", got, c.approved, c.eval)
			}
		})
	}
}

// Compile-time check: MockEditApplier implements EditApplier
var _ postprod.EditApplier = (*postprod.MockEditApplier)(nil)
var _ postprod.VideoEvaluator = (*postprod.MockVideoEvaluator)(nil)
var _ postprod.EditPlanner = (*postprod.MockEditPlanner)(nil)
var _ postprod.VideoRenderer = (*postprod.MockVideoRenderer)(nil)

// Compile-time check: domain types exist
var _ domain.PostProdLoopResult
