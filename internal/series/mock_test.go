package series_test

import (
	"context"
	"errors"
	"testing"

	"github.com/baochen10luo/stagenthand/internal/series"
)

func TestMockRepository_LoadEmpty(t *testing.T) {
	repo := &series.MockRepository{}
	m, err := repo.Load(context.Background())
	if err != nil {
		t.Fatalf("Load error: %v", err)
	}
	if m == nil {
		t.Fatal("Load returned nil")
	}
	if len(m.Episodes) != 0 {
		t.Errorf("expected empty episodes, got %d", len(m.Episodes))
	}
}

func TestMockRepository_SaveLoad(t *testing.T) {
	repo := &series.MockRepository{}
	ctx := context.Background()

	original := &series.SeriesMemory{SeriesTitle: "My Series"}
	if err := repo.Save(ctx, original); err != nil {
		t.Fatalf("Save error: %v", err)
	}

	loaded, err := repo.Load(ctx)
	if err != nil {
		t.Fatalf("Load error: %v", err)
	}
	if loaded.SeriesTitle != "My Series" {
		t.Errorf("SeriesTitle: got %q, want %q", loaded.SeriesTitle, "My Series")
	}
}

func TestMockRepository_Append(t *testing.T) {
	repo := &series.MockRepository{}
	ctx := context.Background()

	ep := series.EpisodeMemory{Episode: 1, KeyEvents: []string{"event"}}
	if err := repo.Append(ctx, ep); err != nil {
		t.Fatalf("Append error: %v", err)
	}

	loaded, err := repo.Load(ctx)
	if err != nil {
		t.Fatalf("Load error: %v", err)
	}
	if len(loaded.Episodes) != 1 {
		t.Fatalf("Episodes length: got %d, want 1", len(loaded.Episodes))
	}
	if loaded.Episodes[0].Episode != 1 {
		t.Errorf("Episode: got %d, want 1", loaded.Episodes[0].Episode)
	}
}

func TestMockRepository_Error(t *testing.T) {
	expectedErr := errors.New("mock error")
	repo := &series.MockRepository{Err: expectedErr}
	ctx := context.Background()

	_, err := repo.Load(ctx)
	if err == nil {
		t.Error("expected Load to return error")
	}

	err = repo.Save(ctx, &series.SeriesMemory{})
	if err == nil {
		t.Error("expected Save to return error")
	}

	err = repo.Append(ctx, series.EpisodeMemory{})
	if err == nil {
		t.Error("expected Append to return error")
	}
}

func TestMockSummarizer_Summarize(t *testing.T) {
	mockSum := &series.MockSummarizer{
		Result: series.EpisodeMemory{KeyEvents: []string{"event A"}},
		Global: "Global summary text",
	}
	ctx := context.Background()

	result, err := mockSum.Summarize(ctx, 3, []byte(`{}`))
	if err != nil {
		t.Fatalf("Summarize error: %v", err)
	}
	if result.Episode != 3 {
		t.Errorf("Episode: got %d, want 3", result.Episode)
	}
	if len(result.KeyEvents) != 1 || result.KeyEvents[0] != "event A" {
		t.Errorf("KeyEvents: got %v", result.KeyEvents)
	}
	if mockSum.SummarizeCalls != 1 {
		t.Errorf("SummarizeCalls: got %d, want 1", mockSum.SummarizeCalls)
	}
}

func TestMockSummarizer_CompressGlobal(t *testing.T) {
	mockSum := &series.MockSummarizer{
		Global: "The story so far",
	}

	result, err := mockSum.CompressGlobal(context.Background(), &series.SeriesMemory{})
	if err != nil {
		t.Fatalf("CompressGlobal error: %v", err)
	}
	if result != "The story so far" {
		t.Errorf("CompressGlobal: got %q, want %q", result, "The story so far")
	}
	if mockSum.CompressGlobalCalls != 1 {
		t.Errorf("CompressGlobalCalls: got %d, want 1", mockSum.CompressGlobalCalls)
	}
}

func TestMockSummarizer_Error(t *testing.T) {
	expectedErr := errors.New("summarizer error")
	mockSum := &series.MockSummarizer{Err: expectedErr}

	_, err := mockSum.Summarize(context.Background(), 1, []byte(`{}`))
	if err == nil {
		t.Error("expected Summarize to return error")
	}

	_, err = mockSum.CompressGlobal(context.Background(), &series.SeriesMemory{})
	if err == nil {
		t.Error("expected CompressGlobal to return error")
	}
}
