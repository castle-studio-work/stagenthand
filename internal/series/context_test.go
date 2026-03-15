package series_test

import (
	"strings"
	"testing"

	"github.com/baochen10luo/stagenthand/internal/series"
)

func makeTestMemory(n int) *series.SeriesMemory {
	m := &series.SeriesMemory{
		GlobalSummary: "The grand saga",
	}
	for i := 1; i <= n; i++ {
		m.Episodes = append(m.Episodes, series.EpisodeMemory{
			Episode:   i,
			KeyEvents: []string{"event A", "event B"},
		})
	}
	return m
}

func TestBuildContextWindow_SlidingWindow(t *testing.T) {
	m := makeTestMemory(5)
	w := series.BuildContextWindow(m, 3)

	if len(w.RecentEpisodes) != 3 {
		t.Fatalf("RecentEpisodes length: got %d, want 3", len(w.RecentEpisodes))
	}
	// Should contain episodes 3, 4, 5
	if w.RecentEpisodes[0].Episode != 3 {
		t.Errorf("RecentEpisodes[0].Episode: got %d, want 3", w.RecentEpisodes[0].Episode)
	}
	if w.RecentEpisodes[2].Episode != 5 {
		t.Errorf("RecentEpisodes[2].Episode: got %d, want 5", w.RecentEpisodes[2].Episode)
	}
}

func TestBuildContextWindow_FewerThanWindow(t *testing.T) {
	m := makeTestMemory(2)
	w := series.BuildContextWindow(m, 3)

	if len(w.RecentEpisodes) != 2 {
		t.Fatalf("RecentEpisodes length: got %d, want 2", len(w.RecentEpisodes))
	}
	if w.RecentEpisodes[0].Episode != 1 {
		t.Errorf("RecentEpisodes[0].Episode: got %d, want 1", w.RecentEpisodes[0].Episode)
	}
	if w.RecentEpisodes[1].Episode != 2 {
		t.Errorf("RecentEpisodes[1].Episode: got %d, want 2", w.RecentEpisodes[1].Episode)
	}
}

func TestBuildContextWindow_AlwaysGlobalSummary(t *testing.T) {
	m := &series.SeriesMemory{
		GlobalSummary: "Always present",
		Episodes:      []series.EpisodeMemory{},
	}
	w := series.BuildContextWindow(m, 3)

	if w.GlobalSummary != "Always present" {
		t.Errorf("GlobalSummary: got %q, want %q", w.GlobalSummary, "Always present")
	}
}

func TestBuildContextWindow_EmptyMemory(t *testing.T) {
	m := &series.SeriesMemory{}
	w := series.BuildContextWindow(m, 3)

	if w.GlobalSummary != "" {
		t.Errorf("expected empty GlobalSummary, got %q", w.GlobalSummary)
	}
	if len(w.RecentEpisodes) != 0 {
		t.Errorf("expected no RecentEpisodes, got %d", len(w.RecentEpisodes))
	}
}

func TestBuildContextWindow_NilMemory(t *testing.T) {
	w := series.BuildContextWindow(nil, 3)

	if w.GlobalSummary != "" {
		t.Errorf("expected empty GlobalSummary for nil memory, got %q", w.GlobalSummary)
	}
}

func TestFormatContextPrompt_ContainsMarkers(t *testing.T) {
	w := series.SeriesContextWindow{
		GlobalSummary: "A hero rises",
		RecentEpisodes: []series.EpisodeMemory{
			{Episode: 1, KeyEvents: []string{"battle won"}},
		},
	}

	result := series.FormatContextPrompt(w)

	if !strings.Contains(result, "[SERIES_CONTEXT]") {
		t.Errorf("missing [SERIES_CONTEXT] opening marker")
	}
	if !strings.Contains(result, "[/SERIES_CONTEXT]") {
		t.Errorf("missing [/SERIES_CONTEXT] closing marker")
	}
	if !strings.Contains(result, "A hero rises") {
		t.Errorf("missing global summary text")
	}
	if !strings.Contains(result, "Ep1:") {
		t.Errorf("missing episode reference")
	}
	if !strings.Contains(result, "battle won") {
		t.Errorf("missing key event text")
	}
}

func TestFormatContextPrompt_EmptyMemory(t *testing.T) {
	// Should not panic with empty window
	w := series.SeriesContextWindow{}

	var result string
	// Ensure no panic
	func() {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("FormatContextPrompt panicked: %v", r)
			}
		}()
		result = series.FormatContextPrompt(w)
	}()

	if !strings.Contains(result, "[SERIES_CONTEXT]") {
		t.Errorf("missing [SERIES_CONTEXT] marker in empty window output")
	}
	if !strings.Contains(result, "(none yet)") {
		t.Errorf("expected '(none yet)' for empty global summary, got: %q", result)
	}
}

func TestFormatContextPrompt_MultipleEpisodes(t *testing.T) {
	w := series.SeriesContextWindow{
		GlobalSummary: "Saga",
		RecentEpisodes: []series.EpisodeMemory{
			{Episode: 3, KeyEvents: []string{"attack", "retreat"}},
			{Episode: 4, KeyEvents: []string{"peace treaty"}},
		},
	}

	result := series.FormatContextPrompt(w)

	if !strings.Contains(result, "Ep3:") {
		t.Errorf("missing Ep3 reference")
	}
	if !strings.Contains(result, "attack; retreat") {
		t.Errorf("missing events for Ep3")
	}
	if !strings.Contains(result, "Ep4:") {
		t.Errorf("missing Ep4 reference")
	}
}
