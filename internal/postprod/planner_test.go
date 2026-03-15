package postprod_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/baochen10luo/stagenthand/internal/domain"
	"github.com/baochen10luo/stagenthand/internal/postprod"
)

func makeProps() domain.RemotionProps {
	return domain.RemotionProps{
		ProjectID: "proj-1",
		Title:     "Test",
		FPS:       24,
		Directives: &domain.Directives{
			StylePrompt:  "cinematic",
			DuckingDepth: 0.5,
		},
		Panels: []domain.Panel{
			{SceneNumber: 1, PanelNumber: 1, Description: "scene1 panel1", DurationSec: 3.0},
			{SceneNumber: 1, PanelNumber: 2, Description: "scene1 panel2", DurationSec: 2.0},
		},
	}
}

func TestLLMEditPlanner_fallback_visual(t *testing.T) {
	mockLLM := &postprod.MockLLMClient{Err: errors.New("llm unavailable")}
	planner := postprod.NewLLMEditPlanner(mockLLM)

	eval := &postprod.EvaluationResult{
		VisualScore:    5,
		AudioSyncScore: 9,
		AdherenceScore: 9,
		ToneScore:      9,
		Action:         "REJECT",
	}

	plan, err := planner.Plan(context.Background(), eval, makeProps())
	if err != nil {
		t.Fatalf("Plan() error: %v", err)
	}
	if plan == nil {
		t.Fatal("Plan() returned nil")
	}

	// Should have patch_global_directive for style with Priority=1
	var found bool
	for _, op := range plan.Operations {
		if op.Type == domain.EditOpPatchGlobalDirective && op.Priority == 1 {
			if sp, ok := op.Params["style_prompt"].(string); ok && sp != "" {
				found = true
			}
		}
	}
	if !found {
		t.Errorf("expected patch_global_directive with style_prompt and priority=1, got ops: %+v", plan.Operations)
	}
}

func TestLLMEditPlanner_fallback_audio(t *testing.T) {
	mockLLM := &postprod.MockLLMClient{Err: errors.New("llm unavailable")}
	planner := postprod.NewLLMEditPlanner(mockLLM)

	eval := &postprod.EvaluationResult{
		VisualScore:    9,
		AudioSyncScore: 5,
		AdherenceScore: 9,
		ToneScore:      9,
		Action:         "REJECT",
	}

	plan, err := planner.Plan(context.Background(), eval, makeProps())
	if err != nil {
		t.Fatalf("Plan() error: %v", err)
	}

	// Should have patch_global_directive adjusting ducking_depth with Priority=1
	var found bool
	for _, op := range plan.Operations {
		if op.Type == domain.EditOpPatchGlobalDirective && op.Priority == 1 {
			if _, ok := op.Params["ducking_depth"]; ok {
				found = true
			}
		}
	}
	if !found {
		t.Errorf("expected patch_global_directive with ducking_depth and priority=1, got ops: %+v", plan.Operations)
	}
}

func TestLLMEditPlanner_fallback_tone(t *testing.T) {
	mockLLM := &postprod.MockLLMClient{Err: errors.New("llm unavailable")}
	planner := postprod.NewLLMEditPlanner(mockLLM)

	eval := &postprod.EvaluationResult{
		VisualScore:    9,
		AudioSyncScore: 9,
		AdherenceScore: 9,
		ToneScore:      5,
		Action:         "REJECT",
	}

	plan, err := planner.Plan(context.Background(), eval, makeProps())
	if err != nil {
		t.Fatalf("Plan() error: %v", err)
	}

	var found bool
	for _, op := range plan.Operations {
		if op.Type == domain.EditOpPatchDuration && op.Priority == 2 {
			found = true
		}
	}
	if !found {
		t.Errorf("expected patch_duration with priority=2, got ops: %+v", plan.Operations)
	}
}

func TestLLMEditPlanner_llm_success(t *testing.T) {
	planJSON := `{
		"version": "v2",
		"generated_at": "2026-03-15T00:00:00Z",
		"operations": [
			{
				"type": "patch_dialogue",
				"target_panel": {"scene_number": 1, "panel_number": 1},
				"params": {"dialogue": "Better line"},
				"priority": 1,
				"rationale": "improve tone"
			}
		],
		"estimated_cost_usd": 0.01,
		"rationale": "LLM plan"
	}`

	mockLLM := &postprod.MockLLMClient{Response: []byte(planJSON)}
	planner := postprod.NewLLMEditPlanner(mockLLM)

	eval := &postprod.EvaluationResult{
		VisualScore:    6,
		AudioSyncScore: 9,
		AdherenceScore: 9,
		ToneScore:      9,
		Action:         "REJECT",
	}

	plan, err := planner.Plan(context.Background(), eval, makeProps())
	if err != nil {
		t.Fatalf("Plan() error: %v", err)
	}
	if plan.Version != "v2" {
		t.Errorf("Version: got %q, want v2", plan.Version)
	}
	if len(plan.Operations) != 1 {
		t.Fatalf("Operations len: got %d, want 1", len(plan.Operations))
	}
	if plan.Operations[0].Type != domain.EditOpPatchDialogue {
		t.Errorf("Op type: got %q, want patch_dialogue", plan.Operations[0].Type)
	}
	_ = time.Time{} // suppress import
}
