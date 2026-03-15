package postprod

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/baochen10luo/stagenthand/internal/domain"
)

// llmClient is the minimal interface needed from the LLM for planning.
// We define it here (not importing llm package) to follow DIP.
type llmClient interface {
	GenerateTransformation(ctx context.Context, systemPrompt string, inputData []byte) ([]byte, error)
}

// LLMEditPlanner implements EditPlanner using an LLM with rule-based fallback.
type LLMEditPlanner struct {
	client llmClient
}

// NewLLMEditPlanner creates a new LLMEditPlanner with the given LLM client.
func NewLLMEditPlanner(client llmClient) *LLMEditPlanner {
	return &LLMEditPlanner{client: client}
}

const plannerSystemPrompt = `You are a video post-production expert. Given an AI Critic evaluation and RemotionProps JSON,
output a JSON EditPlan with specific operations to fix the problems.

The EditPlan must match this exact schema:
{
  "version": "string",
  "generated_at": "RFC3339 timestamp",
  "operations": [
    {
      "type": "regenerate_image|regenerate_audio|replace_bgm|patch_dialogue|patch_duration|patch_panel_directive|patch_global_directive|rerender",
      "target_panel": {"scene_number": N, "panel_number": N},  // omit for global ops
      "params": {},  // operation-specific parameters
      "priority": 1,  // 1=highest priority
      "rationale": "why this fix is needed"
    }
  ],
  "estimated_cost_usd": 0.0,
  "rationale": "overall plan rationale"
}

For patch_global_directive:
  params may contain: "style_prompt" (string), "ducking_depth" (float), "bgm_tags" (string)

For patch_dialogue:
  params must contain: "dialogue" (string)

For patch_duration:
  params must contain: "factor" (float, e.g. 1.2 to extend by 20%)

Respond ONLY with valid JSON, no markdown.`

// Plan generates an EditPlan from the evaluation. Falls back to rule-based on LLM error.
func (p *LLMEditPlanner) Plan(ctx context.Context, eval *EvaluationResult, currentProps domain.RemotionProps) (*domain.EditPlan, error) {
	// Build input payload
	type planInput struct {
		Evaluation    *EvaluationResult    `json:"evaluation"`
		CurrentProps  domain.RemotionProps `json:"current_props"`
	}
	inputData, err := json.Marshal(planInput{
		Evaluation:   eval,
		CurrentProps: currentProps,
	})
	if err != nil {
		return p.ruleBasedPlan(eval, currentProps), nil
	}

	respBytes, err := p.client.GenerateTransformation(ctx, plannerSystemPrompt, inputData)
	if err != nil {
		// Fallback to rule-based
		return p.ruleBasedPlan(eval, currentProps), nil
	}

	var plan domain.EditPlan
	if err := json.Unmarshal(respBytes, &plan); err != nil {
		// Fallback to rule-based
		return p.ruleBasedPlan(eval, currentProps), nil
	}

	return &plan, nil
}

// ruleBasedPlan generates operations based on score thresholds.
func (p *LLMEditPlanner) ruleBasedPlan(eval *EvaluationResult, currentProps domain.RemotionProps) *domain.EditPlan {
	ops := make([]domain.EditOperation, 0, 4)

	// VisualScore < 8 → prepend "highly detailed, 8K" to style_prompt
	if eval.VisualScore < 8 {
		stylePrompt := "highly detailed, 8K"
		if currentProps.Directives != nil && currentProps.Directives.StylePrompt != "" {
			stylePrompt = fmt.Sprintf("highly detailed, 8K, %s", currentProps.Directives.StylePrompt)
		}
		ops = append(ops, domain.EditOperation{
			Type:      domain.EditOpPatchGlobalDirective,
			Priority:  1,
			Rationale: fmt.Sprintf("VisualScore=%d < 8: enhance style prompt for better quality", eval.VisualScore),
			Params:    map[string]interface{}{"style_prompt": stylePrompt},
		})
	}

	// AudioSyncScore < 8 → reduce ducking_depth by 0.1
	if eval.AudioSyncScore < 8 {
		currentDucking := 0.5
		if currentProps.Directives != nil && currentProps.Directives.DuckingDepth > 0 {
			currentDucking = currentProps.Directives.DuckingDepth
		}
		newDucking := currentDucking - 0.1
		if newDucking < 0.1 {
			newDucking = 0.1
		}
		ops = append(ops, domain.EditOperation{
			Type:      domain.EditOpPatchGlobalDirective,
			Priority:  1,
			Rationale: fmt.Sprintf("AudioSyncScore=%d < 8: reduce ducking depth for better voice clarity", eval.AudioSyncScore),
			Params:    map[string]interface{}{"ducking_depth": newDucking},
		})
	}

	// ToneScore < 6 → extend all panel durations by 1.2x
	if eval.ToneScore < 6 {
		for _, panel := range currentProps.Panels {
			ops = append(ops, domain.EditOperation{
				Type: domain.EditOpPatchDuration,
				TargetPanel: &domain.PanelRef{
					SceneNumber: panel.SceneNumber,
					PanelNumber: panel.PanelNumber,
				},
				Priority:  2,
				Rationale: fmt.Sprintf("ToneScore=%d < 6: extend duration for better pacing", eval.ToneScore),
				Params:    map[string]interface{}{"factor": 1.2},
			})
		}
	}

	// AdherenceScore < 8 → regenerate image for first panel (proxy for lowest scoring panel)
	if eval.AdherenceScore < 8 && len(currentProps.Panels) > 0 {
		first := currentProps.Panels[0]
		ops = append(ops, domain.EditOperation{
			Type: domain.EditOpRegenerateImage,
			TargetPanel: &domain.PanelRef{
				SceneNumber: first.SceneNumber,
				PanelNumber: first.PanelNumber,
			},
			Priority:  3,
			Rationale: fmt.Sprintf("AdherenceScore=%d < 8: regenerate image to better match directives", eval.AdherenceScore),
		})
	}

	return &domain.EditPlan{
		Version:       "fallback-v1",
		GeneratedAt:   time.Now(),
		Operations:    ops,
		EstimatedCost: 0.0,
		Rationale:     "Rule-based fallback plan (LLM unavailable)",
	}
}
