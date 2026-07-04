// Package retriever defines the interface for document retrieval strategies.
//
// Retrievers wrap a VectorSearcher and apply strategies like top-K, MMR,
// hybrid search, or multi-query expansion. They can be composed to build
// complex retrieval pipelines.
package retriever

import (
	"context"

	"github.com/bachtiarpanjaitan/ihandai-go/pkg/core"
)

// Retriever retrieves relevant documents for a query vector.
//
// The retriever is a strategy layer on top of VectorSearcher. It may
// transform the query, call the searcher multiple times, merge results,
// or apply re-ranking logic.
type Retriever interface {
	// Retrieve finds documents relevant to the given query vector.
	// The context controls cancellation and carries tracing spans.
	Retrieve(ctx context.Context, query []float64, opts ...RetrieveOption) ([]core.ScoredDocument, error)
}

// RetrieveOption is a functional option for retrieval operations.
type RetrieveOption func(*RetrieveConfig)

// RetrieveConfig holds parameters for a retrieval operation.
type RetrieveConfig struct {
	// TopK is the number of documents to return.
	TopK int

	// Filter is an optional metadata filter.
	Filter map[string]any

	// Strategy-specific configuration.
	Strategy map[string]any
}

// WithTopK sets the number of documents to return.
func WithTopK(k int) RetrieveOption {
	return func(c *RetrieveConfig) {
		c.TopK = k
	}
}

// WithFilter sets metadata filters for the retrieval.
func WithFilter(filter map[string]any) RetrieveOption {
	return func(c *RetrieveConfig) {
		c.Filter = filter
	}
}
