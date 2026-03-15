package series

import (
	"fmt"
	"strings"
)

// BuildContextWindow returns the sliding window context for the next episode.
// windowSize controls how many recent episodes to include fully.
// The GlobalSummary is always included regardless of windowSize.
func BuildContextWindow(m *SeriesMemory, windowSize int) SeriesContextWindow {
	if m == nil {
		return SeriesContextWindow{}
	}

	w := SeriesContextWindow{
		GlobalSummary: m.GlobalSummary,
	}

	episodes := m.Episodes
	if len(episodes) > windowSize && windowSize > 0 {
		episodes = episodes[len(episodes)-windowSize:]
	}

	// Make a copy so callers cannot mutate internal state
	if len(episodes) > 0 {
		w.RecentEpisodes = make([]EpisodeMemory, len(episodes))
		copy(w.RecentEpisodes, episodes)
	}

	return w
}

// FormatContextPrompt formats a SeriesContextWindow as a string block to prepend to story prompts.
//
// Format:
//
//	[SERIES_CONTEXT]
//	Global: <global_summary or "(none yet)">
//	Recent episodes:
//	  Ep1: event1; event2
//	  Ep2: event1; event2
//	[/SERIES_CONTEXT]
func FormatContextPrompt(w SeriesContextWindow) string {
	var sb strings.Builder

	sb.WriteString("[SERIES_CONTEXT]\n")

	globalSummary := w.GlobalSummary
	if globalSummary == "" {
		globalSummary = "(none yet)"
	}
	fmt.Fprintf(&sb, "Global: %s\n", globalSummary)

	if len(w.RecentEpisodes) > 0 {
		sb.WriteString("Recent episodes:\n")
		for _, ep := range w.RecentEpisodes {
			events := strings.Join(ep.KeyEvents, "; ")
			if events == "" {
				events = "(no events)"
			}
			fmt.Fprintf(&sb, "  Ep%d: %s\n", ep.Episode, events)
		}
	}

	sb.WriteString("[/SERIES_CONTEXT]")
	return sb.String()
}
