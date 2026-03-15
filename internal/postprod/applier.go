package postprod

import (
	"context"
	"fmt"
	"sort"

	"github.com/baochen10luo/stagenthand/internal/domain"
)

// DefaultEditApplier implements EditApplier by applying operations to RemotionProps.
type DefaultEditApplier struct{}

// NewDefaultEditApplier creates a new DefaultEditApplier.
func NewDefaultEditApplier() *DefaultEditApplier {
	return &DefaultEditApplier{}
}

// Apply executes all operations in the plan (sorted by priority) against the props.
func (a *DefaultEditApplier) Apply(_ context.Context, plan *domain.EditPlan, props domain.RemotionProps) (*domain.EditResult, error) {
	// Deep-copy props to avoid mutating the input
	updated := copyProps(props)

	// Sort operations by priority (ascending, 1=highest)
	ops := make([]domain.EditOperation, len(plan.Operations))
	copy(ops, plan.Operations)
	sort.Slice(ops, func(i, j int) bool {
		return ops[i].Priority < ops[j].Priority
	})

	applied := 0
	failed := 0
	var errs []string

	for _, op := range ops {
		if err := applyOperation(&updated, op); err != nil {
			failed++
			errs = append(errs, fmt.Sprintf("op %s failed: %v", op.Type, err))
		} else {
			applied++
		}
	}

	return &domain.EditResult{
		PlanVersion:       plan.Version,
		OperationsApplied: applied,
		OperationsFailed:  failed,
		UpdatedProps:      updated,
		Success:           failed == 0,
		Errors:            errs,
	}, nil
}

// applyOperation applies a single EditOperation to the props in-place.
func applyOperation(props *domain.RemotionProps, op domain.EditOperation) error {
	switch op.Type {
	case domain.EditOpPatchDialogue:
		return applyPatchDialogue(props, op)

	case domain.EditOpPatchDuration:
		return applyPatchDuration(props, op)

	case domain.EditOpPatchPanelDirective:
		return applyPatchPanelDirective(props, op)

	case domain.EditOpPatchGlobalDirective:
		return applyPatchGlobalDirective(props, op)

	case domain.EditOpRegenerateImage:
		return applyRegenerateImage(props, op)

	case domain.EditOpRegenerateAudio:
		return applyRegenerateAudio(props, op)

	case domain.EditOpReplaceBGM:
		return applyReplaceBGM(props, op)

	case domain.EditOpRerender:
		// no-op: handled by cmd layer
		return nil

	default:
		return fmt.Errorf("unknown operation type: %q", op.Type)
	}
}

// findPanel returns a pointer to the panel matching the PanelRef, or an error.
func findPanel(props *domain.RemotionProps, ref *domain.PanelRef) (*domain.Panel, error) {
	if ref == nil {
		return nil, fmt.Errorf("operation requires target_panel but none provided")
	}
	for i := range props.Panels {
		p := &props.Panels[i]
		if p.SceneNumber == ref.SceneNumber && p.PanelNumber == ref.PanelNumber {
			return p, nil
		}
	}
	return nil, fmt.Errorf("panel not found: scene=%d panel=%d", ref.SceneNumber, ref.PanelNumber)
}

func applyPatchDialogue(props *domain.RemotionProps, op domain.EditOperation) error {
	panel, err := findPanel(props, op.TargetPanel)
	if err != nil {
		return err
	}
	dialogue, ok := op.Params["dialogue"].(string)
	if !ok {
		return fmt.Errorf("patch_dialogue: params[\"dialogue\"] must be a string")
	}
	panel.Dialogue = dialogue
	return nil
}

func applyPatchDuration(props *domain.RemotionProps, op domain.EditOperation) error {
	panel, err := findPanel(props, op.TargetPanel)
	if err != nil {
		return err
	}
	var factor float64
	switch v := op.Params["factor"].(type) {
	case float64:
		factor = v
	case int:
		factor = float64(v)
	default:
		return fmt.Errorf("patch_duration: params[\"factor\"] must be a number")
	}
	panel.DurationSec = panel.DurationSec * factor
	return nil
}

