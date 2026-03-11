package video_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/baochen10luo/stagenthand/internal/video"
)

func TestGrokClient_GenerateVideo_SuccessWithContent(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"data": [{"content": "fake-video-bytes"}]}`))
	}))
	defer ts.Close()

	c := video.NewGrokClient("test-key", ts.URL)
	bytes, err := c.GenerateVideo(context.Background(), "http://image", "A test prompt")
	if err != nil {
		t.Fatalf("Expected nil, got %v", err)
	}

	if string(bytes) != "fake-video-bytes" {
		t.Errorf("Expected fake bytes, got %s", string(bytes))
	}
}

func TestGrokClient_GenerateVideo_SuccessWithURL(t *testing.T) {
	videoServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("downloaded-video-bytes"))
	}))
	defer videoServer.Close()

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"data": [{"url": "` + videoServer.URL + `"}]}`))
	}))
	defer ts.Close()

	c := video.NewGrokClient("test-key", ts.URL)
	bytes, err := c.GenerateVideo(context.Background(), "", "A test prompt")
	if err != nil {
		t.Fatalf("Expected nil, got %v", err)
	}

	if string(bytes) != "downloaded-video-bytes" {
		t.Errorf("Expected downloaded bytes, got %s", string(bytes))
	}
}

func TestGrokClient_Errors(t *testing.T) {
	c := video.NewGrokClient("", "")
	_, err := c.GenerateVideo(context.Background(), "", "")
	if err == nil {
		t.Errorf("Expected error for empty API key")
	}

	tsError := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer tsError.Close()

	c2 := video.NewGrokClient("key", tsError.URL)
	_, err2 := c2.GenerateVideo(context.Background(), "", "")
	if err2 == nil {
		t.Errorf("Expected error for 500 response")
	}

	tsEmpty := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"data": []}`))
	}))
	defer tsEmpty.Close()

	c3 := video.NewGrokClient("key", tsEmpty.URL)
	_, err3 := c3.GenerateVideo(context.Background(), "", "")
	if err3 == nil {
		t.Errorf("Expected error for empty data array")
	}

	tsInvalidJSON := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{invalid json}`))
	}))
	defer tsInvalidJSON.Close()

	c4 := video.NewGrokClient("key", tsInvalidJSON.URL)
	_, err4 := c4.GenerateVideo(context.Background(), "", "")
	if err4 == nil {
		t.Errorf("Expected error for invalid json")
	}

	c5 := video.NewGrokClient("key", "http://invalid-url-that-does-not-exist:0")
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
	defer cancel()
	_, err5 := c5.GenerateVideo(ctx, "", "")
	if err5 == nil {
		t.Errorf("Expected error for dial error")
	}
}
