// Package loader defines the interface for loading documents from various sources.
//
// Implementations may load from files, URLs, databases, APIs, or any
// other document source.
package loader

import (
	"context"

	"github.com/bachtiarpanjaitan/ihandai-go"
)

// DocumentLoader loads documents from a source.
//
// The source is provider-specific: it could be a file path, URL, database
// connection string, or any other locator that the implementation understands.
type DocumentLoader interface {
	// Load fetches documents from the given source.
	// The context controls cancellation and carries tracing spans.
	Load(ctx context.Context, source string) ([]ihandai.Document, error)
}