func applyPatchPanelDirective(props *domain.RemotionProps, op domain.EditOperation) error {
	panel, err := findPanel(props, op.TargetPanel)
	if err != nil {
		return err
	}
	if panel.Directive == nil {
		panel.Directive = &domain.PanelDirective{}
	}
	for k, v := range op.Params {
		switch k {
		case "motion_effect":
			if s, ok := v.(string); ok {
				panel.Directive.MotionEffect = s
			}
		case "motion_intensity":
			if f, ok := toFloat64(v); ok {
				panel.Directive.MotionIntensity = f
			}
		case "transition_in":
			if s, ok := v.(string); ok {
				panel.Directive.TransitionIn = s
			}
		case "transition_out":
			if s, ok := v.(string); ok {
				panel.Directive.TransitionOut = s
			}
		case "transition_duration_ms":
			if f, ok := toFloat64(v); ok {
				panel.Directive.TransitionDurationMs = int(f)
			}
		case "subtitle_effect":
			if s, ok := v.(string); ok {
				panel.Directive.SubtitleEffect = s
			}
		case "subtitle_font_size":
			if f, ok := toFloat64(v); ok {
				panel.Directive.SubtitleFontSize = int(f)
			}
		case "subtitle_position":
			if s, ok := v.(string); ok {
				panel.Directive.SubtitlePosition = s
			}
		}
	}
	return nil
}

func applyPatchGlobalDirective(props *domain.RemotionProps, op domain.EditOperation) error {
	if props.Directives == nil {
		props.Directives = &domain.Directives{}
	}
	for k, v := range op.Params {
		switch k {
		case "style_prompt":
			if s, ok := v.(string); ok {
				props.Directives.StylePrompt = s
			}
		case "ducking_depth":
			if f, ok := toFloat64(v); ok {
				if f < 0.1 {
					f = 0.1
				}
				props.Directives.DuckingDepth = f
			}
		case "bgm_tags":
			if s, ok := v.(string); ok {
				props.Directives.BGMTags = s
			}
		case "bgm_volume":
			if f, ok := toFloat64(v); ok {
				props.Directives.BGMVolume = f
			}
		case "color_filter":
			if s, ok := v.(string); ok {
				props.Directives.ColorFilter = s
			}
		}
	}
	return nil
}

func applyRegenerateImage(props *domain.RemotionProps, op domain.EditOperation) error {
	panel, err := findPanel(props, op.TargetPanel)
	if err != nil {
		return err
	}
	panel.ImageURL = ""
	return nil
}

func applyRegenerateAudio(props *domain.RemotionProps, op domain.EditOperation) error {
	panel, err := findPanel(props, op.TargetPanel)
	if err != nil {
		return err
	}
	panel.AudioURL = ""
	return nil
}

func applyReplaceBGM(props *domain.RemotionProps, op domain.EditOperation) error {
	props.BGMURL = ""
	if props.Directives == nil {
		props.Directives = &domain.Directives{}
	}
	if tags, ok := op.Params["bgm_tags"].(string); ok {
		props.Directives.BGMTags = tags
	}
	return nil
}

// toFloat64 converts numeric interface values to float64.
func toFloat64(v interface{}) (float64, bool) {
	switch val := v.(type) {
	case float64:
		return val, true
	case float32:
		return float64(val), true
	case int:
		return float64(val), true
	case int64:
		return float64(val), true
	default:
		return 0, false
	}
}

// copyProps performs a shallow copy of RemotionProps with a new Panels slice
// and a new Directives pointer (if non-nil) so mutations don't affect the original.
func copyProps(props domain.RemotionProps) domain.RemotionProps {
	copied := props

	// Copy panels slice
	copied.Panels = make([]domain.Panel, len(props.Panels))
	copy(copied.Panels, props.Panels)

	// Copy panel directives
	for i := range copied.Panels {
		if props.Panels[i].Directive != nil {
			d := *props.Panels[i].Directive
			copied.Panels[i].Directive = &d
		}
		// Copy character refs slice
		if len(props.Panels[i].CharacterRefs) > 0 {
			refs := make([]string, len(props.Panels[i].CharacterRefs))
			copy(refs, props.Panels[i].CharacterRefs)
			copied.Panels[i].CharacterRefs = refs
		}
	}

	// Copy directives
	if props.Directives != nil {
		d := *props.Directives
		copied.Directives = &d
	}

	return copied
}
