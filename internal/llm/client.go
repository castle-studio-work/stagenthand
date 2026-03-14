package llm

import "context"

// Client is the interface that wraps basic LLM generation methods.
// It adheres to the Dependency Inversion Principle, decoupling the
// CLI commands from specific provider implementations (like Gemini, OpenAI).
type Client interface {
	// GenerateTransformation sends a system prompt and input data to the LLM,
	// returning the raw generated output (typically JSON bytes).
	// This ensures that the caller does not know anything about
	// model-specific parameters or HTTP transport behavior.
	GenerateTransformation(ctx context.Context, systemPrompt string, inputData []byte) ([]byte, error)
}

// VideoCriticClient is the interface for multi-modal models that can watch and review videos.
type VideoCriticClient interface {
	ReviewVideo(ctx context.Context, systemPrompt string, inputData []byte, videoFormat string, videoBytes []byte) ([]byte, error)
}
