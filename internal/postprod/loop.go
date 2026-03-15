package postprod

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/baochen10luo/stagenthand/internal/domain"
)

// LoopConfig holds configuration for the PostProdLoop.
type LoopConfig struct {
	MaxIterations int    // maximum number of retry iterations
	OutputDir     string // directory for versioned props, plans, and videos
}

// PostProdLoop orchestrates the agentic post-production feedback cycle.
type PostProdLoop struct {
	evaluator VideoEvaluator
	planner   EditPlanner
	applier   EditApplier
	renderer  VideoRenderer
	cfg       LoopConfig
}

// NewPostProdLoop creates a new PostProdLoop with the given dependencies.
func NewPostProdLoop(
	evaluator VideoEvaluator,
	planner EditPlanner,
	applier EditApplier,
	renderer VideoRenderer,
	cfg LoopConfig,
) *PostProdLoop {
	return &PostProdLoop{
		evaluator: evaluator,
		planner:   planner,
		applier:   applier,
		renderer:  renderer,
		cfg:       cfg,
	}
}

// Run executes the post-production loop until convergence or max iterations.
func (l *PostProdLoop) Run(ctx context.Context, videoPath string, props domain.RemotionProps) (*domain.PostProdLoopResult, error) {
	version := 1
	currentProps := props
	currentVideo := videoPath

	// Ensure output subdirectories exist
	if err := os.MkdirAll(filepath.Join(l.cfg.OutputDir, "edit_plans"), 0755); err != nil {
		return nil, fmt.Errorf("postprod loop: create edit_plans dir: %w", err)
	}
	if err := os.MkdirAll(filepath.Join(l.cfg.OutputDir, "mp4"), 0755); err != nil {
		return nil, fmt.Errorf("postprod loop: create mp4 dir: %w", err)
	}

	for iteration := 1; iteration <= l.cfg.MaxIterations; iteration++ {
		propsJSON, err := json.Marshal(currentProps)
		if err != nil {
			return nil, fmt.Errorf("postprod loop iteration %d: marshal props: %w", iteration, err)
		}

		// Evaluate current video
		eval, err := l.evaluator.Evaluate(ctx, currentVideo, propsJSON)
		if err != nil {
			return nil, fmt.Errorf("postprod loop iteration %d: evaluate: %w", iteration, err)
		}

		if eval.IsApproved() {
			return &domain.PostProdLoopResult{
				Converged:  true,
				Iterations: iteration,
				FinalVideo: currentVideo,
			}, nil
		}

		// Plan edits
		plan, err := l.planner.Plan(ctx, eval, currentProps)
		if err != nil {
			return nil, fmt.Errorf("postprod loop iteration %d: plan: %w", iteration, err)
		}

		// Apply edits
		result, err := l.applier.Apply(ctx, plan, currentProps)
		if err != nil {
			return nil, fmt.Errorf("postprod loop iteration %d: apply: %w", iteration, err)
		}
		currentProps = result.UpdatedProps

		// Write versioned props file
		nextVersion := version + iteration
		propsPath := filepath.Join(l.cfg.OutputDir, fmt.Sprintf("remotion_props_v%d.json", nextVersion))
		updatedPropsJSON, err := json.MarshalIndent(currentProps, "", "  ")
		if err != nil {
			return nil, fmt.Errorf("postprod loop iteration %d: marshal updated props: %w", iteration, err)
		}
		if err := os.WriteFile(propsPath, updatedPropsJSON, 0644); err != nil {
			return nil, fmt.Errorf("postprod loop iteration %d: write props: %w", iteration, err)
		}

		// Write versioned edit plan file
		planPath := filepath.Join(l.cfg.OutputDir, "edit_plans", fmt.Sprintf("edit_plan_v%d.json", nextVersion))
		planJSON, err := json.MarshalIndent(plan, "", "  ")
		if err != nil {
			return nil, fmt.Errorf("postprod loop iteration %d: marshal plan: %w", iteration, err)
		}
		if err := os.WriteFile(planPath, planJSON, 0644); err != nil {
			return nil, fmt.Errorf("postprod loop iteration %d: write plan: %w", iteration, err)
		}

		// Render new video
		newVideoPath := filepath.Join(l.cfg.OutputDir, "mp4", fmt.Sprintf("output_v%d.mp4", nextVersion))
		if err := l.renderer.Render(ctx, updatedPropsJSON, newVideoPath); err != nil {
			return nil, fmt.Errorf("postprod loop iteration %d: render: %w", iteration, err)
		}
		currentVideo = newVideoPath
	}

	return &domain.PostProdLoopResult{
		Converged:  false,
		Iterations: l.cfg.MaxIterations,
		FinalVideo: currentVideo,
	}, nil
}
