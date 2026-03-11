package video

import "context"

// Client defines the interface for video generation providers.
type Client interface {
	GenerateVideo(ctx context.Context, imageURL string, prompt string) ([]byte, error)
}
