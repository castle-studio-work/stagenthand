package store_test

import (
	"testing"
	"time"

	"github.com/baochen10luo/stagenthand/internal/domain"
	"github.com/baochen10luo/stagenthand/internal/store"
)

func TestCheckpointRepository_CreateAndList(t *testing.T) {
	db, _ := store.New(":memory:")
	repo := store.NewGormCheckpointRepository(db)

	cp := &domain.Checkpoint{
		ID:        "cp-001",
		JobID:     "job-001",
		Stage:     domain.StageOutline,
		Status:    domain.CheckpointStatusPending,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	if err := repo.Create(cp); err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	list, err := repo.ListByJobID("job-001")
	if err != nil {
		t.Fatalf("ListByJobID() error = %v", err)
	}
	if len(list) != 1 {
		t.Fatalf("len = %d, want 1", len(list))
	}
	if list[0].Stage != domain.StageOutline {
		t.Errorf("Stage = %q, want outline", list[0].Stage)
	}
}

func TestCheckpointRepository_Approve(t *testing.T) {
	db, _ := store.New(":memory:")
	repo := store.NewGormCheckpointRepository(db)

	cp := &domain.Checkpoint{
		ID:        "cp-002",
		JobID:     "job-002",
		Stage:     domain.StageStoryboard,
		Status:    domain.CheckpointStatusPending,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	_ = repo.Create(cp)

	if err := repo.UpdateStatus("cp-002", domain.CheckpointStatusApproved, "LGTM"); err != nil {
		t.Fatalf("UpdateStatus() error = %v", err)
	}

	got, err := repo.GetByID("cp-002")
	if err != nil {
		t.Fatalf("GetByID() error = %v", err)
	}
	if got.Status != domain.CheckpointStatusApproved {
		t.Errorf("Status = %q, want approved", got.Status)
	}
	if got.Notes != "LGTM" {
		t.Errorf("Notes = %q, want LGTM", got.Notes)
	}
}
