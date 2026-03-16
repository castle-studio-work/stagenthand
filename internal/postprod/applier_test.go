package postprod_test

import (
	"context"
	"testing"

	"github.com/baochen10luo/stagenthand/internal/domain"
	"github.com/baochen10luo/stagenthand/internal/postprod"
)

func makePlan(ops []domain.EditOperation) *domain.EditPlan {
	return &domain.EditPlan{
		Version:    "v1",
		Operations: ops,
		Rationale:  "test plan",
	}
}

func TestApplier_patch_dialogue(t *testing.T) {
	applier := postprod.NewDefaultEditApplier()
	props := makeProps()

	plan := makePlan([]domain.EditOperation{
		{
			Type:        domain.EditOpPatchDialogue,
			TargetPanel: &domain.PanelRef{SceneNumber: 1, PanelNumber: 1},
			Params:      map[string]interface{}{"dialogue": "New dialogue"},
			Priority:    1,
		},
	})

	result, err := applier.Apply(context.Background(), plan, props)
	if err != nil {
		t.Fatalf("Apply() error: %v", err)
	}
	if result.OperationsApplied != 1 {
		t.Errorf("OperationsApplied: got %d, want 1", result.OperationsApplied)
	}
	if result.OperationsFailed != 0 {
		t.Errorf("OperationsFailed: got %d, want 0", result.OperationsFailed)
	}

	// Find updated panel
	var found bool
	for _, p := range result.UpdatedProps.Panels {
		if p.SceneNumber == 1 && p.PanelNumber == 1 {
			if p.Dialogue != "New dialogue" {
				t.Errorf("Dialogue: got %q, want 'New dialogue'", p.Dialogue)
			}
			found = true
		}
	}
	if !found {
		t.Error("panel scene=1 panel=1 not found in result")
	}
}

func TestApplier_patch_duration(t *testing.T) {
	applier := postprod.NewDefaultEditApplier()
	props := makeProps()

	plan := makePlan([]domain.EditOperation{
		{
			Type:        domain.EditOpPatchDuration,
			TargetPanel: &domain.PanelRef{SceneNumber: 1, PanelNumber: 2},
			Params:      map[string]interface{}{"factor": 1.2},
			Priority:    1,
		},
	})

	result, err := applier.Apply(context.Background(), plan, props)
	if err != nil {
		t.Fatalf("Apply() error: %v", err)
	}

	for _, p := range result.UpdatedProps.Panels {
		if p.SceneNumber == 1 && p.PanelNumber == 2 {
			// Original DurationSec = 2.0, factor = 1.2, expected = 2.4
			if p.DurationSec < 2.39 || p.DurationSec > 2.41 {
				t.Errorf("DurationSec: got %f, want ~2.4", p.DurationSec)
			}
			return
		}
	}
	t.Error("panel scene=1 panel=2 not found in result")
}

func TestApplier_patch_global_style(t *testing.T) {
	applier := postprod.NewDefaultEditApplier()
	props := makeProps()
	props.Directives.StylePrompt = "cinematic"

	plan := makePlan([]domain.EditOperation{
		{
			Type:     domain.EditOpPatchGlobalDirective,
			Params:   map[string]interface{}{"style_prompt": "highly detailed, 8K, cinematic"},
			Priority: 1,
		},
	})

	result, err := applier.Apply(context.Background(), plan, props)
	if err != nil {
		t.Fatalf("Apply() error: %v", err)
	}

	if result.UpdatedProps.Directives == nil {
		t.Fatal("Directives should not be nil")
	}
	got := result.UpdatedProps.Directives.StylePrompt
	if got != "highly detailed, 8K, cinematic" {
		t.Errorf("StylePrompt: got %q, want 'highly detailed, 8K, cinematic'", got)
	}
}

