package splitter

import (
	"context"
	"fmt"
	"strings"

	"github.com/bachtiarpanjaitan/ihandai-go/pkg/core"
)

// Recursive is a text splitter that recursively splits text by separators
// in order of priority: paragraphs → sentences → words → characters.
// It tries to keep chunks close to the target chunk size while respecting
// natural boundaries.
type Recursive struct {
	chunkSize int
	overlap   int
}

// NewRecursive creates a new Recursive splitter with the given options.
func NewRecursive(opts ...Option) *Recursive {
	cfg := Config{ChunkSize: 1000, Overlap: 200}
	for _, opt := range opts {
		opt(&cfg)
	}
	return &Recursive{chunkSize: cfg.ChunkSize, overlap: cfg.Overlap}
}

// Split implements TextSplitter.
func (r *Recursive) Split(ctx context.Context, documents []core.Document) ([]core.Chunk, error) {
	_ = ctx // reserved for future use (e.g., cancellation during long splits)
	var chunks []core.Chunk
	for _, doc := range documents {
		docChunks := r.splitText(doc.Content)
		for i, text := range docChunks {
			chunks = append(chunks, core.Chunk{
				ID:       fmt.Sprintf("%s-chunk-%d", doc.ID, i),
				Content:  text,
				ParentID: doc.ID,
				Metadata: doc.Metadata,
			})
		}
	}
	return chunks, nil
}

var separators = []string{"\n\n", "\n", ". ", "? ", "! ", " ", ""}

func (r *Recursive) splitText(text string) []string {
	return r.splitRecursive(text, 0)
}

func (r *Recursive) splitRecursive(text string, level int) []string {
	if level >= len(separators) {
		return []string{text}
	}

	sep := separators[level]
	var parts []string
	if sep == "" {
		// Character-level split
		parts = splitBySize(text, r.chunkSize, r.overlap)
	} else {
		parts = strings.Split(text, sep)
	}

	var result []string
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		if len(part) <= r.chunkSize {
			result = append(result, part)
		} else {
			// Too large, try next separator level
			result = append(result, r.splitRecursive(part, level+1)...)
		}
	}

	// Merge small chunks
	return r.mergeChunks(result)
}

func (r *Recursive) mergeChunks(chunks []string) []string {
	if len(chunks) <= 1 {
		return chunks
	}

	var merged []string
	current := chunks[0]
	for i := 1; i < len(chunks); i++ {
		combined := current + " " + chunks[i]
		if len(combined) <= r.chunkSize {
			current = combined
		} else {
			merged = append(merged, current)
			current = chunks[i]
		}
	}
	merged = append(merged, current)

	// Apply overlap: prepend end of previous chunk to current
	if r.overlap > 0 && len(merged) > 1 {
		for i := 1; i < len(merged); i++ {
			prev := merged[i-1]
			if len(prev) > r.overlap {
				overlapText := prev[len(prev)-r.overlap:]
				merged[i] = overlapText + merged[i]
			}
		}
	}

	return merged
}

func splitBySize(text string, size int, overlap int) []string {
	if len(text) <= size {
		return []string{text}
	}

	var chunks []string
	step := size - overlap
	if step <= 0 {
		step = size
	}

	for i := 0; i < len(text); i += step {
		end := i + size
		if end > len(text) {
			end = len(text)
		}
		chunks = append(chunks, text[i:end])
	}
	return chunks
}
