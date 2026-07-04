// Package embedding defines interfaces for text embedding providers.
//
// Embedding models convert text into fixed-length floating-point vectors
// that capture semantic meaning. These vectors are used for similarity
// search, clustering, and as input to other ML models.
package embedding

import "context"

// Embedder converts text into vector representations.
type Embedder interface {
	// Embed converts a single text into a vector.
	// The context controls cancellation and carries tracing spans.
	Embed(ctx context.Context, text string) ([]float64, error)

	// EmbedBatch converts multiple texts into vectors in a single call.
	// Providers should batch API calls internally for efficiency.
	// The returned slice has the same length and order as the input.
	EmbedBatch(ctx context.Context, texts []string) ([][]float64, error)
}

// Config holds configuration for an embedding provider.
type Config struct {
	// Model is the embedding model name (e.g., "text-embedding-3-small").
	Model string

	// APIKey is the authentication key for the provider.
	APIKey string

	// BaseURL is the optional custom endpoint URL.
	BaseURL string
}

// Option is a functional option for configuring an embedding provider.
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
