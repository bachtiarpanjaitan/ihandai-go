// Package reranker defines the interface for re-ranking retrieved documents.
//
// Re-ranking improves retrieval quality by using a more expensive model
// (cross-encoder or LLM) to score candidate documents against the query.
// This is applied after the initial vector search to narrow down results.
package reranker

import (
	"context"

	"github.com/bachtiarpanjaitan/ihandai-go/pkg/core"
)

// Reranker re-ranks a set of candidate documents against a query.
//
// The input is a list of documents (typically from vector search).
// The output is the same documents sorted by relevance, possibly
// filtered to remove low-relevance results.
type Reranker interface {
	// Rerank scores and re-orders documents by relevance to the query.
	// The context controls cancellation and carries tracing spans.
	Rerank(ctx context.Context, query string, documents []core.Document) ([]core.ScoredDocument, error)
}
