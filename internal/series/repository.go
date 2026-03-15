package series

import "context"

// Repository persists and retrieves SeriesMemory.
type Repository interface {
	// Load returns the stored SeriesMemory. Returns an empty SeriesMemory (not nil) if not found.
	Load(ctx context.Context) (*SeriesMemory, error)
	// Save writes the full SeriesMemory to persistent storage.
	Save(ctx context.Context, m *SeriesMemory) error
	// Append loads the current memory, appends the episode, and saves.
	Append(ctx context.Context, ep EpisodeMemory) error
}
