package character

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// FileRegistry implements Registry using the local filesystem.
// Characters are stored under <rootDir>/characters/<name>/ref.png.
type FileRegistry struct {
	rootDir string
}

// NewFileRegistry creates a FileRegistry backed by the given root directory.
func NewFileRegistry(rootDir string) *FileRegistry {
	return &FileRegistry{rootDir: rootDir}
}

// characterDir returns the directory for the given character name.
func (r *FileRegistry) characterDir(name string) string {
	return filepath.Join(r.rootDir, "characters", name)
}

// Register saves imageBytes as the canonical reference image for the character.
// Overwrites any previously registered image.
func (r *FileRegistry) Register(_ context.Context, name string, imageBytes []byte) (string, error) {
	dir := r.characterDir(name)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", fmt.Errorf("character.Register: create dir %s: %w", dir, err)
	}

	imgPath := filepath.Join(dir, "ref.png")
	if err := os.WriteFile(imgPath, imageBytes, 0644); err != nil {
		return "", fmt.Errorf("character.Register: write image %s: %w", imgPath, err)
	}

	meta := CharacterMeta{
		Name:      name,
		ImagePath: imgPath,
		CreatedAt: time.Now(),
	}
	metaPath := filepath.Join(dir, "meta.json")
	metaBytes, err := json.Marshal(meta)
	if err != nil {
		return "", fmt.Errorf("character.Register: marshal meta: %w", err)
	}
	if err := os.WriteFile(metaPath, metaBytes, 0644); err != nil {
		return "", fmt.Errorf("character.Register: write meta %s: %w", metaPath, err)
	}

	return imgPath, nil
}

// Lookup returns the image path for the given character, or "" if not found.
func (r *FileRegistry) Lookup(_ context.Context, name string) (string, error) {
	imgPath := filepath.Join(r.characterDir(name), "ref.png")
	info, err := os.Stat(imgPath)
	if os.IsNotExist(err) || (err == nil && info.Size() == 0) {
		return "", nil
	}
	if err != nil {
		return "", fmt.Errorf("character.Lookup: stat %s: %w", imgPath, err)
	}
	return imgPath, nil
}

// List returns the names of all registered characters.
func (r *FileRegistry) List(_ context.Context) ([]string, error) {
	baseDir := filepath.Join(r.rootDir, "characters")
	entries, err := os.ReadDir(baseDir)
	if os.IsNotExist(err) {
		return []string{}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("character.List: read dir %s: %w", baseDir, err)
	}

	var names []string
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		// Verify the character has a valid ref.png
		refPath := filepath.Join(baseDir, e.Name(), "ref.png")
		if info, statErr := os.Stat(refPath); statErr == nil && info.Size() > 0 {
			names = append(names, e.Name())
		}
	}
	return names, nil
}
