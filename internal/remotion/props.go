package remotion

import (
	"fmt"
	"strings"

	"github.com/baochen10luo/stagenthand/internal/domain"
	"github.com/baochen10luo/stagenthand/internal/render"
)

const defaultPanelDurationSec = 3.0

// StoryboardToProps converts a Storyboard (with nested Scenes and Panels)
// into a flat RemotionProps ready for the Remotion template.
// All panels are extracted in scene order, preserving panel order within each scene.
// Panels with zero DurationSec are assigned the default duration.
func StoryboardToProps(sb domain.Storyboard, width, height, fps int) domain.RemotionProps {
	panels := flattenPanels(sb.Scenes)
	return domain.RemotionProps{
		ProjectID: sb.ProjectID,
		Title:     "Generated Drama",
		BGMURL:    normalizePath(sb.BGMURL, sb.ProjectID),
		Panels:    panels,
		FPS:       fps,
		Width:     width,
		Height:    height,
	}
}

// PanelsToProps converts a flat []Panel directly into RemotionProps.
// Useful when the pipeline has already extracted panels from the storyboard.
// width and height are explicit overrides; pass 0,0 to derive from format.
func PanelsToProps(projectID string, panels []domain.Panel, width, height, fps int, bgmURL string, directives *domain.Directives) domain.RemotionProps {
	return PanelsToPropsWithFormat(projectID, panels, width, height, fps, bgmURL, directives, render.VideoFormatLandscape)
}

// PanelsToPropsWithFormat converts a flat []Panel into RemotionProps using the given VideoFormat
// to set canvas dimensions. Explicit width/height > 0 override the format dimensions.
func PanelsToPropsWithFormat(projectID string, panels []domain.Panel, width, height, fps int, bgmURL string, directives *domain.Directives, format render.VideoFormat) domain.RemotionProps {
	fw, fh := format.Dimensions()
	if width == 0 {
		width = fw
	}
	if height == 0 {
		height = fh
	}
	normalized := make([]domain.Panel, len(panels))
	for i, p := range panels {
		p = withDefaultDuration(p)
		p.ImageURL = normalizePath(p.ImageURL, projectID)
		p.AudioURL = normalizePath(p.AudioURL, projectID)
		normalized[i] = p
	}
	return domain.RemotionProps{
		ProjectID:  projectID,
		Title:      "Generated Drama",
		BGMURL:     normalizePath(bgmURL, projectID),
		Directives: directives,
		Panels:     normalized,
		FPS:        fps,
		Width:      width,
		Height:     height,
	}
}

func normalizePath(path, projectID string) string {
	if path == "" || strings.HasPrefix(path, "/shand/") {
		return path
	}

	// Look for the "projects/<project_id>/" segment in the absolute path
	marker := fmt.Sprintf("projects/%s/", projectID)
	idx := strings.Index(path, marker)
	if idx != -1 {
		// Convert to virtual path: /shand/<project_id>/...
		return "/shand/" + projectID + "/" + path[idx+len(marker):]
	}

	return path
}


// flattenPanels extracts all panels from scenes in order, applying default durations.
func flattenPanels(scenes []domain.Scene) []domain.Panel {
	var out []domain.Panel
	for _, scene := range scenes {
		for _, p := range scene.Panels {
			out = append(out, withDefaultDuration(p))
		}
	}
	return out
}

// withDefaultDuration ensures a Panel has a non-zero DurationSec.
func withDefaultDuration(p domain.Panel) domain.Panel {
	if p.DurationSec == 0 {
		p.DurationSec = defaultPanelDurationSec
	}
	return p
}
