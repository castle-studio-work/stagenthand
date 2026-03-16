// Package postprod implements the Phase 9.5 Agentic post-production loop.
// All interfaces are defined in the consuming package (Dependency Inversion Principle).
package postprod

import (
	"context"

	"github.com/baochen10luo/stagenthand/internal/domain"
)

// VideoEvaluator evaluates a rendered video and returns structured feedback.
type VideoEvaluator interface {
	Evaluate(ctx context.Context, videoPath string, propsJSON []byte) (*EvaluationResult, error)
}

// EvaluationResult mirrors video.Evaluation but lives in this package (DIP).
type EvaluationResult struct {
	VisualScore    int    `json:"visual_score"`
	AudioSyncScore int    `json:"audio_sync_score"`
	AdherenceScore int    `json:"adherence_score"`
	ToneScore      int    `json:"tone_score"`
	Feedback       string `json:"feedback"`
	Action         string `json:"action"` // "APPROVE" or "REJECT"
}

// IsApproved returns true if the evaluation meets the convergence criteria.
func (e *EvaluationResult) IsApproved() bool {
	if e.VisualScore < 8 || e.AudioSyncScore < 8 {
		return false
	}
	return e.VisualScore+e.AudioSyncScore+e.AdherenceScore+e.ToneScore >= 32
}

// EditPlanner converts an evaluation into an EditPlan.
type EditPlanner interface {
	Plan(ctx context.Context, eval *EvaluationResult, currentProps domain.RemotionProps) (*domain.EditPlan, error)
}

// EditApplier executes the operations in an EditPlan against RemotionProps.
type EditApplier interface {
	Apply(ctx context.Context, plan *domain.EditPlan, props domain.RemotionProps) (*domain.EditResult, error)
}

// VideoRenderer renders RemotionProps into an mp4.
type VideoRenderer interface {
	Render(ctx context.Context, propsJSON []byte, outputPath string) error
}

// PropsEvaluation is the result of a pre-render props-only quality check.
type PropsEvaluation struct {
	Issues []string `json:"issues"`
	OK     bool     `json:"ok"`
}

// PropsEvaluator checks RemotionProps JSON for obvious issues before rendering.
type PropsEvaluator interface {
	Evaluate(ctx context.Context, propsJSON []byte) (*PropsEvaluation, error)
}
