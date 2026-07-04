package retriever

import (
	"context"

	"github.com/bachtiarpanjaitan/ihandai-go/pkg/core"
	"github.com/bachtiarpanjaitan/ihandai-go/pkg/vectordb"
)

// TopK is a simple retriever that delegates to VectorSearcher with top-K strategy.
type TopK struct {
	searcher vectordb.VectorSearcher
}

// NewTopK creates a new TopK retriever wrapping the given searcher.
func NewTopK(searcher vectordb.VectorSearcher) *TopK {
	return &TopK{searcher: searcher}
}

// Retrieve implements Retriever. It passes the query vector through to the
// underlying VectorSearcher with the configured options.
func (t *TopK) Retrieve(ctx context.Context, query []float64, opts ...RetrieveOption) ([]core.ScoredDocument, error) {
	cfg := RetrieveConfig{TopK: 5}
	for _, opt := range opts {
		opt(&cfg)
	}

	searchOpts := []vectordb.SearchOption{
		vectordb.WithTopK(cfg.TopK),
	}
	if cfg.Filter != nil {
		searchOpts = append(searchOpts, vectordb.WithFilter(cfg.Filter))
	}

	return t.searcher.Search(ctx, query, searchOpts...)
}
