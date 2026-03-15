package render

// VideoFormat defines the output video aspect ratio.
type VideoFormat string

const (
	VideoFormatLandscape VideoFormat = "landscape" // 1024×576 (default, 16:9)
	VideoFormatPortrait  VideoFormat = "portrait"  // 576×1024 (9:16, TikTok/Reels)
)

// Dimensions returns (width, height) for the given format.
// Unknown formats default to landscape.
func (f VideoFormat) Dimensions() (width, height int) {
	if f == VideoFormatPortrait {
		return 576, 1024
	}
	return 1024, 576
}

// IsPortrait returns true when height > width.
func (f VideoFormat) IsPortrait() bool {
	w, h := f.Dimensions()
	return h > w
}