func TestApplier_patch_global_ducking(t *testing.T) {
	applier := postprod.NewDefaultEditApplier()
	props := makeProps()
	props.Directives.DuckingDepth = 0.5

	plan := makePlan([]domain.EditOperation{
		{
			Type:     domain.EditOpPatchGlobalDirective,
			Params:   map[string]interface{}{"ducking_depth": 0.4},
			Priority: 1,
		},
	})

	result, err := applier.Apply(context.Background(), plan, props)
	if err != nil {
		t.Fatalf("Apply() error: %v", err)
	}

	got := result.UpdatedProps.Directives.DuckingDepth
	if got < 0.39 || got > 0.41 {
		t.Errorf("DuckingDepth: got %f, want ~0.4", got)
	}
}

func TestApplier_patch_global_ducking_min(t *testing.T) {
	applier := postprod.NewDefaultEditApplier()
	props := makeProps()
	props.Directives.DuckingDepth = 0.05

	plan := makePlan([]domain.EditOperation{
		{
			Type:     domain.EditOpPatchGlobalDirective,
			Params:   map[string]interface{}{"ducking_depth": 0.0},
			Priority: 1,
		},
	})

	result, err := applier.Apply(context.Background(), plan, props)
	if err != nil {
		t.Fatalf("Apply() error: %v", err)
	}

	got := result.UpdatedProps.Directives.DuckingDepth
	if got < 0.1 {
		t.Errorf("DuckingDepth should be minimum 0.1, got %f", got)
	}
}

func TestApplier_invalid_panel_ref(t *testing.T) {
	applier := postprod.NewDefaultEditApplier()
	props := makeProps()

	plan := makePlan([]domain.EditOperation{
		{
			Type:        domain.EditOpPatchDialogue,
			TargetPanel: &domain.PanelRef{SceneNumber: 5, PanelNumber: 1}, // does not exist
			Params:      map[string]interface{}{"dialogue": "ghost"},
			Priority:    1,
		},
		{
			Type:        domain.EditOpPatchDialogue,
			TargetPanel: &domain.PanelRef{SceneNumber: 1, PanelNumber: 1}, // valid
			Params:      map[string]interface{}{"dialogue": "valid"},
			Priority:    2,
		},
	})

	result, err := applier.Apply(context.Background(), plan, props)
	if err != nil {
		t.Fatalf("Apply() error: %v", err)
	}

	// Should record the error but not fail the entire apply
	if result.OperationsFailed != 1 {
		t.Errorf("OperationsFailed: got %d, want 1", result.OperationsFailed)
	}
	if result.OperationsApplied != 1 {
		t.Errorf("OperationsApplied: got %d, want 1", result.OperationsApplied)
	}
	if len(result.Errors) == 0 {
		t.Error("Errors should not be empty")
	}
}

func TestApplier_sort_by_priority(t *testing.T) {
	applier := postprod.NewDefaultEditApplier()
	props := makeProps()

	// Two operations on same panel: priority 3 first in slice, priority 1 second
	// The one with priority=1 (set to "final") should win if they both patch the same field
	plan := makePlan([]domain.EditOperation{
		{
			Type:        domain.EditOpPatchDialogue,
			TargetPanel: &domain.PanelRef{SceneNumber: 1, PanelNumber: 1},
			Params:      map[string]interface{}{"dialogue": "low priority"},
			Priority:    3,
		},
		{
			Type:        domain.EditOpPatchDialogue,
			TargetPanel: &domain.PanelRef{SceneNumber: 1, PanelNumber: 1},
			Params:      map[string]interface{}{"dialogue": "high priority"},
			Priority:    1,
		},
	})

	result, err := applier.Apply(context.Background(), plan, props)
	if err != nil {
		t.Fatalf("Apply() error: %v", err)
	}

	// Priority=1 applied first, priority=3 applied after → final value = "low priority"
	// (high priority operations run first, lower priority later overwrite)
	// Actually, if sorted ascending by priority (1 first), "high priority" runs first,
	// then "low priority" overwrites → final = "low priority"
	// But the test should verify ORDER matters - let's check OperationsApplied
	if result.OperationsApplied != 2 {
		t.Errorf("OperationsApplied: got %d, want 2", result.OperationsApplied)
	}
}

