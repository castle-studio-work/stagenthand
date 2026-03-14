package remotion_test

import (
	"testing"

	"github.com/baochen10luo/stagenthand/internal/domain"
	"github.com/baochen10luo/stagenthand/internal/remotion"
)

func TestStoryboardToRemotionProps_Basic(t *testing.T) {
	sb := domain.Storyboard{
		ProjectID: "proj-123",
		Episode:   1,
		Scenes: []domain.Scene{
			{
				Number:      1,
				Description: "Opening",
				Panels: []domain.Panel{
					{SceneNumber: 1, PanelNumber: 1, Description: "hero walks in", Dialogue: "你好", ImageURL: "https://example.com/p1.png", DurationSec: 3.0},
					{SceneNumber: 1, PanelNumber: 2, Description: "hero smiles", Dialogue: "今天天氣真好", ImageURL: "https://example.com/p2.png", DurationSec: 2.5},
				},
			},
			{
				Number:      2,
				Description: "Conflict",
				Panels: []domain.Panel{
					{SceneNumber: 2, PanelNumber: 1, Description: "boss frowns", Dialogue: "你遲到了", ImageURL: "https://example.com/p3.png", DurationSec: 3.5},
				},
			},
		},
	}

	props := remotion.StoryboardToProps(sb, 1024, 576, 24)

	if props.ProjectID != "proj-123" {
		t.Errorf("ProjectID: want proj-123, got %s", props.ProjectID)
	}
	if props.FPS != 24 {
		t.Errorf("FPS: want 24, got %d", props.FPS)
	}
	if props.Width != 1024 || props.Height != 576 {
		t.Errorf("dimensions: want 1024x576, got %dx%d", props.Width, props.Height)
	}
	if len(props.Panels) != 3 {
		t.Errorf("Panels: want 3 (all panels flattened), got %d", len(props.Panels))
	}
	// Verify panel order is preserved
	if props.Panels[0].Dialogue != "你好" {
		t.Errorf("first panel dialogue: want 你好, got %s", props.Panels[0].Dialogue)
	}
	if props.Panels[2].Dialogue != "你遲到了" {
		t.Errorf("third panel dialogue: want 你遲到了, got %s", props.Panels[2].Dialogue)
	}
}

func TestStoryboardToRemotionProps_EmptyStoryboard(t *testing.T) {
	sb := domain.Storyboard{ProjectID: "empty-proj"}
	props := remotion.StoryboardToProps(sb, 1024, 576, 24)

	if len(props.Panels) != 0 {
		t.Errorf("empty storyboard: want 0 panels, got %d", len(props.Panels))
	}
}

func TestStoryboardToRemotionProps_DefaultDuration(t *testing.T) {
	// Panels with zero DurationSec should default to 3.0
	sb := domain.Storyboard{
		ProjectID: "proj-dur",
		Scenes: []domain.Scene{
			{
				Number: 1,
				Panels: []domain.Panel{
					{SceneNumber: 1, PanelNumber: 1, Description: "scene", Dialogue: "test", ImageURL: "https://example.com/img.png", DurationSec: 0},
				},
			},
		},
	}
	props := remotion.StoryboardToProps(sb, 1024, 576, 24)

	if props.Panels[0].DurationSec != 3.0 {
		t.Errorf("default duration: want 3.0, got %f", props.Panels[0].DurationSec)
	}
}

func TestPanelsToRemotionProps_DirectArray(t *testing.T) {
	panels := []domain.Panel{
		{SceneNumber: 1, PanelNumber: 1, Description: "p1", Dialogue: "Hello", ImageURL: "https://example.com/a.png", DurationSec: 2.0},
		{SceneNumber: 1, PanelNumber: 2, Description: "p2", Dialogue: "World", ImageURL: "https://example.com/b.png", DurationSec: 4.0},
	}

	props := remotion.PanelsToProps("proj-direct", panels, 1920, 1080, 30)

	if len(props.Panels) != 2 {
		t.Fatalf("want 2 panels, got %d", len(props.Panels))
	}
	if props.Width != 1920 || props.FPS != 30 {
		t.Errorf("config not propagated: width=%d fps=%d", props.Width, props.FPS)
	}
}
