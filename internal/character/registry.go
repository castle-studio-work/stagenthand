// Package character manages character reference images for visual consistency.
package character

import (
	"context"
	"time"
)

// Registry manages character reference images for visual consistency.
type Registry interface {
	// Register saves an image as the canonical reference for a character name.
	// Returns the file path where the image was saved.
	Register(ctx context.Context, name string, imageBytes []byte) (string, error)
	// Lookup returns the file path of a character's reference image, or "" if not found.
	Lookup(ctx context.Context, name string) (string, error)
	// List returns all registered character names.
	List(ctx context.Context) ([]string, error)
}

// CharacterMeta holds metadata about a registered character.
type CharacterMeta struct {
	Name      string    `json:"name"`
	ImagePath string    `json:"image_path"`
	CreatedAt time.Time `json:"created_at"`
}
