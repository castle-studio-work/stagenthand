package cmd

import (
	"context"

	"github.com/baochen10luo/stagenthand/internal/pipeline"
	"github.com/baochen10luo/stagenthand/internal/video"
)

// videoCriticAdapter bridges video.Critic to the pipeline.VideoCriticEvaluator interface.
// Lives in cmd layer to avoid circular import (pipeline → video → llm → pipeline).
type videoCriticAdapter struct {
	critic *video.Critic
}

// newVideoCriticAdapter wraps a video.Critic as a pipeline.VideoCriticEvaluator.
func newVideoCriticAdapter(c *video.Critic) pipeline.VideoCriticEvaluator {
	return &videoCriticAdapter{critic: c}
}

// Evaluate delegates to video.Critic.Evaluate and maps Evaluation to CriticResult.
func (a *videoCriticAdapter) Evaluate(ctx context.Context, videoPath string, propsJSON []byte) (*pipeline.CriticResult, error) {
	eval, err := a.critic.Evaluate(ctx, videoPath, propsJSON)
	if err != nil {
		return nil, err
	}
	return &pipeline.CriticResult{
		VisualScore:    eval.VisualScore,
		AudioSyncScore: eval.AudioSyncScore,
		AdherenceScore: eval.AdherenceScore,
		ToneScore:      eval.ToneScore,
		Feedback:       eval.Feedback,
		Action:         eval.Action,
	}, nil
}
