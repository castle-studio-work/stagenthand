package image_test

import (
	"testing"

	"github.com/baochen10luo/stagenthand/config"
	"github.com/baochen10luo/stagenthand/internal/image"
	"github.com/baochen10luo/stagenthand/internal/render"
	"github.com/stretchr/testify/assert"
)

func TestNewClient(t *testing.T) {
	t.Parallel()

	cfg := &config.Config{
		Image: config.ImageConfig{
			APIKey: "test",
		},
	}

	t.Run("dry run", func(t *testing.T) {
		client, err := image.NewClient("nanobanana", true, cfg)
		assert.NoError(t, err)
		_, ok := client.(*image.MockClient)
		assert.True(t, ok)
	})

	t.Run("mock provider", func(t *testing.T) {
		client, err := image.NewClient("mock", false, cfg)
		assert.NoError(t, err)
		_, ok := client.(*image.MockClient)
		assert.True(t, ok)
	})

	t.Run("nanobanana provider", func(t *testing.T) {
		client, err := image.NewClient("nanobanana", false, cfg)
		assert.NoError(t, err)
		_, ok := client.(*image.NanoBananaClient)
		assert.True(t, ok)
	})

	t.Run("unknown provider", func(t *testing.T) {
		client, err := image.NewClient("unknown", false, cfg)
		assert.ErrorContains(t, err, "not implemented")
		assert.Nil(t, client)
	})
}

func TestNewClientWithFormat(t *testing.T) {
	t.Parallel()

	cfg := &config.Config{
		Image: config.ImageConfig{
			APIKey: "test",
		},
	}

	t.Run("portrait format returns nanobanana client", func(t *testing.T) {
		client, err := image.NewClientWithFormat("nanobanana", false, cfg, render.VideoFormatPortrait)
		assert.NoError(t, err)
		_, ok := client.(*image.NanoBananaClient)
		assert.True(t, ok)
	})

	t.Run("landscape format dry run returns mock", func(t *testing.T) {
		client, err := image.NewClientWithFormat("nanobanana", true, cfg, render.VideoFormatLandscape)
		assert.NoError(t, err)
		_, ok := client.(*image.MockClient)
		assert.True(t, ok)
	})
}
