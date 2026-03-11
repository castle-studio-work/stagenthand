package image

import "context"

// Client is the interface that wraps basic image generation methods.
// It adheres to DIP to decouple the application logic from specific providers
// (e.g. NanoBanana, Z-Image).
type Client interface {
	// GenerateImage sends a prompt and an optional list of character reference image paths,
	// and returns the generated image as raw bytes (which could be saved locally or processed further).
	// Because shand prefers local artifacts, another system component will handle writing these bytes to disk.
	GenerateImage(ctx context.Context, prompt string, characterRefs []string) ([]byte, error)
}
