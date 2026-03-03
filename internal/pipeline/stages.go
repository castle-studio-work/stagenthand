package pipeline

import (
	"context"
	"errors"
	"fmt"
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

	PromptOutlineToStoryboard = `You are a storyboard director. Convert the input outline JSON into a localized scene-by-scene storyboard JSON.
Output JSON MUST follow this schema:
{
  "project_id": "...",
  "episode": 1,
  "scenes": [
    {
      "number": 1,
      "description": "..."
    }
  ]
}`

	PromptStoryboardToPanels = `You are a visual panel designer. Convert the input storyboard JSON into a detailed panel-by-panel generation JSON.
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
      "duration_sec": 3.0
    }
  ]
}`
)
