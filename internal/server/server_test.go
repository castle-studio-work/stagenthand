package server_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/baochen10luo/stagenthand/internal/domain"
	"github.com/baochen10luo/stagenthand/internal/server"
	"github.com/baochen10luo/stagenthand/internal/store"
)

func setup() (*server.Server, store.JobRepository, store.CheckpointRepository) {
	jobRepo := store.NewMockJobRepository()
	cpRepo := store.NewMockCheckpointRepository()
	s := server.New(jobRepo, cpRepo)
	return s, jobRepo, cpRepo
}

func TestGetJob_OK(t *testing.T) {
	s, jobRepo, _ := setup()

	_ = jobRepo.Create(&domain.Job{
		ID:        "job-abc",
		ProjectID: "proj-1",
		Status:    domain.JobStatusRunning,
		CreatedAt: time.Now(),
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/jobs/job-abc", nil)
	s.Router().ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", w.Code)
	}
	var got domain.Job
	if err := json.NewDecoder(w.Body).Decode(&got); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if got.ID != "job-abc" {
		t.Errorf("ID = %q, want job-abc", got.ID)
	}
}

func TestGetJob_NotFound(t *testing.T) {
	s, _, _ := setup()
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/jobs/nope", nil)
	s.Router().ServeHTTP(w, req)
	if w.Code != http.StatusNotFound {
		t.Errorf("status = %d, want 404", w.Code)
	}
}

func TestApproveCheckpoint_OK(t *testing.T) {
	s, _, cpRepo := setup()

	_ = cpRepo.Create(&domain.Checkpoint{
		ID:        "cp-xyz",
		JobID:     "job-abc",
		Stage:     domain.StageOutline,
		Status:    domain.CheckpointStatusPending,
		CreatedAt: time.Now(),
	})

	body, _ := json.Marshal(map[string]string{"notes": "looks good"})
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/checkpoints/cp-xyz/approve", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	s.Router().ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", w.Code)
	}

	got, _ := cpRepo.GetByID("cp-xyz")
	if got.Status != domain.CheckpointStatusApproved {
		t.Errorf("Status = %q, want approved", got.Status)
	}
}

func TestRejectCheckpoint_OK(t *testing.T) {
	s, _, cpRepo := setup()

	_ = cpRepo.Create(&domain.Checkpoint{
		ID:        "cp-rej",
		JobID:     "job-abc",
		Stage:     domain.StageStoryboard,
		Status:    domain.CheckpointStatusPending,
		CreatedAt: time.Now(),
	})

	body, _ := json.Marshal(map[string]string{"notes": "needs rework"})
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/checkpoints/cp-rej/reject", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	s.Router().ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", w.Code)
	}
}
