package image

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"time"

	"github.com/go-resty/resty/v2"
)

// NanoBananaClient implements the image.Client interface directly using HTTP,
// bypassing the fragility of sub-processes.
type NanoBananaClient struct {
	client *resty.Client
	apiKey string
	model  string
}

// NewNanoBananaClient initializes the HTTP client with exponential backoff retries.
func NewNanoBananaClient(baseURL, apiKey, model string) *NanoBananaClient {
	if baseURL == "" {
		// Based on Gemini memory rules, route everything through the Zeabur proxy by default
		baseURL = "https://pgb.zeabur.app/v1"
	}
	if model == "" {
		model = "nano-banana-2"
	}

	r := resty.New().
		SetBaseURL(baseURL).
		SetTimeout(120 * time.Second). // Image generation can take a while
		SetRetryCount(3).
		SetRetryWaitTime(2 * time.Second).
		SetRetryMaxWaitTime(10 * time.Second)

	return &NanoBananaClient{
		client: r,
		apiKey: apiKey,
		model:  model,
	}
}

// GenerateImage hits an OpenAI-compatible (or Zeabur-proxy-compatible) image generation endpoint.
// Requesting base64 back preserves the "file-less" pipeline flow in memory until we explicitly write it.
func (c *NanoBananaClient) GenerateImage(ctx context.Context, prompt string, characterRefs []string) ([]byte, error) {
	// This payload assumes the NanoBanana endpoint accepts OpenAI-like /images/generations format.
	// If it needs specific custom fields for Reference Images, they are passed down here.
	type ImageRequest struct {
		Model          string `json:"model"`
		Prompt         string `json:"prompt"`
		ResponseFormat string `json:"response_format"`
		// NanoBanana custom extension for consistent characters
		CharacterRefs []string `json:"character_refs,omitempty"`
	}

	type ImageResponse struct {
		Data []struct {
			B64Json string `json:"b64_json"`
		} `json:"data"`
		Error *struct {
			Message string `json:"message"`
		} `json:"error,omitempty"`
	}

	reqBody := ImageRequest{
		Model:          c.model,
		Prompt:         prompt,
		ResponseFormat: "b64_json",
		CharacterRefs:  characterRefs,
	}

	var resBody ImageResponse

	req := c.client.R().
		SetContext(ctx).
		SetHeader("Authorization", "Bearer "+c.apiKey).
		SetHeader("Content-Type", "application/json").
		SetBody(reqBody).
		SetResult(&resBody).
		SetError(&resBody)

	resp, err := req.Post("/images/generations")
	if err != nil {
		return nil, fmt.Errorf("http request failed: %w", err)
	}

	if resp.IsError() {
		errMsg := "unknown error"
		if resBody.Error != nil && resBody.Error.Message != "" {
			errMsg = resBody.Error.Message
		}
		return nil, fmt.Errorf("API error (status %d): %s", resp.StatusCode(), errMsg)
	}

	if len(resBody.Data) == 0 || resBody.Data[0].B64Json == "" {
		return nil, errors.New("API returned empty data or missing b64_json")
	}

	decoded, err := base64.StdEncoding.DecodeString(resBody.Data[0].B64Json)
	if err != nil {
		return nil, fmt.Errorf("failed to decode b64_json: %w", err)
	}

	return decoded, nil
}
