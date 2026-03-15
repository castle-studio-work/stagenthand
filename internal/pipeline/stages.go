package pipeline

import (
	"context"
	"errors"
	"fmt"

	"github.com/baochen10luo/stagenthand/internal/domain"
)

// Transformer defines the behavior needed to run a transformation stage.
// This is exactly the llm.Client footprint, kept clean.
type Transformer interface {
	GenerateTransformation(ctx context.Context, systemPrompt string, inputData []byte) ([]byte, error)
}

// RunTransformationStage executes a single LLM transformation pipeline step.
func RunTransformationStage(ctx context.Context, transformer Transformer, systemPrompt string, inputData []byte) ([]byte, error) {
	if len(inputData) == 0 {
		return nil, errors.New("input data cannot be empty")
	}

	if systemPrompt == "" {
		return nil, errors.New("system prompt cannot be empty")
	}

	output, err := transformer.GenerateTransformation(ctx, systemPrompt, inputData)
	if err != nil {
		return nil, fmt.Errorf("transformer failed: %w", err)
	}

	if len(output) == 0 {
		return nil, errors.New("transformer returned empty output")
	}

	return output, nil
}

// languageInstructions maps BCP-47 language tags to dialogue instructions appended to PromptStoryboardToPanels.
var languageInstructions = map[string]string{
	"en-US": "IMPORTANT: All 'dialogue' fields MUST be written in English. Use natural American English.",
	"en-GB": "IMPORTANT: All 'dialogue' fields MUST be written in English. Use natural British English.",
	"ja-JP": "IMPORTANT: All 'dialogue' fields MUST be written in Japanese (日本語). Use natural conversational Japanese.",
	"ko-KR": "IMPORTANT: All 'dialogue' fields MUST be written in Korean (한국어). Use natural conversational Korean.",
	"cmn-CN": "IMPORTANT: All 'dialogue' fields MUST be written in Simplified Chinese (简体中文).",
}

// buildStoryboardToPanelsPrompt returns the PromptStoryboardToPanels with optional language instruction appended.
func buildStoryboardToPanelsPrompt(language string, sb domain.Storyboard) string {
	base := PromptStoryboardToPanels

	// Check language from storyboard directives first, then from orchestrator deps language
	lang := language
	if sb.Directives != nil && sb.Directives.Language != "" {
		lang = sb.Directives.Language
	}

	if lang == "" || lang == "zh-TW" {
		return base
	}

	if instruction, ok := languageInstructions[lang]; ok {
		return base + "\n" + instruction
	}
	return base
}

// System prompts for the Phase 2 stages.
const (
	PromptStoryToOutline = `You are an expert story outliner. Read the input story prompt and generate a JSON outline.
Output JSON MUST follow this outline schema:
{
  "project_id": "...",
  "episodes": [
    {
      "number": 1,
      "title": "...",
      "synopsis": "...",
      "hook": "...",
      "cliffhanger": "..."
    }
  ]
}`

	PromptOutlineToStoryboard = `You are a storyboard director. Convert the input outline JSON into a localized scene-by-scene storyboard JSON. Ensure your scenes follow a cohesive 3-act narrative arc (setup, conflict, resolution).
CRITICAL: If the story lacks spoken dialogue, you MUST invent a compelling voiceover (VO) narrator or internal monologue to convey the deeper emotion, sacrifice, or meaning of the scene. Do not leave the story silent, otherwise the audience will not understand the plot.
Output JSON MUST follow this schema:
{
  "project_id": "...",
  "episode": 1,
  "directives": {
    "style_prompt": "YOUR_ACTUAL_STYLE_PROMPT_HERE (e.g., 'photorealistic cyberpunk, dark noir')",
    "color_filter": "cinematic",
    "bgm_tags": "suspense+dark+ambient"
  },
  "scenes": [
    {
      "number": 1,
      "description": "..."
    }
  ]
}`

	PromptStoryboardToPanels = `You are a visual panel designer. Convert the input storyboard JSON into a detailed panel-by-panel generation JSON.
Target total video length: approximately 30–50 seconds. Use 4–7 panels maximum.
Each panel's 'duration_sec' should reflect the time needed to naturally speak the dialogue aloud PLUS viewer breathing time. Estimate ~0.12 seconds per character. Keep individual dialogue short and punchy — no more than 30 words per panel.
CRITICAL: Every panel MUST have a 'dialogue' field. If the character is not speaking, use a VoiceOver (VO) to narrate the emotion, sacrifice, or plot context so the audience understands what is happening.
Output JSON MUST follow this schema:
{
  "project_id": "...",
  "episode": 1,
  "panels": [
    {
      "scene_number": 1,
      "panel_number": 1,
      "description": "...",
      "dialogue": "...",
      "character_refs": [],
      "duration_sec": 4.0
    }
  ]
}`
)
