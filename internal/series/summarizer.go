package series

import (
	"context"
	"encoding/json"
	"fmt"
)

// LLMClient is the minimal interface needed to call an LLM for summarization.
// This is intentionally a local interface (DIP) rather than importing llm.Client directly.
type LLMClient interface {
	GenerateTransformation(ctx context.Context, systemPrompt string, inputData []byte) ([]byte, error)
}

// Summarizer extracts an EpisodeMemory from completed episode data.
type Summarizer interface {
	Summarize(ctx context.Context, episodeNum int, storyboardJSON []byte) (EpisodeMemory, error)
	CompressGlobal(ctx context.Context, m *SeriesMemory) (string, error)
}

// LLMSummarizer implements Summarizer using an LLM.
type LLMSummarizer struct {
	client LLMClient
}

// NewLLMSummarizer constructs an LLMSummarizer backed by the given LLM client.
func NewLLMSummarizer(client LLMClient) *LLMSummarizer {
	return &LLMSummarizer{client: client}
}

// Summarize extracts episode memory from a completed storyboard JSON using the LLM.
func (s *LLMSummarizer) Summarize(ctx context.Context, episodeNum int, storyboardJSON []byte) (EpisodeMemory, error) {
	prompt := fmt.Sprintf(PromptExtractEpisodeMemory, episodeNum)
	output, err := s.client.GenerateTransformation(ctx, prompt, storyboardJSON)
	if err != nil {
		return EpisodeMemory{}, fmt.Errorf("summarize LLM call failed: %w", err)
	}

	var mem EpisodeMemory
	if err := json.Unmarshal(output, &mem); err != nil {
		return EpisodeMemory{}, fmt.Errorf("parsing episode memory JSON: %w", err)
	}
	mem.Episode = episodeNum // ensure episode number is set correctly
	return mem, nil
}

// CompressGlobal generates a compressed global summary from all episode memories.
func (s *LLMSummarizer) CompressGlobal(ctx context.Context, m *SeriesMemory) (string, error) {
	input, err := json.Marshal(m)
	if err != nil {
		return "", fmt.Errorf("marshalling series memory: %w", err)
	}

	output, err := s.client.GenerateTransformation(ctx, PromptCompressGlobalSummary, input)
	if err != nil {
		return "", fmt.Errorf("compress global LLM call failed: %w", err)
	}

	return string(output), nil
}

// PromptExtractEpisodeMemory is the system prompt for extracting episode memory.
// It uses %d as a placeholder for the episode number.
const PromptExtractEpisodeMemory = `You are a story analyst. Given a storyboard JSON, extract:
- key_events: 3-5 bullet points of what happened
- characters: each character's name, description, motivation, and end-of-episode state
- world_facts: persistent world-building facts introduced

Respond with JSON only:
{"episode": %d, "key_events": [...], "characters": [...], "world_facts": [...]}`

// PromptCompressGlobalSummary is the system prompt for compressing all episode memories.
const PromptCompressGlobalSummary = `You are a story analyst. Given a series memory JSON with all episodes,
write a single paragraph (max 150 words) that captures the core story arc, main characters, and world so far.
Respond with the paragraph text only, no JSON.`
