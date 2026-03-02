package store

import (
	"time"

	"github.com/baochen10luo/stagenthand/internal/domain"
	"gorm.io/gorm"
)

// JobRepository defines persistence operations for pipeline jobs.
type JobRepository interface {
	Create(job *domain.Job) error
	GetByID(id string) (*domain.Job, error)
	UpdateStatus(id string, status domain.JobStatus, errMsg string) error
}

// GormJobRepository implements JobRepository using gorm + SQLite.
type GormJobRepository struct {
	db *gorm.DB
}

// NewGormJobRepository constructs a GormJobRepository.
func NewGormJobRepository(db *gorm.DB) JobRepository {
	return &GormJobRepository{db: db}
}

func (r *GormJobRepository) Create(job *domain.Job) error {
	rec := &jobRecord{
		ID:        job.ID,
		ProjectID: job.ProjectID,
		Type:      job.Type,
		Status:    string(job.Status),
		Error:     job.Error,
		CreatedAt: job.CreatedAt.UnixNano(),
		UpdatedAt: job.UpdatedAt.UnixNano(),
	}
	return r.db.Create(rec).Error
}

func (r *GormJobRepository) GetByID(id string) (*domain.Job, error) {
	var rec jobRecord
	result := r.db.First(&rec, "id = ?", id)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, ErrNotFound
		}
		return nil, result.Error
	}
	return rec.toDomain(), nil
}

func (r *GormJobRepository) UpdateStatus(id string, status domain.JobStatus, errMsg string) error {
	result := r.db.Model(&jobRecord{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"status":     string(status),
			"error":      errMsg,
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
