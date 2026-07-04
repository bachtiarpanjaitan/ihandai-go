// Package llm defines interfaces for Large Language Model providers.
//
// It provides small, composable interfaces following Go idioms:
// one interface, one responsibility. Consumers declare only the
// methods they need.
//
// One provider (e.g., OpenAI) typically implements all three interfaces
// (ChatCompleter, StreamCompleter, TokenCounter).
package llm

import (
	"context"

	"github.com/bachtiarpanjaitan/ihandai-go"
)

// ChatCompleter is the core LLM interface for synchronous chat completion.
// Implementations send messages to an LLM and return a single response.
type ChatCompleter interface {
	// Chat sends messages to the LLM and returns the generated response.
	// The context controls cancellation, timeouts, and carries tracing spans.
	Chat(ctx context.Context, messages []ihandai.Message) (*ihandai.Response, error)
}

// StreamCompleter provides streaming chat completion.
// Instead of waiting for the full response, the LLM streams tokens
// as they are generated through a channel.
type StreamCompleter interface {
	// ChatStream sends messages and returns a channel of response chunks.
	// The channel is closed when generation is complete.
	// The caller must read from the channel until it closes.
	ChatStream(ctx context.Context, messages []ihandai.Message) (<-chan Chunk, error)
}

// TokenCounter estimates token usage for a given text and model.
// Useful for context window management and cost estimation.
type TokenCounter interface {
	// CountTokens returns the estimated number of tokens in the text.
	CountTokens(ctx context.Context, model, text string) (int, error)
}

// Chunk represents a single token or group of tokens from a streaming response.
type Chunk struct {
	// Content is the text content of this chunk.
	Content string

	// FinishReason indicates why the stream ended, if this is the final chunk.
	// Empty string means more chunks are coming.
	FinishReason string
}

// Config holds configuration for an LLM provider.
type Config struct {
	// Model is the model name (e.g., "gpt-4o", "claude-sonnet-5").
	Model string

	// APIKey is the authentication key for the provider.
	APIKey string

	// BaseURL is the optional custom endpoint URL.
	BaseURL string
}

// Option is a functional option for configuring an LLM provider.
type Option func(*Config)

// WithModel sets the model name.
func WithModel(model string) Option {
	return func(c *Config) {
		c.Model = model
	}
}

// WithAPIKey sets the API key.
func WithAPIKey(key string) Option {
	return func(c *Config) {
		c.APIKey = key
	}
}

// WithBaseURL sets a custom base URL for the provider.
func WithBaseURL(url string) Option {
	return func(c *Config) {
		c.BaseURL = url
	}
}
