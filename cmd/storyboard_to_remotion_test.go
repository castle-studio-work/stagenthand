package cmd

import (
	"encoding/json"
	"testing"

	"github.com/baochen10luo/stagenthand/internal/domain"
)

func TestParseAndConvert_Storyboard(t *testing.T) {
	sb := domain.Storyboard{
		ProjectID: "proj-cmd",
		Episode:   1,
		Scenes: []domain.Scene{
			{
				Number: 1,
				Panels: []domain.Panel{
					{SceneNumber: 1, PanelNumber: 1, Description: "test panel", Dialogue: "測試", ImageURL: "https://example.com/img.png", DurationSec: 2.5},
				},
			},
		},
	}
	raw, _ := json.Marshal(sb)

	props, err := parseAndConvert(raw, 1024, 576)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if props.ProjectID != "proj-cmd" {
		t.Errorf("ProjectID: want proj-cmd, got %s", props.ProjectID)
	}
	if len(props.Panels) != 1 {
		t.Errorf("Panels: want 1, got %d", len(props.Panels))
	}
}

func TestParseAndConvert_PanelArray(t *testing.T) {
	panels := []domain.Panel{
		{SceneNumber: 1, PanelNumber: 1, Description: "p1", Dialogue: "Hello", ImageURL: "https://example.com/a.png", DurationSec: 3.0},
		{SceneNumber: 1, PanelNumber: 2, Description: "p2", Dialogue: "World", ImageURL: "https://example.com/b.png", DurationSec: 2.0},
	}
	raw, _ := json.Marshal(panels)

	props, err := parseAndConvert(raw, 1920, 1080)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(props.Panels) != 2 {
		t.Errorf("Panels: want 2, got %d", len(props.Panels))
	}
	if props.Width != 1920 || props.Height != 1080 {
		t.Errorf("dimensions: want 1920x1080, got %dx%d", props.Width, props.Height)
	}
}

func TestParseAndConvert_InvalidInput(t *testing.T) {
	_, err := parseAndConvert([]byte(`{"not":"valid"}`), 1024, 576)
	if err == nil {
		t.Error("expected error for unrecognized input, got nil")
	}
}
