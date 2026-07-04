package ollama

import (
	"context"
	"fmt"

	"github.com/bachtiarpanjaitan/ihandai-go/pkg/embedding"
)

// Compile-time interface satisfaction check.
var _ embedding.Embedder = (*Embedder)(nil)

// Embedder implements embedding.Embedder using the Ollama embeddings API.
type Embedder struct {
	cfg Config
}

// NewEmbedder creates a new Ollama Embedder with the given options.
func NewEmbedder(opts ...Option) *Embedder {
	return &Embedder{cfg: applyOptions(opts)}
}

// Embed converts a single text to a vector using Ollama.
func (e *Embedder) Embed(ctx context.Context, text string) ([]float64, error) {
	reqBody := embedRequest{
		Model:  e.cfg.Model,
		Prompt: text,
	}

	var result embedResponse
	url := e.cfg.BaseURL + "/api/embeddings"
	if err := doRequest(ctx, e.cfg.HTTPClient, url, reqBody, &result); err != nil {
		return nil, fmt.Errorf("ollama: embed: %w", err)
	}

	return result.Embedding, nil
}

// EmbedBatch converts multiple texts to vectors. Ollama's API does not
// support batch embeddings natively, so we issue one request per text.
func (e *Embedder) EmbedBatch(ctx context.Context, texts []string) ([][]float64, error) {
	results := make([][]float64, len(texts))
	for i, text := range texts {
		vec, err := e.Embed(ctx, text)
		if err != nil {
			return nil, fmt.Errorf("ollama: embed batch [%d]: %w", i, err)
		}
		results[i] = vec
	}
	return results, nil
}

// init registers the Ollama embedding provider.
func init() {
	embedding.Register("ollama", func(cfg embedding.Config) (embedding.Embedder, error) {
		opts := []Option{WithModel(cfg.Model)}
		if cfg.BaseURL != "" {
			opts = append(opts, WithBaseURL(cfg.BaseURL))
		}
		return NewEmbedder(opts...), nil
	})
}
