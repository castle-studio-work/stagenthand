package series_test

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/baochen10luo/stagenthand/internal/series"
)

func TestSeriesMemory_JSONRoundTrip(t *testing.T) {
	original := series.SeriesMemory{
		SeriesTitle:   "Test Series",
		GlobalSummary: "A tale of two worlds",
		Episodes: []series.EpisodeMemory{
			{
				Episode:    1,
				KeyEvents:  []string{"hero arrives", "meets mentor"},
				Characters: []series.CharacterSnapshot{{Name: "Hero", Description: "brave", Motivation: "save world", State: "hopeful"}},
				WorldFacts: []string{"magic exists"},
			},
		},
		UpdatedAt: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
	}

	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	var decoded series.SeriesMemory
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if decoded.SeriesTitle != original.SeriesTitle {
		t.Errorf("SeriesTitle mismatch: got %q, want %q", decoded.SeriesTitle, original.SeriesTitle)
	}
	if decoded.GlobalSummary != original.GlobalSummary {
		t.Errorf("GlobalSummary mismatch: got %q, want %q", decoded.GlobalSummary, original.GlobalSummary)
	}
	if len(decoded.Episodes) != 1 {
		t.Fatalf("Episodes length: got %d, want 1", len(decoded.Episodes))
	}
	ep := decoded.Episodes[0]
	if ep.Episode != 1 {
		t.Errorf("Episode number: got %d, want 1", ep.Episode)
	}
	if len(ep.KeyEvents) != 2 {
		t.Errorf("KeyEvents length: got %d, want 2", len(ep.KeyEvents))
	}
	if ep.KeyEvents[0] != "hero arrives" {
		t.Errorf("KeyEvents[0]: got %q, want %q", ep.KeyEvents[0], "hero arrives")
	}
	if len(ep.Characters) != 1 {
		t.Errorf("Characters length: got %d, want 1", len(ep.Characters))
	}
	if ep.Characters[0].Name != "Hero" {
		t.Errorf("Character Name: got %q, want %q", ep.Characters[0].Name, "Hero")
	}
}

func TestEpisodeMemory_EmptySlices(t *testing.T) {
	// nil slices should marshal as [] in JSON (not null)
	ep := series.EpisodeMemory{
		Episode:    1,
		KeyEvents:  []string{},
		Characters: []series.CharacterSnapshot{},
		WorldFacts: []string{},
	}

	data, err := json.Marshal(ep)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	// Verify [] not null
	var raw map[string]interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatalf("Unmarshal raw failed: %v", err)
	}

	checkArray := func(key string) {
		v, ok := raw[key]
		if !ok {
			t.Errorf("key %q missing from JSON", key)
			return
		}
		arr, ok := v.([]interface{})
		if !ok {
			t.Errorf("key %q: expected array, got %T", key, v)
			return
		}
		if len(arr) != 0 {
			t.Errorf("key %q: expected empty array, got length %d", key, len(arr))
		}
	}

	checkArray("key_events")
	checkArray("characters")
	checkArray("world_facts")
}
