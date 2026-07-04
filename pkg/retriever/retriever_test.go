package retriever

import (
	"context"
	"testing"

	"github.com/bachtiarpanjaitan/ihandai-go"
)

// Compile-time interface satisfaction check
var _ Retriever = (*mockRetriever)(nil)

type mockRetriever struct{}

func (m mockRetriever) Retrieve(ctx context.Context, query []float64, opts ...RetrieveOption) ([]ihandai.ScoredDocument, error) {
	return []ihandai.ScoredDocument{
		{Document: ihandai.Document{ID: "1", Content: "result"}, Score: 0.99},
	}, nil
}

func TestRetriever_Mock(t *testing.T) {
	var r Retriever = mockRetriever{}
	docs, err := r.Retrieve(context.Background(), []float64{0.1, 0.2})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(docs) != 1 {
		t.Errorf("got %d docs, want 1", len(docs))
	}
	if docs[0].Score != 0.99 {
		t.Errorf("Score: got %f, want 0.99", docs[0].Score)
	}
}

func TestRetrieveOptions(t *testing.T) {
	cfg := RetrieveConfig{}
	WithTopK(5)(&cfg)
	WithFilter(map[string]any{"type": "article"})(&cfg)

	if cfg.TopK != 5 {
		t.Errorf("TopK: got %d, want 5", cfg.TopK)
	}
	if cfg.Filter["type"] != "article" {
		t.Errorf("Filter[type]: got %v, want article", cfg.Filter["type"])
	}
}
