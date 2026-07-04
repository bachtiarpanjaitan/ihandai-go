# Core Interfaces

> **Rule**: All interfaces have 1-3 methods and follow the `-er` suffix convention.
> **Rule**: Every I/O method accepts `ctx context.Context` as its first parameter.
> **Rule**: Implementations satisfy interfaces — never depend on concrete types.

---

## LLM (`pkg/llm/`)

```go
type ChatCompleter interface {
    Chat(ctx context.Context, messages []Message) (*Response, error)
}

type StreamCompleter interface {
    ChatStream(ctx context.Context, messages []Message) (<-chan Chunk, error)
}

type TokenCounter interface {
    CountTokens(ctx context.Context, model, text string) (int, error)
}
```

One provider can implement all three interfaces. Consumers declare only what they need:
```go
func simple(c ChatCompleter)          { ... }
func withStream(c StreamCompleter)    { ... }
func advanced(c ChatCompleter, t TokenCounter) { ... }
```

---

## Embedding (`pkg/embedding/`)

```go
type Embedder interface {
    Embed(ctx context.Context, text string) ([]float64, error)
    EmbedBatch(ctx context.Context, texts []string) ([][]float64, error)
}
```

> Note: `Embed` for single texts, `EmbedBatch` for bulk. Providers may batch API calls
> internally for `EmbedBatch` to reduce latency.

---

## Vector Store (`pkg/vectordb/`)

```go
type VectorSearcher interface {
    Search(ctx context.Context, vector []float64, opts ...SearchOption) ([]ScoredDocument, error)
}

type VectorInserter interface {
    Insert(ctx context.Context, documents []Document) error
}

type VectorDeleter interface {
    Delete(ctx context.Context, ids []string) error
}
```

> RAG indexing needs `VectorInserter`. RAG querying needs `VectorSearcher`.
> Simple stores implement all three; read-only replicas might only implement `VectorSearcher`.

---

## Document Loader (`pkg/loader/`)

```go
type DocumentLoader interface {
    Load(ctx context.Context, source string) ([]Document, error)
}
```

Implementations: file loader, URL loader, PDF parser, database loader.

---

## Text Splitter (`pkg/splitter/`)

```go
type TextSplitter interface {
    Split(ctx context.Context, documents []Document) ([]Chunk, error)
}
```

Implementations: recursive character split, token-based split, semantic split.

---

## Retriever (`pkg/retriever/`)

```go
type Retriever interface {
    Retrieve(ctx context.Context, query []float64, opts ...RetrieveOption) ([]ScoredDocument, error)
}
```

Wraps `VectorSearcher` with strategy: top-K, MMR, hybrid, multi-query.

---

## Reranker (`pkg/reranker/`)

```go
type Reranker interface {
    Rerank(ctx context.Context, query string, documents []Document) ([]ScoredDocument, error)
}
```

Cross-encoder or LLM-based re-ranking for improved result quality.

---

## Prompt Builder (`pkg/prompt/`)

```go
type PromptBuilder interface {
    Build(ctx context.Context, template string, context map[string]any) ([]Message, error)
}
```

Builds the final prompt from templates, retrieved documents, and conversation history.
Supports chat message formatting (system, user, assistant roles).

---

## Tool (`pkg/tools/`)

```go
type Tool interface {
    Name() string
    Description() string
    InputSchema() *JSONSchema
    Execute(ctx context.Context, input json.RawMessage) (json.RawMessage, error)
}
```

LLM function calling. One tool = one function the LLM can request.
See `docs/12-TOOLS.md` for details.

---

## Future Interfaces (Defined in Later Phases)

### Memory (`pkg/memory/`) — Phase 6
```go
type ConversationStore interface {
    Append(ctx context.Context, msg Message) error
    History(ctx context.Context, limit int) ([]Message, error)
}
```

### Agent (`pkg/agent/`) — Phase 7
```go
type Planner interface { ... }
type Executor interface { ... }
```

### Workflow (`pkg/workflow/`) — Phase 8
```go
type WorkflowRunner interface { ... }
```

---

## Summary

| Interface | Package | Methods |
|-----------|---------|---------|
| `ChatCompleter` | `pkg/llm/` | `Chat` |
| `StreamCompleter` | `pkg/llm/` | `ChatStream` |
| `TokenCounter` | `pkg/llm/` | `CountTokens` |
| `Embedder` | `pkg/embedding/` | `Embed`, `EmbedBatch` |
| `VectorSearcher` | `pkg/vectordb/` | `Search` |
| `VectorInserter` | `pkg/vectordb/` | `Insert` |
| `VectorDeleter` | `pkg/vectordb/` | `Delete` |
| `DocumentLoader` | `pkg/loader/` | `Load` |
| `TextSplitter` | `pkg/splitter/` | `Split` |
| `Retriever` | `pkg/retriever/` | `Retrieve` |
| `Reranker` | `pkg/reranker/` | `Rerank` |
| `PromptBuilder` | `pkg/prompt/` | `Build` |
| `Tool` | `pkg/tools/` | `Name`, `Description`, `InputSchema`, `Execute` |
