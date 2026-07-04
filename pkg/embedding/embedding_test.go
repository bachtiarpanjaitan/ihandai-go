package embedding

import (
	"context"
	"errors"
	"testing"
)

// Compile-time interface satisfaction check
var _ Embedder = (*mockEmbedder)(nil)

type mockEmbedder struct{}

func (m mockEmbedder) Embed(ctx context.Context, text string) ([]float64, error) {
	return []float64{0.1, 0.2, 0.3}, nil
}

func (m mockEmbedder) EmbedBatch(ctx context.Context, texts []string) ([][]float64, error) {
	result := make([][]float64, len(texts))
	for i := range texts {
		result[i] = []float64{0.1, 0.2, 0.3}
	}
	return result, nil
}

type failingEmbedder struct{}

func (f failingEmbedder) Embed(ctx context.Context, text string) ([]float64, error) {
	return nil, errors.New("embedding failed")
}

func (f failingEmbedder) EmbedBatch(ctx context.Context, texts []string) ([][]float64, error) {
	return nil, errors.New("embedding batch failed")
}

func TestEmbedder_Mock(t *testing.T) {
	var e Embedder = mockEmbedder{}
	vec, err := e.Embed(context.Background(), "hello")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(vec) != 3 {
		t.Errorf("got %d dimensions, want 3", len(vec))
	}
}

func TestEmbedder_EmbedBatch(t *testing.T) {
	var e Embedder = mockEmbedder{}
	texts := []string{"a", "b", "c"}
	vecs, err := e.EmbedBatch(context.Background(), texts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(vecs) != 3 {
		t.Errorf("got %d vectors, want 3", len(vecs))
	}
}

func TestEmbedder_ContextCancelled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err := (&failingEmbedder{}).Embed(ctx, "test")
	if err == nil {
		t.Error("expected error for canceled context")
	}
}

func TestEmbedder_Failure(t *testing.T) {
	var e Embedder = failingEmbedder{}
	_, err := e.Embed(context.Background(), "test")
	if err == nil {
		t.Error("expected error from failing embedder")
	}
}

func TestConfig_Options(t *testing.T) {
	cfg := Config{}
	WithModel("text-embedding-3-small")(&cfg)
	WithAPIKey("sk-123")(&cfg)
	WithBaseURL("https://api.example.com")(&cfg)

	if cfg.Model != "text-embedding-3-small" {
		t.Errorf("Model: got %q, want %q", cfg.Model, "text-embedding-3-small")
	}
	if cfg.APIKey != "sk-123" {
		t.Errorf("APIKey: got %q, want %q", cfg.APIKey, "sk-123")
	}
}
