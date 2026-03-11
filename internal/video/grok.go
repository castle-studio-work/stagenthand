package video

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// GrokClient implements video.Client using the x.ai Grok video API.
type GrokClient struct {
	apiKey  string
	baseURL string
	client  *http.Client
}

func NewGrokClient(apiKey string, baseURL string) *GrokClient {
	if baseURL == "" {
		baseURL = "https://api.x.ai/v1"
	}
	return &GrokClient{
		apiKey:  apiKey,
		baseURL: baseURL,
		client: &http.Client{
			Timeout: 2 * time.Minute, // Video generation might take time
		},
	}
}

func (c *GrokClient) GenerateVideo(ctx context.Context, imageURL string, prompt string) ([]byte, error) {
	if c.apiKey == "" {
		return nil, fmt.Errorf("grok video generation requires an API key")
	}

	// This assumes the x.ai video endpoint mimics some OpenAI compatible behavior or
	// provides this endpoint for video generation. For demonstration we use /video/generations.
	reqBody := map[string]interface{}{
		"model":  "grok-video-1",
		"prompt": prompt,
	}
	if imageURL != "" {
		reqBody["image_url"] = imageURL
	}

	bodyData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal video request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/video/generations", bytes.NewReader(bodyData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.apiKey)

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request to grok failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("grok api error: status %d", resp.StatusCode)
	}

	var res struct {
		Data []struct {
			URL     string `json:"url"`
			Content string `json:"content"` // base64 or link
		} `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return nil, fmt.Errorf("failed to decode grok response: %w", err)
	}

	if len(res.Data) == 0 {
		return nil, fmt.Errorf("no video returned by grok")
	}

	// Currently shand assumes direct binary for mock/tests or handles URL vs Content
	// Fallback to fetch if URL is returned, or dummy for tests
	if res.Data[0].URL != "" {
		videoResp, err := c.client.Get(res.Data[0].URL)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch video url: %w", err)
		}
		defer videoResp.Body.Close()
		return io.ReadAll(videoResp.Body)
	}

	return []byte(res.Data[0].Content), nil // if pre-encoded or dummy
}