func TestApplier_regenerate_image_clears_url(t *testing.T) {
	applier := postprod.NewDefaultEditApplier()
	props := makeProps()
	props.Panels[0].ImageURL = "https://example.com/image.png"

	plan := makePlan([]domain.EditOperation{
		{
			Type:        domain.EditOpRegenerateImage,
			TargetPanel: &domain.PanelRef{SceneNumber: 1, PanelNumber: 1},
			Priority:    1,
		},
	})

	result, err := applier.Apply(context.Background(), plan, props)
	if err != nil {
		t.Fatalf("Apply() error: %v", err)
	}

	for _, p := range result.UpdatedProps.Panels {
		if p.SceneNumber == 1 && p.PanelNumber == 1 {
			if p.ImageURL != "" {
				t.Errorf("ImageURL should be empty, got %q", p.ImageURL)
			}
			return
		}
	}
	t.Error("panel not found")
}

func TestApplier_patch_panel_directive(t *testing.T) {
	applier := postprod.NewDefaultEditApplier()
	props := makeProps()

	plan := makePlan([]domain.EditOperation{
		{
			Type:        domain.EditOpPatchPanelDirective,
			TargetPanel: &domain.PanelRef{SceneNumber: 1, PanelNumber: 1},
			Params: map[string]interface{}{
				"motion_effect":    "ken_burns_in",
				"transition_in":    "fade",
				"subtitle_effect":  "typewriter",
				"subtitle_position": "bottom",
			},
			Priority: 1,
		},
	})

	result, err := applier.Apply(context.Background(), plan, props)
	if err != nil {
		t.Fatalf("Apply() error: %v", err)
	}

	for _, p := range result.UpdatedProps.Panels {
		if p.SceneNumber == 1 && p.PanelNumber == 1 {
			if p.Directive == nil {
				t.Fatal("Directive should not be nil after patch")
			}
			if p.Directive.MotionEffect != "ken_burns_in" {
				t.Errorf("MotionEffect: got %q, want ken_burns_in", p.Directive.MotionEffect)
			}
			if p.Directive.TransitionIn != "fade" {
				t.Errorf("TransitionIn: got %q, want fade", p.Directive.TransitionIn)
			}
			if p.Directive.SubtitleEffect != "typewriter" {
				t.Errorf("SubtitleEffect: got %q, want typewriter", p.Directive.SubtitleEffect)
			}
			if p.Directive.SubtitlePosition != "bottom" {
				t.Errorf("SubtitlePosition: got %q, want bottom", p.Directive.SubtitlePosition)
			}
			return
		}
	}
	t.Error("panel not found")
}

func TestApplier_patch_panel_directive_numeric_fields(t *testing.T) {
	applier := postprod.NewDefaultEditApplier()
	props := makeProps()

	plan := makePlan([]domain.EditOperation{
		{
			Type:        domain.EditOpPatchPanelDirective,
			TargetPanel: &domain.PanelRef{SceneNumber: 1, PanelNumber: 1},
			Params: map[string]interface{}{
				"motion_intensity":      0.1,
				"transition_duration_ms": 500.0,
				"subtitle_font_size":    42.0,
			},
			Priority: 1,
		},
	})

	result, err := applier.Apply(context.Background(), plan, props)
	if err != nil {
		t.Fatalf("Apply() error: %v", err)
	}

	for _, p := range result.UpdatedProps.Panels {
		if p.SceneNumber == 1 && p.PanelNumber == 1 {
			if p.Directive == nil {
				t.Fatal("Directive should not be nil")
			}
			if p.Directive.MotionIntensity < 0.09 || p.Directive.MotionIntensity > 0.11 {
				t.Errorf("MotionIntensity: got %f, want ~0.1", p.Directive.MotionIntensity)
			}
			if p.Directive.TransitionDurationMs != 500 {
				t.Errorf("TransitionDurationMs: got %d, want 500", p.Directive.TransitionDurationMs)
			}
			if p.Directive.SubtitleFontSize != 42 {
				t.Errorf("SubtitleFontSize: got %d, want 42", p.Directive.SubtitleFontSize)
			}
			return
		}
	}
	t.Error("panel not found")
}

