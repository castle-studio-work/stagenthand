package series

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"time"
)

// FileRepository persists SeriesMemory as JSON in a local file.
// Writes are atomic: data is written to a temp file then renamed.
type FileRepository struct {
	path string
}

// NewFileRepository constructs a FileRepository backed by the given file path.
func NewFileRepository(path string) *FileRepository {
	return &FileRepository{path: path}
}

// Load reads the SeriesMemory from disk.
// If the file does not exist, an empty (non-nil) SeriesMemory is returned.
func (r *FileRepository) Load(_ context.Context) (*SeriesMemory, error) {
	data, err := os.ReadFile(r.path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return &SeriesMemory{}, nil
		}
		return nil, err
	}
	var m SeriesMemory
	if err := json.Unmarshal(data, &m); err != nil {
		return nil, err
	}
	return &m, nil
}

// Save atomically writes m to the file.
func (r *FileRepository) Save(_ context.Context, m *SeriesMemory) error {
	m.UpdatedAt = time.Now()
	data, err := json.Marshal(m)
	if err != nil {
		return err
	}

	// Atomic write: write to temp, then rename
	dir := filepath.Dir(r.path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	tmp, err := os.CreateTemp(dir, "series_memory_*.tmp")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()

	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		os.Remove(tmpName)
		return err
	}
	if err := tmp.Close(); err != nil {
		os.Remove(tmpName)
		return err
	}

	return os.Rename(tmpName, r.path)
}

// Append loads the current memory, appends the given episode, and saves.
func (r *FileRepository) Append(ctx context.Context, ep EpisodeMemory) error {
	m, err := r.Load(ctx)
	if err != nil {
		return err
	}
	m.Episodes = append(m.Episodes, ep)
	return r.Save(ctx, m)
}
