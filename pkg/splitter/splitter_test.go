package splitter

import (
	"context"
	"testing"

	"github.com/bachtiarpanjaitan/ihandai-go/pkg/core"
)

// Compile-time interface satisfaction check
var _ TextSplitter = (*mockSplitter)(nil)

type mockSplitter struct{}

func (m mockSplitter) Split(ctx context.Context, docs []core.Document) ([]core.Chunk, error) {
	var chunks []core.Chunk
	for _, doc := range docs {
		chunks = append(chunks, core.Chunk{
			ID:       doc.ID + "-chunk-0",
			Content:  doc.Content,
			ParentID: doc.ID,
		})
	}
	return chunks, nil
}

func TestTextSplitter_Mock(t *testing.T) {
	var s TextSplitter = mockSplitter{}
	docs := []core.Document{
		{ID: "1", Content: "hello world"},
		{ID: "2", Content: "foo bar"},
	}
	chunks, err := s.Split(context.Background(), docs)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(chunks) != 2 {
		t.Errorf("got %d chunks, want 2", len(chunks))
	}
	if chunks[0].ParentID != "1" {
		t.Errorf("ParentID: got %q, want %q", chunks[0].ParentID, "1")
	}
}

func TestTextSplitter_EmptyInput(t *testing.T) {
	var s TextSplitter = mockSplitter{}
	chunks, err := s.Split(context.Background(), nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(chunks) != 0 {
		t.Errorf("expected empty/nil slice, got %d chunks", len(chunks))
	}
}

func TestConfig_Options(t *testing.T) {
	cfg := Config{}
	WithChunkSize(1000)(&cfg)
	WithOverlap(200)(&cfg)

	if cfg.ChunkSize != 1000 {
		t.Errorf("ChunkSize: got %d, want 1000", cfg.ChunkSize)
	}
	if cfg.Overlap != 200 {
		t.Errorf("Overlap: got %d, want 200", cfg.Overlap)
	}
}
