# RAG Engine

Retrieval-Augmented Generation (RAG) is the first major feature of ihandai.
It combines document retrieval with LLM generation to produce grounded, factual responses.

## Architecture

```
Index Phase                         Query Phase
───────────                         ──────────
DocumentLoader.Load(ctx, src)       Embedder.Embed(ctx, query)
  → []Document                        → []float64
TextSplitter.Split(ctx, docs)       VectorSearcher.Search(ctx, vector)
  → []Chunk                           → []ScoredDocument
Embedder.EmbedBatch(ctx, chunks)    Reranker.Rerank(ctx, query, docs)
  → [][]float64                       → []ScoredDocument
VectorInserter.Insert(ctx, docs)    PromptBuilder.Build(ctx, template, ctx)
  → nil                               → []Message
                                    ChatCompleter.Chat(ctx, messages)
                                      → *Response
```

Both pipelines run through `Client.Ask()` and `Client.Index()` respectively.
Every step accepts `ctx context.Context` for cancellation and tracing.

> **Graceful Degradation**: If `Ask()` is called without a configured embedding provider
> or vector store, it falls back to [`Chat()`](13-CLIENT#chat--simple-llm-without-rag) — a
> simple LLM call with no retrieval. This lets you start with minimal setup and add RAG
> later without changing your call sites.

## Milestones

### 1. Simple Similarity Search
**Status**: Phase 5.0 | **Depends on**: Phase 4 (Pipeline)

Basic cosine similarity search. Embed query → find K nearest vectors → return documents.

```go
resp, err := ai.Ask(ctx, "What is RAG?")
// Default: top-K=5, cosine similarity
```

### 2. Metadata Filtering
**Status**: Phase 5.1

Filter search results by metadata before or after vector search.

```go
resp, err := ai.Ask(ctx, "What is RAG?",
    ihandai.WithFilter(ihandai.Filter{
        "source": "docs/",
        "date":   ihandai.After(time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)),
    }),
)
```

### 3. MMR (Maximal Marginal Relevance)
**Status**: Phase 5.2

Balance relevance with diversity. Avoid returning near-duplicate documents.

```go
resp, err := ai.Ask(ctx, "What is RAG?",
    ihandai.WithRetrievalStrategy(ihandai.MMR),
    ihandai.WithTopK(10),
    ihandai.WithLambda(0.7),  // relevance vs diversity tradeoff
)
```

### 4. Hybrid Search
**Status**: Phase 5.3

Combine vector similarity with keyword-based search (BM25) for better recall.
Useful when exact term matching matters alongside semantic similarity.

```go
resp, err := ai.Ask(ctx, "What is RAG?",
    ihandai.WithRetrievalStrategy(ihandai.Hybrid),
    ihandai.WithHybridWeight(0.7),  // 0.7 vector + 0.3 keyword
)

// Requires VectorStore to support both vector and full-text search
```

### 5. Multi-Query Retrieval
**Status**: Phase 5.4

Generate multiple query variants from the original question, retrieve for each,
and merge deduplicated results. Improves recall for complex or ambiguous queries.

```go
resp, err := ai.Ask(ctx, "How do I optimize performance?",
    ihandai.WithRetrievalStrategy(ihandai.MultiQuery),
    ihandai.WithQueryVariants(3),  // generate 3 query variants
)

// Internally: ChatCompleter generates variants → parallel searches → dedup → merge
```

### 6. Parent Document Retrieval
**Status**: Phase 5.5

Index small chunks for precise matching, but retrieve the full parent document for
context. Essential when chunk boundaries cut through important context.

```
Index: Large doc → split into small chunks → embed each chunk
                                           → store chunk + parent reference
Query: Find relevant chunks → fetch parent documents → send parents to LLM
```

```go
resp, err := ai.Ask(ctx, "Explain the architecture",
    ihandai.WithRetrievalStrategy(ihandai.ParentDocument),
    ihandai.WithChunkSize(256),       // search with small chunks
    ihandai.WithParentChunkSize(1024), // retrieve larger context
)
```

### 7. Context Compression
**Status**: Phase 5.6

Use an LLM to compress retrieved documents before sending them as context.
Reduces token usage and removes irrelevant information from retrieved chunks.

```go
resp, err := ai.Ask(ctx, "Summarize the key findings",
    ihandai.WithRetrievalStrategy(ihandai.CompressContext),
    ihandai.WithCompressionModel("openai", llm.WithModel("gpt-3.5-turbo")),
)

// Internally: retrieve → LLM summarizes → compressed context → final LLM call
```

## Retrieval Strategies

All strategies implement the `Retriever` interface:

```go
type Retriever interface {
    Retrieve(ctx context.Context, query []float64, opts ...RetrieveOption) ([]ScoredDocument, error)
}
```

| Strategy | `RetrieveOption` | Description |
|----------|-----------------|-------------|
| `TopK` | `WithTopK(int)` | Simple K-nearest-neighbor (default) |
| `MMR` | `WithLambda(float64)` | Diversity-aware retrieval |
| `Hybrid` | `WithHybridWeight(float64)` | Vector + keyword combined |
| `MultiQuery` | `WithQueryVariants(int)` | Multi-variant query expansion |
| `ParentDocument` | `WithParentChunkSize(int)` | Small chunks → parent docs |
| `CompressContext` | `WithCompressionModel(...)` | LLM-based compression |

## Error Handling

```go
resp, err := ai.Ask(ctx, "query")
if err != nil {
    var pipelineErr *ihandai.PipelineError
    if errors.As(err, &pipelineErr) {
        // Step identifies where in the pipeline it failed
        switch pipelineErr.Step {
        case "embed":
            log.Printf("Embedding step failed: %v", pipelineErr.Err)
        case "search":
            log.Printf("Vector search failed: %v", pipelineErr.Err)
        case "chat":
            log.Printf("LLM call failed: %v", pipelineErr.Err)
            // Drilled down — could be rate limit, auth, etc.
            var rateLimitErr *ihandai.RateLimitError
            if errors.As(pipelineErr.Err, &rateLimitErr) {
                time.Sleep(rateLimitErr.RetryAfter)
            }
        }
    }
}
```

## Performance Considerations

- `EmbedBatch()` should batch multiple texts into a single API call where possible
- Multi-query runs `Embed()` calls concurrently with a `sync.WaitGroup`
- Vector search latency depends on the store — local (Chroma, pgvector) vs cloud (Pinecone, Qdrant Cloud)
- Context compression adds a full LLM round-trip — use only for complex documents with high noise
