package audio

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

// MusicClient defines the interface for fetching background music.
type MusicClient interface {
	SearchAndDownload(ctx context.Context, tags string) ([]byte, error)
}

// JamendoClient implements MusicClient using the Jamendo v3 API.
type JamendoClient struct {
	clientID string
}

func NewJamendoClient(clientID string) *JamendoClient {
	if clientID == "" {
		// Public test key commonly used in docs, though limits apply.
		clientID = "56d30c95"
	}
	return &JamendoClient{clientID: clientID}
}

// SearchAndDownload searches Jamendo by tags, picks the first match, and downloads its MP3 audio.
func (c *JamendoClient) SearchAndDownload(ctx context.Context, tags string) ([]byte, error) {
	// 1. Search for a track
	apiURL := fmt.Sprintf("https://api.jamendo.com/v3.0/tracks/?client_id=%s&format=json&limit=1&tags=%s",
		c.clientID, url.QueryEscape(tags))

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, apiURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create jamendo request: %w", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("jamendo API request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("jamendo API returned status %d", resp.StatusCode)
	}

	var result struct {
		Headers struct {
			Status       string `json:"status"`
			ErrorCode    int    `json:"code"`
			ErrorMessage string `json:"error_message"`
		} `json:"headers"`
		Results []struct {
			ID    string `json:"id"`
			Name  string `json:"name"`
			Audio string `json:"audio"` // URL to the mp3
		} `json:"results"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode jamendo response: %w", err)
	}

	if result.Headers.Status != "success" {
		return nil, fmt.Errorf("jamendo API error: %s", result.Headers.ErrorMessage)
	}
	if len(result.Results) == 0 {
		return nil, fmt.Errorf("no tracks found for tags: %s", tags)
	}

	track := result.Results[0]
	if track.Audio == "" {
		return nil, fmt.Errorf("jamendo track %s has no audio url", track.ID)
	}

	// 2. Download the track audio
	return c.downloadAudio(ctx, track.Audio)
}

func (c *JamendoClient) downloadAudio(ctx context.Context, audioURL string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, audioURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create download request: %w", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("audio download failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("audio download returned status %d", resp.StatusCode)
	}

	return io.ReadAll(resp.Body)
}
