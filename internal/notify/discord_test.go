package notify_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/baochen10luo/stagenthand/internal/notify"
)

func TestDiscordNotifier_Notify_Success(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("Expected POST method, got %s", r.Method)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer ts.Close()

	n := notify.NewDiscordNotifier(ts.URL)
	err := n.Notify(context.Background(), "Title", "Message", 0x00FF00)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
}

func TestDiscordNotifier_Notify_RateLimitRetry(t *testing.T) {
	var callCount int
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		if callCount == 1 {
			w.WriteHeader(http.StatusTooManyRequests)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer ts.Close()

	n := notify.NewDiscordNotifier(ts.URL)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second) // enough for one 1s retry
	defer cancel()

	err := n.Notify(ctx, "Title", "Message", 0)
	if err != nil {
		t.Fatalf("Expected no error after retry, got %v", err)
	}
	if callCount != 2 {
		t.Errorf("Expected 2 calls, got %d", callCount)
	}
}

func TestDiscordNotifier_Notify_EmptyURL(t *testing.T) {
	n := notify.NewDiscordNotifier("")
	err := n.Notify(context.Background(), "T", "M", 0)
	if err != nil {
		t.Fatalf("Expected no error for empty URL, got %v", err)
	}
}

func TestDiscordNotifier_Notify_Failure(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer ts.Close()

	n := notify.NewDiscordNotifier(ts.URL)
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
	defer cancel()

	err := n.Notify(ctx, "T", "M", 0)
	if err == nil {
		t.Fatalf("Expected error due to context timeout or 500, got nil")
	}
}

func TestDiscordNotifier_Notify_InvalidURL(t *testing.T) {
	n := notify.NewDiscordNotifier("http://invalid-url-that-does-not-exist:0")
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
	defer cancel()
	err := n.Notify(ctx, "T", "M", 0)
	if err == nil {
		t.Fatalf("Expected error due to dial error, got nil")
	}
}
