package retriever

import (
	"context"
	"testing"

	"github.com/bachtiarpanjaitan/ihandai-go/pkg/core"
	"github.com/bachtiarpanjaitan/ihandai-go/pkg/vectordb"
)

// Compile-time checks
var _ Retriever = (*TopK)(nil)
var _ Retriever = (*MMR)(nil)
var _ Retriever = (*MultiQuery)(nil)

type mockRetriever struct{}

func (m mockRetriever) Retrieve(ctx context.Context, query []float64, opts ...RetrieveOption) ([]core.ScoredDocument, error) {
	return []core.ScoredDocument{
		{Document: core.Document{ID: "1", Content: "result"}, Score: 0.99},
	}, nil
}

type mockStore struct {
	docs map[string]core.Document
}

func newMockStore() *mockStore {
	return &mockStore{docs: map[string]core.Document{
		"1": {ID: "1", Content: "Machine learning basics"},
		"2": {ID: "2", Content: "Machine learning advanced"},
		"3": {ID: "3", Content: "Deep learning fundamentals"},
		"4": {ID: "4", Content: "Neural networks"},
		"5": {ID: "5", Content: "Python guide"},
		"6": {ID: "6", Content: "Go language"},
	}}
}

func (m *mockStore) Search(ctx context.Context, vector []float64, opts ...vectordb.SearchOption) ([]core.ScoredDocument, error) {
	cfg := vectordb.SearchConfig{TopK: 5}
	for _, opt := range opts {
		opt(&cfg)
	}
	var results []core.ScoredDocument
	for _, doc := range m.docs {
		results = append(results, core.ScoredDocument{Document: doc, Score: 0.95})
	}
	if len(results) > cfg.TopK {
		results = results[:cfg.TopK]
	}
	return results, nil
}

func (m *mockStore) Insert(ctx context.Context, documents []core.Document) error {
	for _, doc := range documents {
		m.docs[doc.ID] = doc
	}
	return nil
}

func (m *mockStore) Delete(ctx context.Context, ids []string) error {
	for _, id := range ids {
		delete(m.docs, id)
	}
	return nil
}

func TestRetriever_Mock(t *testing.T) {
	var r Retriever = mockRetriever{}
	docs, err := r.Retrieve(context.Background(), []float64{0.1})
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

func TestTopK_Retrieve(t *testing.T) {
	store := newMockStore()
	r := NewTopK(store)

	docs, err := r.Retrieve(context.Background(), []float64{0.1}, WithTopK(3))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(docs) != 3 {
		t.Errorf("got %d docs, want 3", len(docs))
	}
}

func TestMMR_Retrieve(t *testing.T) {
	store := newMockStore()
	r := NewMMR(store, 0.7)

	docs, err := r.Retrieve(context.Background(), []float64{0.1}, WithTopK(3))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(docs) != 3 {
		t.Errorf("got %d docs, want 3", len(docs))
	}

	docs2, err := r.Retrieve(context.Background(), []float64{0.1}, WithTopK(1))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(docs2) != 1 {
		t.Errorf("got %d docs, want 1", len(docs2))
	}
}

func TestMMR_FewerResults(t *testing.T) {
	store := &mockStore{docs: map[string]core.Document{
		"1": {ID: "1", Content: "only doc"},
	}}
	r := NewMMR(store, 0.7)

	docs, err := r.Retrieve(context.Background(), []float64{0.1}, WithTopK(5))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(docs) != 1 {
		t.Errorf("got %d docs, want 1", len(docs))
	}
}

func TestCosineSimilarity(t *testing.T) {
	a := []float64{1, 0, 0}
	b := []float64{0, 1, 0}
	c := []float64{1, 0, 0}

	if CosineSimilarity(a, b) != 0 {
		t.Error("orthogonal vectors should have similarity 0")
	}
	if CosineSimilarity(a, c) != 1 {
		t.Errorf("identical vectors should have similarity 1, got %f", CosineSimilarity(a, c))
	}
	if CosineSimilarity(a, []float64{1}) != 0 {
		t.Error("different lengths should return 0")
	}
}
