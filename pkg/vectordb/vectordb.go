// Package vectordb defines interfaces for vector database providers.
//
// Vector databases store and search high-dimensional vectors with metadata.
// They are the storage and retrieval backbone of RAG (Retrieval-Augmented Generation).
//
// A single provider (e.g., Qdrant) typically implements all three interfaces.
// Read-only replicas may only implement VectorSearcher.
package vectordb

import (
	"context"

	"github.com/bachtiarpanjaitan/ihandai-go"
)

// VectorSearcher performs similarity search over stored vectors.
type VectorSearcher interface {
	// Search finds documents whose vectors are most similar to the query vector.
	// Options control the number of results, filters, and search strategy.
	Search(ctx context.Context, vector []float64, opts ...SearchOption) ([]ihandai.ScoredDocument, error)
}

// VectorInserter stores documents with their vectors for later retrieval.
type VectorInserter interface {
	// Insert stores documents and their vectors.
	// Documents must have embeddings pre-computed (stored in metadata or a
	// provider-specific field).
	Insert(ctx context.Context, documents []ihandai.Document) error
}

// VectorDeleter removes documents from the vector store.
type VectorDeleter interface {
	// Delete removes documents by their IDs.
	// Deleting a non-existent ID is not an error.
	Delete(ctx context.Context, ids []string) error
}

// Config holds configuration for a vector database provider.
type Config struct {
	// URL is the connection URL or host:port for the database.
	URL string

	// Collection is the name of the collection/table/index to use.
	Collection string

	// APIKey is the authentication key (for cloud providers like Pinecone).
	APIKey string
}

// SearchOption is a functional option for vector search queries.
type SearchOption func(*SearchConfig)

// SearchConfig holds parameters for a single search operation.
type SearchConfig struct {
	// TopK is the number of results to return.
	TopK int

	// Filter is a metadata filter applied before or after search.
	Filter map[string]any

	// ScoreThreshold is the minimum score for returned results.
	ScoreThreshold float64
}

// WithTopK sets the number of results to return.
func WithTopK(k int) SearchOption {
	return func(c *SearchConfig) {
		c.TopK = k
	}
}

// WithFilter sets metadata filters for the search.
func WithFilter(filter map[string]any) SearchOption {
	return func(c *SearchConfig) {
		c.Filter = filter
	}
}

// WithScoreThreshold sets the minimum score for returned results.
func WithScoreThreshold(threshold float64) SearchOption {
	return func(c *SearchConfig) {
		c.ScoreThreshold = threshold
	}
}

// Option is a functional option for configuring a vector database provider.
type Option func(*Config)

// WithURL sets the connection URL.
func WithURL(url string) Option {
	return func(c *Config) {
		c.URL = url
	}
}

// WithCollection sets the collection name.
func WithCollection(name string) Option {
	return func(c *Config) {
		c.Collection = name
	}
}

// WithAPIKey sets the API key.
func WithAPIKey(key string) Option {
	return func(c *Config) {
		c.APIKey = key
	}
}
