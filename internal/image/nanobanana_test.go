package image_test

import (
	"context"
	"encoding/base64"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/baochen10luo/stagenthand/internal/image"
	"github.com/stretchr/testify/assert"
)

func TestNanoBananaClient_GenerateImage(t *testing.T) {
	t.Parallel()

	// Create a dummy image payload.
	dummyImg := []byte("dummy-image-data")
	b64Img := base64.StdEncoding.EncodeToString(dummyImg)

	// Spin up a mock HTTP server.
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/images/generations", r.URL.Path)
		assert.Equal(t, "Bearer test-key", r.Header.Get("Authorization"))

		// Return OpenAI-compatible success payload
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{
			"data": [
				{ "b64_json": "` + b64Img + `" }
			]
		}`))
	}))
	defer server.Close()

	// Initialize the NanoBananaClient using the mock server URL.
	client := image.NewNanoBananaClient(server.URL, "test-key", "test-model")

	ctx := context.Background()
	res, err := client.GenerateImage(ctx, "A test prompt", []string{"/path/to/hero.png"})

	assert.NoError(t, err)
	assert.Equal(t, dummyImg, res)
}
