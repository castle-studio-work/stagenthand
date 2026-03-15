package series_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/baochen10luo/stagenthand/internal/series"
)

func TestFileRepository_Load_NotFound(t *testing.T) {
	dir := t.TempDir()
	repo := series.NewFileRepository(filepath.Join(dir, "nonexistent.json"))

	m, err := repo.Load(context.Background())
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}
	if m == nil {
		t.Fatal("Load returned nil, want empty SeriesMemory")
	}
	if m.SeriesTitle != "" || len(m.Episodes) != 0 {
		t.Errorf("expected empty SeriesMemory, got %+v", m)
	}
}

func TestFileRepository_SaveLoad(t *testing.T) {
	dir := t.TempDir()
	repo := series.NewFileRepository(filepath.Join(dir, "memory.json"))
	ctx := context.Background()

	original := &series.SeriesMemory{
		SeriesTitle:   "Epic Journey",
		GlobalSummary: "A hero's quest begins",
		Episodes: []series.EpisodeMemory{
			{
				Episode:    1,
				KeyEvents:  []string{"hero starts"},
				Characters: []series.CharacterSnapshot{{Name: "Hero", State: "ready"}},
				WorldFacts: []string{"world is at war"},
			},
		},
	}

	if err := repo.Save(ctx, original); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	loaded, err := repo.Load(ctx)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if loaded.SeriesTitle != original.SeriesTitle {
		t.Errorf("SeriesTitle: got %q, want %q", loaded.SeriesTitle, original.SeriesTitle)
	}
	if loaded.GlobalSummary != original.GlobalSummary {
		t.Errorf("GlobalSummary: got %q, want %q", loaded.GlobalSummary, original.GlobalSummary)
	}
	if len(loaded.Episodes) != 1 {
		t.Fatalf("Episodes length: got %d, want 1", len(loaded.Episodes))
	}
	if loaded.Episodes[0].Episode != 1 {
		t.Errorf("Episode number: got %d, want 1", loaded.Episodes[0].Episode)
	}
}

func TestFileRepository_Append(t *testing.T) {
	dir := t.TempDir()
	repo := series.NewFileRepository(filepath.Join(dir, "memory.json"))
	ctx := context.Background()

	ep1 := series.EpisodeMemory{
		Episode:    1,
		KeyEvents:  []string{"event one"},
		Characters: []series.CharacterSnapshot{},
		WorldFacts: []string{},
	}
	ep2 := series.EpisodeMemory{
		Episode:    2,
		KeyEvents:  []string{"event two"},
		Characters: []series.CharacterSnapshot{},
		WorldFacts: []string{},
	}

	if err := repo.Append(ctx, ep1); err != nil {
		t.Fatalf("Append ep1 failed: %v", err)
	}
	if err := repo.Append(ctx, ep2); err != nil {
		t.Fatalf("Append ep2 failed: %v", err)
	}

	loaded, err := repo.Load(ctx)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}
	if len(loaded.Episodes) != 2 {
		t.Fatalf("Episodes length: got %d, want 2", len(loaded.Episodes))
	}
	if loaded.Episodes[0].Episode != 1 {
		t.Errorf("Episodes[0].Episode: got %d, want 1", loaded.Episodes[0].Episode)
	}
	if loaded.Episodes[1].Episode != 2 {
		t.Errorf("Episodes[1].Episode: got %d, want 2", loaded.Episodes[1].Episode)
	}
}

func TestFileRepository_Save_AtomicWrite(t *testing.T) {
	// Verify the file is created (and directory is created if needed)
	dir := t.TempDir()
	subDir := filepath.Join(dir, "nested", "dir")
	repo := series.NewFileRepository(filepath.Join(subDir, "memory.json"))
	ctx := context.Background()

	m := &series.SeriesMemory{SeriesTitle: "Test"}
	if err := repo.Save(ctx, m); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	if _, err := os.Stat(filepath.Join(subDir, "memory.json")); err != nil {
		t.Errorf("file not created: %v", err)
	}
}

func TestFileRepository_Load_InvalidJSON(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "bad.json")

	// Write invalid JSON
	if err := os.WriteFile(path, []byte("not valid json!!!"), 0644); err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}

	repo := series.NewFileRepository(path)
	_, err := repo.Load(context.Background())
	if err == nil {
		t.Error("expected error for invalid JSON, got nil")
	}
}

func TestFileRepository_UpdatedAt_SetOnSave(t *testing.T) {
	dir := t.TempDir()
	repo := series.NewFileRepository(filepath.Join(dir, "memory.json"))
	ctx := context.Background()

	m := &series.SeriesMemory{SeriesTitle: "Time Test"}
	if err := repo.Save(ctx, m); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	loaded, err := repo.Load(ctx)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}
	if loaded.UpdatedAt.IsZero() {
		t.Error("UpdatedAt should be set after Save, got zero time")
	}
}
