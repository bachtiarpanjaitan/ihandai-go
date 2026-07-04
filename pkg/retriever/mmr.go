package retriever

import (
	"context"
	"math"

	"github.com/bachtiarpanjaitan/ihandai-go/pkg/core"
	"github.com/bachtiarpanjaitan/ihandai-go/pkg/vectordb"
)

// MMR implements Maximal Marginal Relevance retrieval.
// It balances relevance (how similar a document is to the query) with
// diversity (how different it is from already-selected documents).
//
//	score = λ × relevance − (1−λ) × max_similarity_to_selected
//
// λ=1: pure relevance (no diversity penalty) → equivalent to TopK
// λ=0: pure diversity (no relevance consideration)
// λ=0.7: moderate diversity (recommended default)
type MMR struct {
	searcher vectordb.VectorSearcher
	lambda   float64
}

// NewMMR creates a new MMR retriever.
// lambda controls the relevance-diversity tradeoff.
func NewMMR(searcher vectordb.VectorSearcher, lambda float64) *MMR {
	return &MMR{searcher: searcher, lambda: lambda}
}

// Retrieve implements Retriever.
// It fetches extra candidates and then selects a diverse subset.
func (m *MMR) Retrieve(ctx context.Context, query []float64, opts ...RetrieveOption) ([]core.ScoredDocument, error) {
	cfg := RetrieveConfig{TopK: 5}
	for _, opt := range opts {
		opt(&cfg)
	}

	k := cfg.TopK
	fetchK := k * 3 // fetch more than needed for diversity selection
	if fetchK < 15 {
		fetchK = 15
	}

	candidates, err := m.searcher.Search(ctx, query, vectordb.WithTopK(fetchK))
	if err != nil {
		return nil, err
	}

	if len(candidates) <= k {
		return candidates, nil
	}

	selected := m.selectDiverse(query, candidates, k)
	return selected, nil
}

func (m *MMR) selectDiverse(query []float64, candidates []core.ScoredDocument, k int) []core.ScoredDocument {
	if k >= len(candidates) {
		return candidates
	}

	selected := make([]core.ScoredDocument, 0, k)
	available := make([]core.ScoredDocument, len(candidates))
	copy(available, candidates)

	// Embed each candidate's text for diversity comparison.
	// If we had embeddings stored, we'd use them directly.
	// For now, we use score as a proxy for relevance.

	// Select first: highest relevance score
	selected = append(selected, available[0])
	available = available[1:]

	for len(selected) < k && len(available) > 0 {
		bestIdx := 0
		bestScore := math.Inf(-1)

		for i, cand := range available {
			// Compute max similarity to already-selected
			maxSim := 0.0
			for _, sel := range selected {
				sim := scoreSimilarity(cand.Score, sel.Score)
				if sim > maxSim {
					maxSim = sim
				}
			}

			// MMR formula: λ * relevance − (1−λ) * max_similarity
			mmrScore := m.lambda*cand.Score - (1.0-m.lambda)*maxSim

			if mmrScore > bestScore {
				bestScore = mmrScore
				bestIdx = i
			}
		}

		selected = append(selected, available[bestIdx])
		available = append(available[:bestIdx], available[bestIdx+1:]...)
	}

	return selected
}

// scoreSimilarity computes how "similar" two scores are.
// We use score difference as a proxy since we don't have document embeddings.
func scoreSimilarity(a, b float64) float64 {
	diff := a - b
	if diff < 0 {
		diff = -diff
	}
	return math.Exp(-diff * 10) // exponential decay of similarity
}
