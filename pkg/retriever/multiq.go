package retriever

import (
	"context"
	"sort"
	"strings"

	"github.com/bachtiarpanjaitan/ihandai-go/pkg/core"
	"github.com/bachtiarpanjaitan/ihandai-go/pkg/embedding"
	"github.com/bachtiarpanjaitan/ihandai-go/pkg/llm"
	"github.com/bachtiarpanjaitan/ihandai-go/pkg/vectordb"
)

// MultiQuery generates multiple query variants from the original query,
// retrieves documents for each variant, and merges deduplicated results.
//
// This improves recall for complex or ambiguous queries by searching
// from multiple angles.
type MultiQuery struct {
	searcher  vectordb.VectorSearcher
	llm       llm.ChatCompleter
	embedder  embedding.Embedder
	variants  int
}

// NewMultiQuery creates a new MultiQuery retriever.
// variants controls how many query variants to generate (default: 3).
func NewMultiQuery(searcher vectordb.VectorSearcher, llm llm.ChatCompleter, embedder embedding.Embedder, variants int) *MultiQuery {
	if variants < 2 {
		variants = 3
	}
	return &MultiQuery{
		searcher: searcher,
		llm:      llm,
		embedder: embedder,
		variants: variants,
	}
}

// Retrieve implements Retriever.
func (m *MultiQuery) Retrieve(ctx context.Context, query []float64, opts ...RetrieveOption) ([]core.ScoredDocument, error) {
	cfg := RetrieveConfig{TopK: 5}
	for _, opt := range opts {
		opt(&cfg)
	}

	// Step 1: Generate query variants
	variants, err := m.generateVariants(ctx, query, cfg)
	if err != nil {
		// Fall back to single query
		return retrieverSearch(ctx, m.searcher, query, cfg)
	}

	// Step 2: Retrieve for each variant in parallel
	type result struct {
		docs []core.ScoredDocument
		err  error
	}
	results := make(chan result, len(variants))

	for _, v := range variants {
		go func(v []float64) {
			docs, err := retrieverSearch(ctx, m.searcher, v, cfg)
			results <- result{docs, err}
		}(v)
	}

	// Step 3: Collect, deduplicate by ID, keep highest score
	seen := make(map[string]float64)
	var allDocs []core.ScoredDocument

	for range variants {
		r := <-results
		if r.err != nil {
			continue
		}
		for _, doc := range r.docs {
			if existing, ok := seen[doc.ID]; ok {
				if doc.Score > existing {
					seen[doc.ID] = doc.Score
				}
			} else {
				seen[doc.ID] = doc.Score
				allDocs = append(allDocs, doc)
			}
		}
	}

	// Update scores from dedup map
	for i := range allDocs {
		if s, ok := seen[allDocs[i].ID]; ok {
			allDocs[i].Score = s
		}
	}

	// Sort by score descending
	sort.Slice(allDocs, func(i, j int) bool {
		return allDocs[i].Score > allDocs[j].Score
	})

	if len(allDocs) > cfg.TopK {
		allDocs = allDocs[:cfg.TopK]
	}

	return allDocs, nil
}

// generateVariants uses the LLM to create query variants.
func (m *MultiQuery) generateVariants(ctx context.Context, baseEmbedding []float64, cfg RetrieveConfig) (vectors [][]float64, err error) {
	// Generate variants via LLM
	prompt := []core.Message{
		{
			Role:    "system",
			Content: "You are a query expansion assistant. Given a user query, generate alternative phrasings that capture different aspects or angles of the same information need. Output each variant on a new line, no numbering or bullets.",
		},
		{
			Role:    "user",
			Content: "Generate " + itos(m.variants-1) + " alternative ways to ask this question. Output only the variants, one per line.",
		},
	}

	// We embed the original query and use the LLM to generate text variants
	variants := make([][]float64, 0, m.variants)
	variants = append(variants, baseEmbedding) // include original

	// Generate text variants via LLM
	resp, err := m.llm.Chat(ctx, prompt)
	if err != nil {
		return variants, nil // return original only
	}

	// Parse variants and embed them
	lines := strings.Split(strings.TrimSpace(resp.Content), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || len(line) < 3 {
			continue
		}
		vec, err := m.embedder.Embed(ctx, line)
		if err != nil {
			continue
		}
		variants = append(variants, vec)
		if len(variants) >= m.variants {
			break
		}
	}

	return variants, nil
}

func itos(n int) string {
	switch n {
	case 1: return "1"
	case 2: return "2"
	case 3: return "3"
	case 4: return "4"
	case 5: return "5"
	default: return "2"
	}
}

func retrieverSearch(ctx context.Context, searcher vectordb.VectorSearcher, query []float64, cfg RetrieveConfig) ([]core.ScoredDocument, error) {
	searchOpts := []vectordb.SearchOption{
		vectordb.WithTopK(cfg.TopK),
	}
	if cfg.Filter != nil {
		searchOpts = append(searchOpts, vectordb.WithFilter(cfg.Filter))
	}
	return searcher.Search(ctx, query, searchOpts...)
}
