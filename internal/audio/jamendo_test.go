package audio

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestJamendoClient_SearchAndDownload(t *testing.T) {
	// 1. Create a fake Jamendo API server
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check the URL path to distinguish between search and download
		if r.URL.Path == "/search/" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{
				"headers": {"status": "success", "code": 0},
				"results": [{"id": "123", "name": "Fake Track", "audio": "http://` + r.Host + `/download/123.mp3"}]
			}`))
		} else if r.URL.Path == "/download/123.mp3" {
			w.Header().Set("Content-Type", "audio/mpeg")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("fake-mp3-music"))
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer ts.Close()

	client := NewJamendoClient("test-key")
	client.baseURL = ts.URL + "/search/"

	ctx := context.Background()
	data, err := client.SearchAndDownload(ctx, "happy")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(data) != "fake-mp3-music" {
		t.Errorf("got %q, want %q", string(data), "fake-mp3-music")
	}
}

func TestJamendoClient_NoResults(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{
			"headers": {"status": "success", "code": 0},
			"results": []
		}`))
	}))
	defer ts.Close()

	client := NewJamendoClient("test-key")
	client.baseURL = ts.URL

	_, err := client.SearchAndDownload(context.Background(), "notfound")
	if err == nil {
		t.Error("expected error for no results, got nil")
	}
}

func TestJamendoClient_APIError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer ts.Close()

	client := NewJamendoClient("test-key")
	client.baseURL = ts.URL

	_, err := client.SearchAndDownload(context.Background(), "error")
	if err == nil {
		t.Error("expected error for non-200 status, got nil")
	}
}
