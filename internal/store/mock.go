// Package store provides in-memory mocks for testing.
package store

import (
	"sync"

	"github.com/baochen10luo/stagenthand/internal/domain"
)

// MockJobRepository is an in-memory JobRepository for tests.
type MockJobRepository struct {
	mu   sync.RWMutex
	jobs map[string]*domain.Job
	Fail bool
}

func NewMockJobRepository() *MockJobRepository {
	return &MockJobRepository{jobs: make(map[string]*domain.Job)}
}

func (m *MockJobRepository) Create(job *domain.Job) error {
	if m.Fail {
		return ErrInternal
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	m.jobs[job.ID] = job
	return nil
}

func (m *MockJobRepository) GetByID(id string) (*domain.Job, error) {
	if m.Fail {
		return nil, ErrInternal
	}
	m.mu.RLock()
	defer m.mu.RUnlock()
	j, ok := m.jobs[id]
	if !ok {
		return nil, ErrNotFound
	}
	return j, nil
}

func (m *MockJobRepository) UpdateStatus(id string, status domain.JobStatus, errMsg string) error {
	if m.Fail {
		return ErrInternal
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	j, ok := m.jobs[id]
	if !ok {
		return ErrNotFound
	}
	j.Status = status
	j.Error = errMsg
	return nil
}

// MockCheckpointRepository is an in-memory CheckpointRepository for tests.
type MockCheckpointRepository struct {
	mu   sync.RWMutex
	cps  map[string]*domain.Checkpoint
	Fail bool
}

func NewMockCheckpointRepository() *MockCheckpointRepository {
	return &MockCheckpointRepository{cps: make(map[string]*domain.Checkpoint)}
}

func (m *MockCheckpointRepository) Create(cp *domain.Checkpoint) error {
	if m.Fail {
		return ErrInternal
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	m.cps[cp.ID] = cp
	return nil
}

func (m *MockCheckpointRepository) GetByID(id string) (*domain.Checkpoint, error) {
	if m.Fail {
		return nil, ErrInternal
	}
	m.mu.RLock()
	defer m.mu.RUnlock()
	cp, ok := m.cps[id]
	if !ok {
		return nil, ErrNotFound
	}
	return cp, nil
}

func (m *MockCheckpointRepository) ListByJobID(jobID string) ([]*domain.Checkpoint, error) {
	if m.Fail {
		return nil, ErrInternal
	}
	m.mu.RLock()
	defer m.mu.RUnlock()
	var result []*domain.Checkpoint
	for _, cp := range m.cps {
		if cp.JobID == jobID {
			result = append(result, cp)
		}
	}
	return result, nil
}

func (m *MockCheckpointRepository) UpdateStatus(id string, status domain.CheckpointStatus, notes string) error {
	if m.Fail {
		return ErrInternal
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	cp, ok := m.cps[id]
	if !ok {
		return ErrNotFound
	}
	cp.Status = status
	cp.Notes = notes
	return nil
}
