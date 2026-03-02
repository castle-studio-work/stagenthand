// Package store handles SQLite persistence via gorm.
package store

import (
	"errors"

	"github.com/baochen10luo/stagenthand/internal/domain"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// ErrNotFound is returned when a requested record does not exist.
var ErrNotFound = errors.New("record not found")

// DB wraps a gorm.DB instance.
type DB = gorm.DB

// New opens (or creates) a SQLite database at the given DSN and auto-migrates schemas.
func New(dsn string) (*gorm.DB, error) {
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	if err := db.AutoMigrate(&jobRecord{}, &checkpointRecord{}); err != nil {
		return nil, err
	}
	return db, nil
}

// --- ORM models (separate from domain to keep domain zero-dependency) ---

type jobRecord struct {
	ID        string `gorm:"primaryKey"`
	ProjectID string `gorm:"index"`
	Type      string
	Status    string
	Error     string
	CreatedAt int64 // unix nano
	UpdatedAt int64
}

type checkpointRecord struct {
	ID        string `gorm:"primaryKey"`
	JobID     string `gorm:"index"`
	Stage     string
	Status    string
	Notes     string
	CreatedAt int64
	UpdatedAt int64
}

// toJob converts a jobRecord to a domain.Job.
func (r *jobRecord) toDomain() *domain.Job {
	return &domain.Job{
		ID:        r.ID,
		ProjectID: r.ProjectID,
		Type:      r.Type,
		Status:    domain.JobStatus(r.Status),
		Error:     r.Error,
	}
}
