package store

import (
	"time"

	"github.com/baochen10luo/stagenthand/internal/domain"
	"gorm.io/gorm"
)

// CheckpointRepository defines persistence operations for HITL checkpoints.
type CheckpointRepository interface {
	Create(cp *domain.Checkpoint) error
	GetByID(id string) (*domain.Checkpoint, error)
	ListByJobID(jobID string) ([]*domain.Checkpoint, error)
	UpdateStatus(id string, status domain.CheckpointStatus, notes string) error
}

// GormCheckpointRepository implements CheckpointRepository using gorm + SQLite.
type GormCheckpointRepository struct {
	db *gorm.DB
}

// NewGormCheckpointRepository constructs a GormCheckpointRepository.
func NewGormCheckpointRepository(db *gorm.DB) CheckpointRepository {
	return &GormCheckpointRepository{db: db}
}

func (r *GormCheckpointRepository) Create(cp *domain.Checkpoint) error {
	rec := &checkpointRecord{
		ID:        cp.ID,
		JobID:     cp.JobID,
		Stage:     string(cp.Stage),
		Status:    string(cp.Status),
		Notes:     cp.Notes,
		CreatedAt: cp.CreatedAt.UnixNano(),
		UpdatedAt: cp.UpdatedAt.UnixNano(),
	}
	return r.db.Create(rec).Error
}

func (r *GormCheckpointRepository) GetByID(id string) (*domain.Checkpoint, error) {
	var rec checkpointRecord
	result := r.db.First(&rec, "id = ?", id)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, ErrNotFound
		}
		return nil, result.Error
	}
	return rec.toCheckpointDomain(), nil
}

func (r *GormCheckpointRepository) ListByJobID(jobID string) ([]*domain.Checkpoint, error) {
	var recs []checkpointRecord
	if err := r.db.Where("job_id = ?", jobID).Find(&recs).Error; err != nil {
		return nil, err
	}
	result := make([]*domain.Checkpoint, len(recs))
	for i := range recs {
		result[i] = recs[i].toCheckpointDomain()
	}
	return result, nil
}

func (r *GormCheckpointRepository) UpdateStatus(id string, status domain.CheckpointStatus, notes string) error {
	result := r.db.Model(&checkpointRecord{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"status":     string(status),
			"notes":      notes,
			"updated_at": time.Now().UnixNano(),
		})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *checkpointRecord) toCheckpointDomain() *domain.Checkpoint {
	return &domain.Checkpoint{
		ID:     r.ID,
		JobID:  r.JobID,
		Stage:  domain.CheckpointStage(r.Stage),
		Status: domain.CheckpointStatus(r.Status),
		Notes:  r.Notes,
	}
}
