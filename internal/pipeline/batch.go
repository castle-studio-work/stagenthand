package pipeline

import (
	"context"
	"sync"
)

// BatchConfig controls batch episode production.
type BatchConfig struct {
	Episodes    int // number of episodes to produce
	Concurrency int // max concurrent workers (default 2)
}

// EpisodeResult holds the result of producing a single episode.
type EpisodeResult struct {
	Episode int            `json:"episode"`
	Result  *PipelineResult `json:"result,omitempty"`
	Error   string         `json:"error,omitempty"`
}

// BatchResult holds all episode results.
type BatchResult struct {
	TotalEpisodes int             `json:"total_episodes"`
	Succeeded     int             `json:"succeeded"`
	Failed        int             `json:"failed"`
	Episodes      []EpisodeResult `json:"episodes"`
}

// RunBatch executes the pipeline for multiple episodes concurrently.
// Each episode uses the same input data. Uses bounded concurrency via semaphore.
func RunBatch(ctx context.Context, orch *Orchestrator, inputData []byte, cfg BatchConfig) (*BatchResult, error) {
	if cfg.Concurrency <= 0 {
		cfg.Concurrency = 2
	}

	results := make([]EpisodeResult, cfg.Episodes)
	sem := make(chan struct{}, cfg.Concurrency)
	var mu sync.Mutex
	var wg sync.WaitGroup

	for i := 0; i < cfg.Episodes; i++ {
		wg.Add(1)
		go func(ep int) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			result, err := orch.Run(ctx, inputData)
			mu.Lock()
			defer mu.Unlock()
			if err != nil {
				results[ep] = EpisodeResult{Episode: ep + 1, Error: err.Error()}
			} else {
				results[ep] = EpisodeResult{Episode: ep + 1, Result: result}
			}
		}(i)
	}
	wg.Wait()

	succeeded := 0
	failed := 0
	for _, r := range results {
		if r.Error != "" {
			failed++
		} else {
			succeeded++
		}
	}

	return &BatchResult{
		TotalEpisodes: cfg.Episodes,
		Succeeded:     succeeded,
		Failed:        failed,
		Episodes:      results,
	}, nil
}
