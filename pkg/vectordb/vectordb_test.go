package vectordb

import (
	"context"
	"errors"
	"testing"

	"github.com/bachtiarpanjaitan/ihandai-go/pkg/core"
)

// Compile-time interface satisfaction checks
var _ VectorSearcher = (*mockStore)(nil)
var _ VectorInserter = (*mockStore)(nil)
var _ VectorDeleter = (*mockStore)(nil)

type mockStore struct {
	docs map[string]core.Document
}

func newMockStore() *mockStore {
	return &mockStore{docs: make(map[string]core.Document)}
}

func (m *mockStore) Search(ctx context.Context, vector []float64, opts ...SearchOption) ([]core.ScoredDocument, error) {
	result := make([]core.ScoredDocument, 0, len(m.docs))
	for _, doc := range m.docs {
		result = append(result, core.ScoredDocument{Document: doc, Score: 0.95})
	}
	return result, nil
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

type failingSearcher struct{}

func (f failingSearcher) Search(ctx context.Context, vector []float64, opts ...SearchOption) ([]core.ScoredDocument, error) {
	return nil, errors.New("search failed")
}

func TestVectorSearcher_Mock(t *testing.T) {
	store := newMockStore()
	store.Insert(context.Background(), []core.Document{
		{ID: "1", Content: "doc 1"},
		{ID: "2", Content: "doc 2"},
	})

	results, err := store.Search(context.Background(), []float64{0.1, 0.2})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 2 {
		t.Errorf("got %d results, want 2", len(results))
	}
}

func TestVectorInserter_Mock(t *testing.T) {
	store := newMockStore()
	err := store.Insert(context.Background(), []core.Document{
		{ID: "1", Content: "doc"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(store.docs) != 1 {
		t.Errorf("got %d docs in store, want 1", len(store.docs))
	}
}

func TestVectorDeleter_Mock(t *testing.T) {
	store := newMockStore()
	store.Insert(context.Background(), []core.Document{
		{ID: "1", Content: "doc 1"},
		{ID: "2", Content: "doc 2"},
	})
	err := store.Delete(context.Background(), []string{"1"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(store.docs) != 1 {
		t.Errorf("got %d docs after delete, want 1", len(store.docs))
	}
}

func TestVectorSearcher_Failure(t *testing.T) {
	var s VectorSearcher = failingSearcher{}
	_, err := s.Search(context.Background(), nil)
	if err == nil {
		t.Error("expected error from failing searcher")
	}
}

func TestContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	s := failingSearcher{}
	_, err := s.Search(ctx, nil)
	if err == nil {
		t.Error("expected error from cancelled context")
	}
}

func TestSearchOptions(t *testing.T) {
	cfg := SearchConfig{}
	WithTopK(10)(&cfg)
	WithFilter(map[string]any{"source": "docs/"})(&cfg)
	WithScoreThreshold(0.8)(&cfg)

	if cfg.TopK != 10 {
		t.Errorf("TopK: got %d, want 10", cfg.TopK)
	}
	if cfg.Filter["source"] != "docs/" {
		t.Errorf("Filter[source]: got %v, want docs/", cfg.Filter["source"])
	}
	if cfg.ScoreThreshold != 0.8 {
		t.Errorf("ScoreThreshold: got %f, want 0.8", cfg.ScoreThreshold)
	}
}

func TestConfigOptions(t *testing.T) {
	cfg := Config{}
	WithURL("http://localhost:6333")(&cfg)
	WithCollection("test-collection")(&cfg)
	WithAPIKey("secret")(&cfg)

	if cfg.URL != "http://localhost:6333" {
		t.Errorf("URL: got %q, want %q", cfg.URL, "http://localhost:6333")
	}
	if cfg.Collection != "test-collection" {
		t.Errorf("Collection: got %q, want %q", cfg.Collection, "test-collection")
	}
	if cfg.APIKey != "secret" {
		t.Errorf("APIKey: got %q, want %q", cfg.APIKey, "secret")
	}
}
