package store

import (
	"testing"

	"github.com/baochen10luo/stagenthand/internal/domain"
	"github.com/stretchr/testify/assert"
)

func TestMockJobRepository(t *testing.T) {
	repo := NewMockJobRepository()

	t.Run("Create and GetByID", func(t *testing.T) {
		job := &domain.Job{
			ID: "job-1",
		}
		err := repo.Create(job)
		assert.NoError(t, err)

		got, err := repo.GetByID("job-1")
		assert.NoError(t, err)
		assert.Equal(t, job, got)
	})

	t.Run("GetByID NotFound", func(t *testing.T) {
		_, err := repo.GetByID("non-existent")
		assert.ErrorIs(t, err, ErrNotFound)
	})

	t.Run("UpdateStatus", func(t *testing.T) {
		job := &domain.Job{
			ID: "job-2",
		}
		repo.Create(job)

		err := repo.UpdateStatus("job-2", domain.JobStatusCompleted, "no error")
		assert.NoError(t, err)

		got, _ := repo.GetByID("job-2")
		assert.Equal(t, domain.JobStatusCompleted, got.Status)
		assert.Equal(t, "no error", got.Error)
	})

	t.Run("UpdateStatus NotFound", func(t *testing.T) {
		err := repo.UpdateStatus("non-existent", domain.JobStatusCompleted, "")
		assert.ErrorIs(t, err, ErrNotFound)
	})
}

func TestMockCheckpointRepository(t *testing.T) {
	repo := NewMockCheckpointRepository()

	t.Run("Create and GetByID", func(t *testing.T) {
		cp := &domain.Checkpoint{
			ID: "cp-1",
		}
		err := repo.Create(cp)
		assert.NoError(t, err)

		got, err := repo.GetByID("cp-1")
		assert.NoError(t, err)
		assert.Equal(t, cp, got)
	})

	t.Run("GetByID NotFound", func(t *testing.T) {
		_, err := repo.GetByID("non-existent")
		assert.ErrorIs(t, err, ErrNotFound)
	})

	t.Run("ListByJobID", func(t *testing.T) {
		cp1 := &domain.Checkpoint{ID: "cp-2", JobID: "job-1"}
		cp2 := &domain.Checkpoint{ID: "cp-3", JobID: "job-1"}
		cp3 := &domain.Checkpoint{ID: "cp-4", JobID: "job-2"}

		repo.Create(cp1)
		repo.Create(cp2)
		repo.Create(cp3)

		list, err := repo.ListByJobID("job-1")
		assert.NoError(t, err)
		assert.Len(t, list, 2)
		// Map iteration order is random, so check existence
		ids := make(map[string]bool)
		for _, cp := range list {
			ids[cp.ID] = true
		}
		assert.True(t, ids["cp-2"])
		assert.True(t, ids["cp-3"])
	})

	t.Run("UpdateStatus", func(t *testing.T) {
		cp := &domain.Checkpoint{
			ID: "cp-5",
		}
		repo.Create(cp)

		err := repo.UpdateStatus("cp-5", domain.CheckpointStatusApproved, "looks good")
		assert.NoError(t, err)

		got, _ := repo.GetByID("cp-5")
		assert.Equal(t, domain.CheckpointStatusApproved, got.Status)
		assert.Equal(t, "looks good", got.Notes)
	})

	t.Run("UpdateStatus NotFound", func(t *testing.T) {
		err := repo.UpdateStatus("non-existent", domain.CheckpointStatusApproved, "")
		assert.ErrorIs(t, err, ErrNotFound)
	})
}
