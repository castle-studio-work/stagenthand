package render_test

import (
	"testing"

	"github.com/baochen10luo/stagenthand/internal/render"
)

func TestVideoFormat_Dimensions(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		format      render.VideoFormat
		wantWidth   int
		wantHeight  int
	}{
		{
			name:       "landscape returns 1024x576",
			format:     render.VideoFormatLandscape,
			wantWidth:  1024,
			wantHeight: 576,
		},
		{
			name:       "portrait returns 576x1024",
			format:     render.VideoFormatPortrait,
			wantWidth:  576,
			wantHeight: 1024,
		},
		{
			name:       "unknown defaults to landscape 1024x576",
			format:     render.VideoFormat("unknown"),
			wantWidth:  1024,
			wantHeight: 576,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			w, h := tc.format.Dimensions()
			if w != tc.wantWidth || h != tc.wantHeight {
				t.Errorf("Dimensions() = (%d, %d), want (%d, %d)", w, h, tc.wantWidth, tc.wantHeight)
			}
		})
	}
}

func TestVideoFormat_IsPortrait(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		format render.VideoFormat
		want   bool
	}{
		{
			name:   "landscape is not portrait",
			format: render.VideoFormatLandscape,
			want:   false,
		},
		{
			name:   "portrait is portrait",
			format: render.VideoFormatPortrait,
			want:   true,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := tc.format.IsPortrait()
			if got != tc.want {
				t.Errorf("IsPortrait() = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestVideoFormat_String(t *testing.T) {
	t.Parallel()

	if string(render.VideoFormatLandscape) != "landscape" {
		t.Errorf("VideoFormatLandscape string = %q, want %q", render.VideoFormatLandscape, "landscape")
	}
	if string(render.VideoFormatPortrait) != "portrait" {
		t.Errorf("VideoFormatPortrait string = %q, want %q", render.VideoFormatPortrait, "portrait")
	}
}
