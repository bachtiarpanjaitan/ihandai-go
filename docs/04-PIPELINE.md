# Pipeline

The `Client` orchestrates two built-in pipelines. Every step receives `ctx context.Context`
for cancellation, timeout, and distributed tracing.

---

## Index Pipeline

```
DocumentLoader.Load(ctx, source)
    → []Document
→ TextSplitter.Split(ctx, documents)
    → []Chunk
→ Embedder.EmbedBatch(ctx, texts)
    → [][]float64
→ VectorInserter.Insert(ctx, documents)
    → error (or nil on success)
```

**Error handling**: If any step fails, the pipeline stops and returns a wrapped error
identifying which step failed. The caller can retry from the failure point.

**Provider flexibility**: The embedding provider used for indexing can differ from the
one used for querying. E.g., use Ollama (local, free) for bulk indexing, OpenAI for queries.

```go
// Build pipeline with different embedding providers
ai := ihandai.New(
    ihandai.WithIndexEmbedding("ollama", embedding.WithModel("nomic-embed-text")),
    ihandai.WithEmbedding("openai", embedding.WithModel("text-embedding-3-small")),
    ihandai.WithVectorStore("qdrant", vectordb.WithURL("http://localhost:6333")),
)

// Index uses ollama
err := ai.Index(ctx, documents)

// Ask uses openai embedding + openai LLM
resp, err := ai.Ask(ctx, "What is RAG?")
```

---

## Ask Pipeline (RAG Query)

```
Embedder.Embed(ctx, query)
    → queryVector []float64
→ VectorSearcher.Search(ctx, queryVector, opts...)
    → []ScoredDocument
→ Reranker.Rerank(ctx, query, documents)
    → []ScoredDocument (reranked)
→ PromptBuilder.Build(ctx, template, context)
    → []Message
→ ChatCompleter.Chat(ctx, messages)
    → *Response
```

**Error handling**: Each step can fail independently. Possible failures and handling:

| Step | Failure | Handling |
|------|---------|----------|
| Embed | Network timeout, rate limit | Retry with backoff, or fallback to secondary embedding provider |
| Search | Connection error | Retry, or return error to caller |
| Rerank | Model unavailable | Skip reranking, continue with raw search results |
| Build | Template error | Return configuration error (not recoverable) |
| Chat | Rate limit, auth error | Retry (rate limit), or surface to caller (auth) |

The pipeline wraps errors with context: `"pipeline: step 3 (rerank): connection refused"`.

---

## Concurrency

- `Index()` and `Ask()` are safe to call concurrently from multiple goroutines.
- Each call creates its own context chain — no shared mutable state between calls.
- The `Client` is immutable after `New()`, so goroutines share it safely.

---

## Custom Pipelines

Users who need custom pipelines can bypass `Client.Ask()` / `Client.Index()` and
compose the interfaces directly:

```go
// Advanced: manual pipeline with custom logic
embedder := ai.Embedding()
chat := ai.LLM()
store := ai.VectorStore()

vec, _ := embedder.Embed(ctx, query)
docs, _ := store.Search(ctx, vec, vectordb.WithTopK(20))

// Custom filtering before LLM call
filtered := myCustomFilter(docs)

msgs := []ihandai.Message{
    {Role: "system", Content: "You are a helpful assistant."},
    {Role: "user", Content: fmt.Sprintf("Context: %s\n\nQuestion: %s", filtered, query)},
}
resp, _ := chat.Chat(ctx, msgs)
```

---

## Future: Streaming Pipeline

When streaming is implemented (Phase 4+), `AskStream()` will return a channel:

```go
func (c *Client) AskStream(ctx context.Context, query string) (<-chan Chunk, error)
```
