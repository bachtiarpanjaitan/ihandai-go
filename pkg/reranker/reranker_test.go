package reranker

import (
	"context"
	"errors"
	"testing"

	"github.com/bachtiarpanjaitan/ihandai-go"
)

// Compile-time interface satisfaction check
var _ Reranker = (*mockReranker)(nil)
var _ Reranker = (*failingReranker)(nil)

type mockReranker struct{}

func (m mockReranker) Rerank(ctx context.Context, query string, docs []ihandai.Document) ([]ihandai.ScoredDocument, error) {
	result := make([]ihandai.ScoredDocument, len(docs))
	for i, doc := range docs {
		result[i] = ihandai.ScoredDocument{
			Document: doc,
			Score:    1.0 - float64(i)*0.1, // decreasing scores
		}
	}
	return result, nil
}

type failingReranker struct{}

func (f failingReranker) Rerank(ctx context.Context, query string, docs []ihandai.Document) ([]ihandai.ScoredDocument, error) {
	return nil, errors.New("rerank failed")
}

func TestReranker_Mock(t *testing.T) {
	var r Reranker = mockReranker{}
	docs := []ihandai.Document{
		{ID: "1", Content: "first"},
		{ID: "2", Content: "second"},
	}
	results, err := r.Rerank(context.Background(), "query", docs)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 2 {
		t.Errorf("got %d results, want 2", len(results))
	}
	if results[0].Score <= results[1].Score {
		t.Error("first result should have higher score than second")
	}
}

func TestReranker_EmptyInput(t *testing.T) {
	var r Reranker = mockReranker{}
	results, err := r.Rerank(context.Background(), "query", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if results == nil {
		t.Error("expected empty slice, got nil")
	}
}

func TestReranker_Failure(t *testing.T) {
	var r Reranker = failingReranker{}
	_, err := r.Rerank(context.Background(), "query", nil)
	if err == nil {
		t.Error("expected error from failing reranker")
	}
}
