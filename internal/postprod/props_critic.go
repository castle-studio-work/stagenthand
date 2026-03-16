package postprod

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/baochen10luo/stagenthand/internal/llm"
)

const propsCriticSystemPrompt = `You are a QA bot that checks a RemotionProps JSON for obvious issues BEFORE the expensive video render step.

Inspect the JSON and report problems in these categories:
1. Metadata prefix leakage: Any DialogueLine.Text or Panel.Dialogue containing prefixes like "VO:", "Narrator:", "V.O.", "Speaker:", or a character-name followed by a colon (e.g. "小明:") that should have been stripped.
2. Duration too short: Any Panel with DurationSec < 2.0 (too brief to read subtitles).
3. Semantic conflict: The Directives.BGMTags and Directives.StylePrompt should be tonally compatible. Flag obvious mismatches (e.g. bgm_tags="happy_pop" with style_prompt="dark_cyberpunk").

Respond ONLY with valid JSON matching this structure (no markdown):
{
  "issues": ["description of issue 1", "description of issue 2"],
  "ok": true
}

Set "ok" to true only if the "issues" array is empty. Otherwise set "ok" to false.`

// PropsCritic performs a cheap text-only LLM check on RemotionProps JSON
// before committing to an expensive video render + video-critic cycle.
type PropsCritic struct {
	client llm.Client
}

// NewPropsCritic creates a PropsCritic backed by the given LLM client.
func NewPropsCritic(client llm.Client) *PropsCritic {
	return &PropsCritic{client: client}
}

// Evaluate sends the props JSON to an LLM for pre-render quality checking.
func (pc *PropsCritic) Evaluate(ctx context.Context, propsJSON []byte) (*PropsEvaluation, error) {
	respBytes, err := pc.client.GenerateTransformation(ctx, propsCriticSystemPrompt, propsJSON)
	if err != nil {
		return nil, fmt.Errorf("props critic LLM call failed: %w", err)
	}

	var eval PropsEvaluation
	if err := json.Unmarshal(respBytes, &eval); err != nil {
		return nil, fmt.Errorf("props critic: failed to parse LLM response: %w (raw: %s)", err, string(respBytes))
	}

	return &eval, nil
}
