package retriever

import "math"

// CosineSimilarity returns the cosine similarity between two vectors.
// Range: [-1, 1], where 1 means identical direction.
func CosineSimilarity(a, b []float64) float64 {
	if len(a) != len(b) {
		return 0
	}
	var dot, normA, normB float64
	for i := range a {
		dot += a[i] * b[i]
		normA += a[i] * a[i]
		normB += b[i] * b[i]
	}
	if normA == 0 || normB == 0 {
		return 0
	}
	return dot / (math.Sqrt(normA) * math.Sqrt(normB))
}

// Embedder is a minimal interface for getting embeddings.
// Used by retrievers that need to embed query variants.
type Embedder interface {
	Embed(text string) ([]float64, error)
}
