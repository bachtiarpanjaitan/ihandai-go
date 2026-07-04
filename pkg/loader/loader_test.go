package loader

import (
	"context"
	"errors"
	"testing"

	"github.com/bachtiarpanjaitan/ihandai-go/pkg/core"
)

// Compile-time interface satisfaction check
var _ DocumentLoader = (*mockLoader)(nil)
var _ DocumentLoader = (*failingLoader)(nil)

type mockLoader struct{}

func (m mockLoader) Load(ctx context.Context, source string) ([]core.Document, error) {
	return []core.Document{
		{ID: "1", Content: "loaded from " + source},
	}, nil
}

type failingLoader struct{}

func (f failingLoader) Load(ctx context.Context, source string) ([]core.Document, error) {
	return nil, errors.New("load failed")
}

func TestDocumentLoader_Mock(t *testing.T) {
	var l DocumentLoader = mockLoader{}
	docs, err := l.Load(context.Background(), "test.txt")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(docs) != 1 {
		t.Errorf("got %d docs, want 1", len(docs))
	}
	if docs[0].Content != "loaded from test.txt" {
		t.Errorf("got %q, want %q", docs[0].Content, "loaded from test.txt")
	}
}

func TestDocumentLoader_Failure(t *testing.T) {
	var l DocumentLoader = failingLoader{}
	_, err := l.Load(context.Background(), "test.txt")
	if err == nil {
		t.Error("expected error from failing loader")
	}
}

func TestDocumentLoader_ContextCancelled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err := (&failingLoader{}).Load(ctx, "test.txt")
	if err == nil {
		t.Error("expected error from cancelled context")
	}
}
