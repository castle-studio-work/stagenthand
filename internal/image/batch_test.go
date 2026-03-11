package image_test

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/baochen10luo/stagenthand/internal/domain"
	"github.com/baochen10luo/stagenthand/internal/image"
	"github.com/stretchr/testify/assert"
)

func TestGenerateBatch(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	t.Run("success", func(t *testing.T) {
		client := &image.MockClient{
			GenerateImageFunc: func(ctx context.Context, prompt string, characterRefs []string) ([]byte, error) {
				return []byte("testdata"), nil
			},
		}

		panels := []domain.Panel{
			{SceneNumber: 1, PanelNumber: 1, Description: "1"},
			{SceneNumber: 1, PanelNumber: 2, Description: "2"},
		}

		out, errs := image.GenerateBatch(context.Background(), client, panels, tmpDir, 2)
		assert.Len(t, errs, 0)
		assert.Len(t, out, 2)

		assert.Equal(t, 2, client.CallCount)
		assert.Contains(t, out[0].ImageURL, "scene_1_panel_1.png")
		assert.Contains(t, out[1].ImageURL, "scene_1_panel_2.png")

		// Verify disk writes
		d1, _ := os.ReadFile(out[0].ImageURL)
		assert.Equal(t, []byte("testdata"), d1)
	})

	t.Run("partial failure", func(t *testing.T) {
		client := &image.MockClient{
			GenerateImageFunc: func(ctx context.Context, prompt string, characterRefs []string) ([]byte, error) {
				if prompt == "fail" {
					return nil, errors.New("api limit")
				}
				return []byte("ok"), nil
			},
		}

		panels := []domain.Panel{
			{SceneNumber: 2, PanelNumber: 1, Description: "fail"},
			{SceneNumber: 2, PanelNumber: 2, Description: "ok"},
		}

		out, errs := image.GenerateBatch(context.Background(), client, panels, tmpDir, 2)
		assert.Len(t, errs, 1)
		assert.ErrorContains(t, errs[0], "api limit")

		var url1, url2 string
		for _, p := range out {
			if p.Description == "fail" {
				url1 = p.ImageURL
			} else {
				url2 = p.ImageURL
			}
		}

		assert.Equal(t, "error.png", url1)
		assert.Contains(t, url2, "scene_2_panel_2.png")
	})

	t.Run("invalid output dir", func(t *testing.T) {
		client := &image.MockClient{}
		panels := []domain.Panel{{PanelNumber: 1}}

		// Unwritable path to trigger os.MkdirAll error
		os.MkdirAll(filepath.Join(tmpDir, "restricted"), 0500) // read only

		// If os.MkdirAll doesn't fail based on permissions (like if running as root in tests),
		// we can pass a file as a directory to force failure.
		fileAsDir := filepath.Join(tmpDir, "file_as_dir")
		os.WriteFile(fileAsDir, []byte("dummy"), 0644)

		_, errs := image.GenerateBatch(context.Background(), client, panels, fileAsDir, 1)
		assert.Len(t, errs, 1)
		assert.ErrorContains(t, errs[0], "creating output dir")
	})
}
