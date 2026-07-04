// Package splitter defines the interface for splitting documents into chunks.
//
// Splitting is a critical step in RAG pipelines: chunks that are too large
// lose precision in similarity search; chunks that are too small lose context.
package splitter

import (
	"context"

	"github.com/bachtiarpanjaitan/ihandai-go"
)

// TextSplitter splits documents into smaller chunks.
//
// Implementations may split by character count, token count, sentences,
// paragraphs, or use semantic boundaries.
type TextSplitter interface {
	// Split divides documents into chunks suitable for embedding and retrieval.
	// The context controls cancellation and carries tracing spans.
	Split(ctx context.Context, documents []ihandai.Document) ([]ihandai.Chunk, error)
}

// Config holds configuration for a text splitter.
type Config struct {
	// ChunkSize is the target size of each chunk (in characters or tokens,
	// depending on the implementation).
	ChunkSize int

	// Overlap is the number of characters/tokens that adjacent chunks share.
	// Overlap helps prevent splitting in the middle of a relevant passage.
	Overlap int
}

// Option is a functional option for configuring a text splitter.
type Option func(*Config)

// WithChunkSize sets the target chunk size.
func WithChunkSize(size int) Option {
	return func(c *Config) {
		c.ChunkSize = size
	}
}

// WithOverlap sets the overlap between adjacent chunks.
func WithOverlap(overlap int) Option {
	return func(c *Config) {
		c.Overlap = overlap
	}
}
