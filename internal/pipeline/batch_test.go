package pipeline_test

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"

	"github.com/baochen10luo/stagenthand/internal/domain"
	"github.com/baochen10luo/stagenthand/internal/pipeline"
)

// mockOrchestratorForBatch adapts an Orchestrator for batch testing via a function.
type batchMockOrch struct {
	runFunc func(ctx context.Context, inputData []byte) (*pipeline.PipelineResult, error)
}

func (m *batchMockOrch) Run(ctx context.Context, inputData []byte) (*pipeline.PipelineResult, error) {
	return m.runFunc(ctx, inputData)
}

func TestRunBatch_all_success(t *testing.T) {
	calls := int32(0)
	orch := pipeline.NewOrchestrator(pipeline.OrchestratorDeps{
		LLM: &mockTransformer{
			GenerateFunc: func(_ context.Context, _ string, _ []byte) ([]byte, error) {
				atomic.AddInt32(&calls, 1)
				return []byte(`{"panels":[]}`), nil
			},
		},
		Images:      &mockImageBatcher{},
		Checkpoints: &mockCheckpointStore{approved: true},
		DryRun:      true,
		SkipHITL:    true,
	})

	inputData := []byte(`{"panels":[{"scene_number":1,"panel_number":1,"description":"p","dialogue":"hi","duration_sec":3.0}]}`)
	batchResult, err := pipeline.RunBatch(context.Background(), orch, inputData, pipeline.BatchConfig{
		Episodes:    2,
		Concurrency: 2,
	})
	if err != nil {
		t.Fatalf("RunBatch error: %v", err)
	}
	if batchResult.TotalEpisodes != 2 {
		t.Errorf("TotalEpisodes = %d, want 2", batchResult.TotalEpisodes)
	}
	if batchResult.Succeeded != 2 {
		t.Errorf("Succeeded = %d, want 2", batchResult.Succeeded)
	}
	if batchResult.Failed != 0 {
		t.Errorf("Failed = %d, want 0", batchResult.Failed)
	}
}

func TestRunBatch_partial_failure(t *testing.T) {
	// Use a failingImageBatcher that fails on every other call.
	callCount := int32(0)
	failBatcher := &failingImageBatcher{
		failFunc: func() bool {
			n := atomic.AddInt32(&callCount, 1)
			return n%2 == 0 // fail on even calls
		},
	}

	orch := pipeline.NewOrchestrator(pipeline.OrchestratorDeps{
		LLM:         &mockTransformer{output: []byte(`{"panels":[{"scene_number":1,"panel_number":1,"description":"p","dialogue":"hi","duration_sec":3.0}]}`)},
		Images:      failBatcher,
		Checkpoints: &mockCheckpointStore{approved: true},
		DryRun:      false, // must be false so image batcher is called
		SkipHITL:    true,
	})

	// Use story-style input so LLM is called (triggers image generation)
	inputData := []byte(`{"episodes":[{"number":1,"title":"T","synopsis":"S","hook":"H","cliffhanger":"C"}]}`)
	batchResult, err := pipeline.RunBatch(context.Background(), orch, inputData, pipeline.BatchConfig{
		Episodes:    2,
		Concurrency: 1,
	})
	if err != nil {
		t.Fatalf("RunBatch error: %v", err)
	}
	if batchResult.Failed != 1 {
		t.Errorf("Failed = %d, want 1", batchResult.Failed)
	}
	if batchResult.Succeeded != 1 {
		t.Errorf("Succeeded = %d, want 1", batchResult.Succeeded)
	}
}

// failingImageBatcher is an ImageBatcher that fails when failFunc returns true.
type failingImageBatcher struct {
	failFunc func() bool
}

func (f *failingImageBatcher) BatchGenerateImages(_ context.Context, panels []domain.Panel, _ string) ([]domain.Panel, error) {
	if f.failFunc() {
		return nil, errors.New("image generation failed")
	}
	return panels, nil
}

func TestRunBatch_concurrency_limit(t *testing.T) {
	// Verify that with concurrency=1, episodes run serially (max 1 in-flight at a time).
	active := int32(0)
	maxActive := int32(0)

	orch := pipeline.NewOrchestrator(pipeline.OrchestratorDeps{
		LLM: &mockTransformer{
			GenerateFunc: func(_ context.Context, _ string, _ []byte) ([]byte, error) {
				cur := atomic.AddInt32(&active, 1)
				// Update max
				for {
					m := atomic.LoadInt32(&maxActive)
					if cur <= m {
						break
					}
					if atomic.CompareAndSwapInt32(&maxActive, m, cur) {
						break
					}
				}
				time.Sleep(10 * time.Millisecond)
				atomic.AddInt32(&active, -1)
				return []byte(`{"panels":[]}`), nil
			},
		},
		Images:      &mockImageBatcher{},
		Checkpoints: &mockCheckpointStore{approved: true},
		DryRun:      true,
		SkipHITL:    true,
	})

	inputData := []byte(`{"panels":[{"scene_number":1,"panel_number":1,"description":"p","dialogue":"hi","duration_sec":3.0}]}`)
	_, err := pipeline.RunBatch(context.Background(), orch, inputData, pipeline.BatchConfig{
		Episodes:    4,
		Concurrency: 1,
	})
	if err != nil {
		t.Fatalf("RunBatch error: %v", err)
	}
	if maxActive > 1 {
		t.Errorf("concurrency=1 but max concurrent was %d", maxActive)
	}
}
