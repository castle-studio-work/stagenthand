package series

import "time"

// CharacterSnapshot captures a character's state at the end of an episode.
type CharacterSnapshot struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Motivation  string `json:"motivation"`
	State       string `json:"state"`
}

// EpisodeMemory holds the distilled memory of one completed episode.
type EpisodeMemory struct {
	Episode    int                 `json:"episode"`
	KeyEvents  []string            `json:"key_events"`
	Characters []CharacterSnapshot `json:"characters"`
	WorldFacts []string            `json:"world_facts"`
}

// SeriesMemory accumulates memories across all episodes of a series.
type SeriesMemory struct {
	SeriesTitle   string          `json:"series_title"`
	GlobalSummary string          `json:"global_summary"` // compressed across all episodes
	Episodes      []EpisodeMemory `json:"episodes"`
	UpdatedAt     time.Time       `json:"updated_at"`
}

// SeriesContextWindow is the context injected into the next episode's prompt.
type SeriesContextWindow struct {
	GlobalSummary  string          `json:"global_summary"`
	RecentEpisodes []EpisodeMemory `json:"recent_episodes"`
}
