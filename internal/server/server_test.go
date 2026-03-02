package server_test

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/baochen10luo/stagenthand/internal/domain"
	"github.com/baochen10luo/stagenthand/internal/server"
	"github.com/baochen10luo/stagenthand/internal/store"
)

func TestServer_GetJob(t *testing.T) {
	mockJobs := store.NewMockJobRepository()
	mockCP := store.NewMockCheckpointRepository()
	srv := server.New(mockJobs, mockCP)

	t.Run("200 OK", func(t *testing.T) {
		job := &domain.Job{ID: "job-1", Status: domain.JobStatusPending}
		mockJobs.Create(job)

		req := httptest.NewRequest(http.MethodGet, "/jobs/job-1", nil)
		w := httptest.NewRecorder()
		srv.Router().ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("status = %d, want %d", w.Code, http.StatusOK)
		}
	})

	t.Run("404 Not Found", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/jobs/not-exist", nil)
		w := httptest.NewRecorder()
		srv.Router().ServeHTTP(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("status = %d, want %d", w.Code, http.StatusNotFound)
		}
	})

	t.Run("500 Internal Error", func(t *testing.T) {
		mockJobs.Fail = true
		defer func() { mockJobs.Fail = false }()

		req := httptest.NewRequest(http.MethodGet, "/jobs/job-1", nil)
		w := httptest.NewRecorder()
		srv.Router().ServeHTTP(w, req)

		if w.Code != http.StatusInternalServerError {
			t.Errorf("status = %d, want %d", w.Code, http.StatusInternalServerError)
		}
	})
}

func TestServer_GetCheckpoint(t *testing.T) {
	mockCP := store.NewMockCheckpointRepository()
	mockJobs := store.NewMockJobRepository()
	srv := server.New(mockJobs, mockCP)

	t.Run("200 OK", func(t *testing.T) {
		cp := &domain.Checkpoint{ID: "cp-1", Status: domain.CheckpointStatusPending}
		mockCP.Create(cp)

		req := httptest.NewRequest(http.MethodGet, "/checkpoints/cp-1", nil)
		w := httptest.NewRecorder()
		srv.Router().ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("status = %d, want %d", w.Code, http.StatusOK)
		}
	})

	t.Run("500 Internal Error", func(t *testing.T) {
		mockCP.Fail = true
		defer func() { mockCP.Fail = false }()

		req := httptest.NewRequest(http.MethodGet, "/checkpoints/cp-1", nil)
		w := httptest.NewRecorder()
		srv.Router().ServeHTTP(w, req)

		if w.Code != http.StatusInternalServerError {
			t.Errorf("status = %d, want %d", w.Code, http.StatusInternalServerError)
		}
	})
}

func TestServer_ApproveCheckpoint(t *testing.T) {
	mockCP := store.NewMockCheckpointRepository()
	mockJobs := store.NewMockJobRepository()
	srv := server.New(mockJobs, mockCP)

	t.Run("200 OK", func(t *testing.T) {
		cp := &domain.Checkpoint{ID: "cp-1", Status: domain.CheckpointStatusPending}
		mockCP.Create(cp)

		body := bytes.NewReader([]byte(`{"notes": "ok"}`))
		req := httptest.NewRequest(http.MethodPost, "/checkpoints/cp-1/approve", body)
		w := httptest.NewRecorder()
		srv.Router().ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("status = %d, want %d", w.Code, http.StatusOK)
		}
	})

	t.Run("500 Internal Error", func(t *testing.T) {
		cp := &domain.Checkpoint{ID: "cp-err", Status: domain.CheckpointStatusPending}
		mockCP.Create(cp)

		mockCP.Fail = true
		defer func() { mockCP.Fail = false }()

		body := bytes.NewReader([]byte(`{"notes": "ok"}`))
		req := httptest.NewRequest(http.MethodPost, "/checkpoints/cp-err/approve", body)
		w := httptest.NewRecorder()
		srv.Router().ServeHTTP(w, req)

		if w.Code != http.StatusInternalServerError {
			t.Errorf("status = %d, want %d", w.Code, http.StatusInternalServerError)
		}
	})
}

func TestServer_RejectCheckpoint(t *testing.T) {
	mockCP := store.NewMockCheckpointRepository()
	mockJobs := store.NewMockJobRepository()
	srv := server.New(mockJobs, mockCP)

	t.Run("200 OK", func(t *testing.T) {
		cp := &domain.Checkpoint{ID: "cp-1", Status: domain.CheckpointStatusPending}
		mockCP.Create(cp)

		body := bytes.NewReader([]byte(`{"notes": "no"}`))
		req := httptest.NewRequest(http.MethodPost, "/checkpoints/cp-1/reject", body)
		w := httptest.NewRecorder()
		srv.Router().ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("status = %d, want %d", w.Code, http.StatusOK)
		}
	})

	t.Run("500 Internal Error", func(t *testing.T) {
		cp := &domain.Checkpoint{ID: "cp-err", Status: domain.CheckpointStatusPending}
		mockCP.Create(cp)

		mockCP.Fail = true
		defer func() { mockCP.Fail = false }()

		body := bytes.NewReader([]byte(`{"notes": "no"}`))
		req := httptest.NewRequest(http.MethodPost, "/checkpoints/cp-err/reject", body)
		w := httptest.NewRecorder()
		srv.Router().ServeHTTP(w, req)

		if w.Code != http.StatusInternalServerError {
			t.Errorf("status = %d, want %d", w.Code, http.StatusInternalServerError)
		}
	})
}

func TestServer_ApproveReject_BadJSON(t *testing.T) {
	tests := []struct {
		name string
		path string
	}{
		{"approve bad json", "/checkpoints/cp-1/approve"},
		{"reject bad json", "/checkpoints/cp-1/reject"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockCP := store.NewMockCheckpointRepository()
			mockJobs := store.NewMockJobRepository()
			mockCP.Create(&domain.Checkpoint{ID: "cp-1", Status: domain.CheckpointStatusPending})
			srv := server.New(mockJobs, mockCP)

			body := bytes.NewReader([]byte(`{bad json`))
			req := httptest.NewRequest(http.MethodPost, tt.path, body)
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			srv.Router().ServeHTTP(w, req)

			if w.Code != http.StatusBadRequest {
				t.Errorf("%s: status = %d, want %d", tt.name, w.Code, http.StatusBadRequest)
			}
		})
	}
}
