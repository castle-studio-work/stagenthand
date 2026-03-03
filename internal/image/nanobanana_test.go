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

	dummyImg := []byte("dummy-image-data")
	b64Img := base64.StdEncoding.EncodeToString(dummyImg)

	t.Run("success", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "/images/generations", r.URL.Path)
			assert.Equal(t, "Bearer test-key", r.Header.Get("Authorization"))

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{
				"data": [
					{ "b64_json": "` + b64Img + `" }
				]
			}`))
		}))
		defer server.Close()

		client := image.NewNanoBananaClient(server.URL, "test-key", "test-model")
		res, err := client.GenerateImage(context.Background(), "A test prompt", []string{"/path"})
		assert.NoError(t, err)
		assert.Equal(t, dummyImg, res)
	})

	t.Run("api error 400", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(`{"error": {"message": "bad prompt"}}`))
		}))
		defer server.Close()

		client := image.NewNanoBananaClient(server.URL, "test-key", "test-model")
		_, err := client.GenerateImage(context.Background(), "A test prompt", nil)
		assert.ErrorContains(t, err, "bad prompt")
	})

	t.Run("empty response data", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"data": []}`))
		}))
		defer server.Close()

		client := image.NewNanoBananaClient(server.URL, "test-key", "test-model")
		_, err := client.GenerateImage(context.Background(), "A test prompt", nil)
		assert.ErrorContains(t, err, "empty data")
	})

	t.Run("invalid base64", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"data": [{ "b64_json": "invalid!@#" }]}`))
		}))
		defer server.Close()

		client := image.NewNanoBananaClient(server.URL, "test-key", "test-model")
		_, err := client.GenerateImage(context.Background(), "A test prompt", nil)
		assert.ErrorContains(t, err, "failed to decode")
	})

	t.Run("http failed", func(t *testing.T) {
		client := image.NewNanoBananaClient("http://127.0.0.1:0", "test-key", "test-model")
		_, err := client.GenerateImage(context.Background(), "A test prompt", nil)
		assert.ErrorContains(t, err, "http request failed")
	})
}