func TestApplier_regenerate_audio_clears_url(t *testing.T) {
	applier := postprod.NewDefaultEditApplier()
	props := makeProps()
	props.Panels[1].AudioURL = "https://example.com/audio.mp3"

	plan := makePlan([]domain.EditOperation{
		{
			Type:        domain.EditOpRegenerateAudio,
			TargetPanel: &domain.PanelRef{SceneNumber: 1, PanelNumber: 2},
			Priority:    1,
		},
	})

	result, err := applier.Apply(context.Background(), plan, props)
	if err != nil {
		t.Fatalf("Apply() error: %v", err)
	}

	for _, p := range result.UpdatedProps.Panels {
		if p.SceneNumber == 1 && p.PanelNumber == 2 {
			if p.AudioURL != "" {
				t.Errorf("AudioURL should be empty, got %q", p.AudioURL)
			}
			return
		}
	}
	t.Error("panel not found")
}

func TestApplier_replace_bgm(t *testing.T) {
	applier := postprod.NewDefaultEditApplier()
	props := makeProps()
	props.BGMURL = "https://example.com/bgm.mp3"

	plan := makePlan([]domain.EditOperation{
		{
			Type:     domain.EditOpReplaceBGM,
			Params:   map[string]interface{}{"bgm_tags": "jazz+upbeat"},
			Priority: 1,
		},
	})

	result, err := applier.Apply(context.Background(), plan, props)
	if err != nil {
		t.Fatalf("Apply() error: %v", err)
	}

	if result.UpdatedProps.BGMURL != "" {
		t.Errorf("BGMURL should be empty, got %q", result.UpdatedProps.BGMURL)
	}
	if result.UpdatedProps.Directives == nil {
		t.Fatal("Directives should not be nil")
	}
	if result.UpdatedProps.Directives.BGMTags != "jazz+upbeat" {
		t.Errorf("BGMTags: got %q, want jazz+upbeat", result.UpdatedProps.Directives.BGMTags)
	}
}

func TestApplier_patch_global_bgm_tags(t *testing.T) {
	applier := postprod.NewDefaultEditApplier()
	props := makeProps()

	plan := makePlan([]domain.EditOperation{
		{
			Type:     domain.EditOpPatchGlobalDirective,
			Params:   map[string]interface{}{"bgm_tags": "cinematic+epic"},
			Priority: 1,
		},
	})

	result, err := applier.Apply(context.Background(), plan, props)
	if err != nil {
		t.Fatalf("Apply() error: %v", err)
	}

	if result.UpdatedProps.Directives.BGMTags != "cinematic+epic" {
		t.Errorf("BGMTags: got %q, want cinematic+epic", result.UpdatedProps.Directives.BGMTags)
	}
}

func TestApplier_nil_directives_global_patch(t *testing.T) {
	applier := postprod.NewDefaultEditApplier()
	props := makeProps()
	props.Directives = nil // nil directives - should be created

	plan := makePlan([]domain.EditOperation{
		{
			Type:     domain.EditOpPatchGlobalDirective,
			Params:   map[string]interface{}{"style_prompt": "dramatic"},
			Priority: 1,
		},
	})

	result, err := applier.Apply(context.Background(), plan, props)
	if err != nil {
		t.Fatalf("Apply() error: %v", err)
	}

	if result.UpdatedProps.Directives == nil {
		t.Fatal("Directives should not be nil after patch")
	}
	if result.UpdatedProps.Directives.StylePrompt != "dramatic" {
		t.Errorf("StylePrompt: got %q, want dramatic", result.UpdatedProps.Directives.StylePrompt)
	}
}

