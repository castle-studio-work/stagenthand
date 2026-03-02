package store_test

import (
	"testing"
	"time"

	"github.com/baochen10luo/stagenthand/internal/domain"
	"github.com/baochen10luo/stagenthand/internal/store"
)

func TestJobRepository_CreateAndGet(t *testing.T) {
	db, _ := store.New(":memory:")
	repo := store.NewGormJobRepository(db)

	job := &domain.Job{
		ID:        "job-001",
		ProjectID: "proj-001",
		Type:      "pipeline",
		Status:    domain.JobStatusPending,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := repo.Create(job); err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	got, err := repo.GetByID("job-001")
	if err != nil {
		t.Fatalf("GetByID() error = %v", err)
	}
	if got.ID != job.ID {
		t.Errorf("ID = %q, want %q", got.ID, job.ID)
	}
	if got.Status != domain.JobStatusPending {
		t.Errorf("Status = %q, want pending", got.Status)
	}
}

func TestJobRepository_UpdateStatus(t *testing.T) {
	db, _ := store.New(":memory:")
	repo := store.NewGormJobRepository(db)

	job := &domain.Job{
		ID:        "job-002",
		ProjectID: "proj-001",
		Type:      "panel-to-image",
		Status:    domain.JobStatusPending,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	_ = repo.Create(job)

	if err := repo.UpdateStatus("job-002", domain.JobStatusRunning, ""); err != nil {
		t.Fatalf("UpdateStatus() error = %v", err)
	}

	got, _ := repo.GetByID("job-002")
	if got.Status != domain.JobStatusRunning {
		t.Errorf("Status = %q, want running", got.Status)
	}
}

func TestJobRepository_GetByID_NotFound(t *testing.T) {
	db, _ := store.New(":memory:")
	repo := store.NewGormJobRepository(db)

	_, err := repo.GetByID("nonexistent")
	if err != store.ErrNotFound {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}
