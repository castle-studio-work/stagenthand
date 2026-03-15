package pipeline

import (
	"context"
	"fmt"
	"sync"

	"github.com/baochen10luo/stagenthand/internal/domain"
	"github.com/baochen10luo/stagenthand/internal/series"
)

// BatchConfig controls batch episode production.
type BatchConfig struct {
	Episodes    int // number of episodes to produce
	Concurrency int // max concurrent workers (default 2)

	// Series memory (optional — nil disables series continuity)
	SeriesRepo     series.Repository
	Summarizer     series.Summarizer
	WindowSize     int            // sliding window size, default 3
	CheckpointGate CheckpointGate // for StageSeriesSummary HITL
}

// EpisodeResult holds the result of producing a single episode.
type EpisodeResult struct {
	Episode int             `json:"episode"`
	Result  *PipelineResult `json:"result,omitempty"`
	Error   string          `json:"error,omitempty"`
}

// BatchResult holds all episode results.
type BatchResult struct {
	TotalEpisodes int             `json:"total_episodes"`
	Succeeded     int             `json:"succeeded"`
	Failed        int             `json:"failed"`
	Episodes      []EpisodeResult `json:"episodes"`
}

// RunBatch executes the pipeline for multiple episodes.
//
// When cfg.SeriesRepo is nil: episodes run concurrently up to cfg.Concurrency.
// When cfg.SeriesRepo is set: episodes run serially so each episode can inject
// context from the previous one.
//
// TODO: decouple narrative and production phases for parallel production
func RunBatch(ctx context.Context, orch *Orchestrator, inputData []byte, cfg BatchConfig) (*BatchResult, error) {
	if cfg.Concurrency <= 0 {
		cfg.Concurrency = 2
	}
	if cfg.WindowSize <= 0 {
		cfg.WindowSize = 3
	}

	if cfg.SeriesRepo != nil {
		return runBatchSerial(ctx, orch, inputData, cfg)
	}
	return runBatchConcurrent(ctx, orch, inputData, cfg)
}

// runBatchConcurrent runs episodes concurrently with bounded semaphore.
// Used when series continuity is disabled.
func runBatchConcurrent(ctx context.Context, orch *Orchestrator, inputData []byte, cfg BatchConfig) (*BatchResult, error) {
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

	return tallied(cfg.Episodes, results), nil
}

// runBatchSerial runs episodes one at a time, injecting series context from previous episodes.
// Used when series continuity is enabled (cfg.SeriesRepo != nil).
func runBatchSerial(ctx context.Context, orch *Orchestrator, inputData []byte, cfg BatchConfig) (*BatchResult, error) {
	results := make([]EpisodeResult, cfg.Episodes)

	for i := 0; i < cfg.Episodes; i++ {
		ep := i + 1

		// 1. Build context-injected input for this episode
		episodeInput, err := injectSeriesContext(ctx, inputData, cfg)
		if err != nil {
			results[i] = EpisodeResult{Episode: ep, Error: fmt.Sprintf("context injection failed: %v", err)}
			continue
		}

		// 2. Run full pipeline for this episode
		result, err := orch.Run(ctx, episodeInput)
		if err != nil {
			results[i] = EpisodeResult{Episode: ep, Error: err.Error()}
			continue
		}
		results[i] = EpisodeResult{Episode: ep, Result: result}

		// 3. Extract episode memory via LLM summarizer
		if cfg.Summarizer != nil {
			storyboardJSON, _ := marshalStoryboard(result)
			epMem, sumErr := cfg.Summarizer.Summarize(ctx, ep, storyboardJSON)
			if sumErr != nil {
				// Non-fatal: log but continue
				results[i].Error = fmt.Sprintf("summarize failed: %v", sumErr)
			} else {
				// 4. Optional HITL checkpoint for series summary review
				if cfg.CheckpointGate != nil {
					jobID := fmt.Sprintf("batch-ep%d", ep)
					if ckptErr := cfg.CheckpointGate.CreateAndWait(ctx, jobID, domain.StageSeriesSummary); ckptErr != nil {
						results[i].Error = fmt.Sprintf("series summary checkpoint failed: %v", ckptErr)
						continue
					}
				}

				// 5. Append episode memory to series repository
				if appendErr := cfg.SeriesRepo.Append(ctx, epMem); appendErr != nil {
					// Non-fatal
					_ = appendErr
				}

				// 6. Compress global summary and save
				if updated, loadErr := cfg.SeriesRepo.Load(ctx); loadErr == nil {
					if globalSummary, cErr := cfg.Summarizer.CompressGlobal(ctx, updated); cErr == nil {
						updated.GlobalSummary = globalSummary
						_ = cfg.SeriesRepo.Save(ctx, updated)
					}
				}
			}
		}
	}

	return tallied(cfg.Episodes, results), nil
}

// injectSeriesContext prepends the series context window to inputData.
func injectSeriesContext(ctx context.Context, inputData []byte, cfg BatchConfig) ([]byte, error) {
	mem, err := cfg.SeriesRepo.Load(ctx)
	if err != nil {
		return nil, err
	}

	window := series.BuildContextWindow(mem, cfg.WindowSize)
	contextBlock := series.FormatContextPrompt(window)

	// Prepend context as plain text before the input data
	injected := []byte(contextBlock + "\n\n" + string(inputData))
	return injected, nil
}

// marshalStoryboard serialises the storyboard from a PipelineResult.
func marshalStoryboard(result *PipelineResult) ([]byte, error) {
	if result == nil {
		return []byte("{}"), nil
	}
	data, err := jsonMarshal(result.Storyboard)
	if err != nil {
		return []byte("{}"), err
	}
	return data, nil
}

// tallied builds a BatchResult from the episode results slice.
func tallied(totalEpisodes int, results []EpisodeResult) *BatchResult {
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
		TotalEpisodes: totalEpisodes,
		Succeeded:     succeeded,
		Failed:        failed,
		Episodes:      results,
	}
}