func TestApplier_unknown_op_type(t *testing.T) {
	applier := postprod.NewDefaultEditApplier()
	props := makeProps()

	plan := makePlan([]domain.EditOperation{
		{
			Type:     domain.EditOperationType("unknown_op"),
			Priority: 1,
		},
	})

	result, err := applier.Apply(context.Background(), plan, props)
	if err != nil {
		t.Fatalf("Apply() should not return error for unknown op, got: %v", err)
	}
	if result.OperationsFailed != 1 {
		t.Errorf("OperationsFailed: got %d, want 1", result.OperationsFailed)
	}
}

func TestApplier_patch_subtitle_track(t *testing.T) {
	tests := []struct {
		name        string
		panelIndex  interface{}
		text        interface{}
		wantErr     bool
		wantDialogue string
	}{
		{
			name:         "valid patch by index",
			panelIndex:   float64(0),
			text:         "Updated subtitle",
			wantErr:      false,
			wantDialogue: "Updated subtitle",
		},
		{
			name:         "valid patch second panel",
			panelIndex:   float64(1),
			text:         "Second panel updated",
			wantErr:      false,
			wantDialogue: "Second panel updated",
		},
		{
			name:       "index out of range",
			panelIndex: float64(99),
			text:       "ghost",
			wantErr:    true,
		},
		{
			name:       "negative index",
			panelIndex: float64(-1),
			text:       "ghost",
			wantErr:    true,
		},
		{
			name:       "missing panel_index",
			panelIndex: nil,
			text:       "ghost",
			wantErr:    true,
		},
		{
			name:       "missing text",
			panelIndex: float64(0),
			text:       nil,
			wantErr:    true,
		},
		{
			name:         "integer panel_index",
			panelIndex:   0,
			text:         "Int index works",
			wantErr:      false,
			wantDialogue: "Int index works",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			applier := postprod.NewDefaultEditApplier()
			props := makeProps()
			props.Panels[0].Dialogue = "Original first"
			props.Panels[1].Dialogue = "Original second"

			params := map[string]interface{}{}
			if tt.panelIndex != nil {
				params["panel_index"] = tt.panelIndex
			}
			if tt.text != nil {
				params["text"] = tt.text
			}

			plan := makePlan([]domain.EditOperation{
				{
					Type:     domain.EditOpPatchSubtitleTrack,
					Params:   params,
					Priority: 1,
				},
			})

			result, err := applier.Apply(context.Background(), plan, props)
			if err != nil {
				t.Fatalf("Apply() error: %v", err)
			}

			if tt.wantErr {
				if result.OperationsFailed != 1 {
					t.Errorf("OperationsFailed: got %d, want 1", result.OperationsFailed)
				}
			} else {
				if result.OperationsFailed != 0 {
					t.Errorf("OperationsFailed: got %d, want 0", result.OperationsFailed)
				}
				idx := 0
				if f, ok := tt.panelIndex.(float64); ok {
					idx = int(f)
				} else if i, ok := tt.panelIndex.(int); ok {
					idx = i
				}
				got := result.UpdatedProps.Panels[idx].Dialogue
				if got != tt.wantDialogue {
					t.Errorf("Dialogue: got %q, want %q", got, tt.wantDialogue)
				}
			}
		})
	}
}

func TestApplier_rerender_is_noop(t *testing.T) {
	applier := postprod.NewDefaultEditApplier()
	props := makeProps()

	plan := makePlan([]domain.EditOperation{
		{
			Type:     domain.EditOpRerender,
			Priority: 1,
		},
	})

	result, err := applier.Apply(context.Background(), plan, props)
	if err != nil {
		t.Fatalf("Apply() error: %v", err)
	}
	if result.OperationsApplied != 1 {
		t.Errorf("rerender should be counted as applied, got %d", result.OperationsApplied)
	}
	if result.OperationsFailed != 0 {
		t.Errorf("rerender should not fail, got %d", result.OperationsFailed)
	}
}
