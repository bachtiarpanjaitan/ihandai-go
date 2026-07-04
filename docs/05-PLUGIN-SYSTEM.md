# Plugin System

## Two Layers

The plugin system has two layers:

| Layer | Audience | Purpose |
|-------|----------|---------|
| **Registry** (internal) | Plugin authors | Per-interface registries with type-safe registration |
| **Client** (user-facing) | Application developers | Single facade that wires providers together |

---

## Layer 1: Per-Interface Registries

Each interface has its own registry. This ensures **type safety** — a vector database
cannot be accidentally registered as an LLM provider.

### LLM Registry (`pkg/llm/`)

```go
// llm/registry.go

type ChatFactory func(cfg Config) (ChatCompleter, error)

func Register(name string, factory ChatFactory)
func Open(name string, opts ...Option) (ChatCompleter, error)
```

### Embedding Registry (`pkg/embedding/`)

```go
// embedding/registry.go

type EmbedderFactory func(cfg Config) (Embedder, error)

func Register(name string, factory EmbedderFactory)
func Open(name string, opts ...Option) (Embedder, error)
```

### Vector Store Registry (`pkg/vectordb/`)

```go
// vectordb/registry.go

type SearcherFactory func(cfg Config) (VectorSearcher, error)

func Register(name string, factory SearcherFactory)
func Open(name string, opts ...Option) (VectorSearcher, error)
```

### How Providers Register

A single Go module can register into multiple registries. For example, OpenAI provides
both LLM and Embedding:

```go
// plugins/openai/openai.go
package openai

import (
    "github.com/bachtiarpanjaitan/ihandai-go/pkg/llm"
    "github.com/bachtiarpanjaitan/ihandai-go/pkg/embedding"
)

func init() {
    llm.Register("openai", newChatCompleter)
    embedding.Register("openai", newEmbedder)
}
```

Users activate providers with blank imports (following `database/sql` convention):

```go
import (
    _ "github.com/bachtiarpanjaitan/ihandai-go-plugins/openai"
    _ "github.com/bachtiarpanjaitan/ihandai-go-plugins/ollama"
    _ "github.com/bachtiarpanjaitan/ihandai-go-plugins/qdrant"
)
```

---

## Layer 2: Client Facade (Root Package)

The Client provides a unified interface that discovers registered providers and wires
them together:

```go
import "github.com/bachtiarpanjaitan/ihandai-go"

ai := ihandai.New(
    ihandai.WithLLM("openai",
        llm.WithModel("gpt-4o"),
        llm.WithAPIKey(os.Getenv("OPENAI_API_KEY")),
    ),
    ihandai.WithEmbedding("ollama",
        embedding.WithModel("nomic-embed-text"),
        embedding.WithBaseURL("http://localhost:11434"),
    ),
    ihandai.WithIndexEmbedding("ollama",        // optional: different provider for indexing
        embedding.WithModel("nomic-embed-text"),
    ),
    ihandai.WithVectorStore("qdrant",
        vectordb.WithURL("http://localhost:6333"),
    ),
)
```

Behind the scenes, `ihandai.New()` calls `llm.Open("openai", ...)`, `embedding.Open("ollama", ...)`, etc.

---

## Two Usage Modes

### Mode 1: Convenience (90% use case)

One client, one-liner calls. Framework handles the pipeline.

```go
ai := ihandai.New(
    ihandai.WithLLM("openai", ...),
    ihandai.WithEmbedding("ollama", ...),
)
resp, _ := ai.Ask(ctx, "What is RAG?")
```

### Mode 2: Advanced (full control)

Access individual providers directly for custom pipelines.

```go
// Direct provider access — bypass the pipeline
chat, _ := ai.LLM()        // ChatCompleter
embed, _ := ai.Embedding()  // Embedder
store, _ := ai.VectorStore() // VectorSearcher

// Or even bypass the Client entirely
chat, _ := llm.Open("openai", llm.WithModel("gpt-4o"))
embed, _ := embedding.Open("ollama", embedding.WithModel("nomic-embed-text"))
```

---

## Provider Categories

### LLM Providers

| Name | Type | Interface |
|------|------|-----------|
| OpenAI | Cloud | `ChatCompleter`, `StreamCompleter`, `TokenCounter` |
| Anthropic | Cloud | `ChatCompleter`, `StreamCompleter` |
| Google Gemini | Cloud | `ChatCompleter`, `StreamCompleter` |
| Ollama | Local | `ChatCompleter`, `StreamCompleter` |
| Groq | Cloud | `ChatCompleter`, `StreamCompleter` |
| Azure OpenAI | Cloud | `ChatCompleter`, `StreamCompleter` |

### Embedding Providers

| Name | Type | Interface |
|------|------|-----------|
| OpenAI | Cloud | `Embedder` |
| Ollama | Local | `Embedder` |
| Cohere | Cloud | `Embedder` |
| HuggingFace | Local/Cloud | `Embedder` |
| Voyage AI | Cloud | `Embedder` |

### Vector Store Providers

| Name | Type | Interface |
|------|------|-----------|
| Qdrant | Self-hosted / Cloud | `VectorSearcher`, `VectorInserter`, `VectorDeleter` |
| pgvector | PostgreSQL extension | `VectorSearcher`, `VectorInserter`, `VectorDeleter` |
| Milvus | Self-hosted / Cloud | `VectorSearcher`, `VectorInserter`, `VectorDeleter` |
| Pinecone | Cloud | `VectorSearcher`, `VectorInserter`, `VectorDeleter` |
| Weaviate | Self-hosted / Cloud | `VectorSearcher`, `VectorInserter`, `VectorDeleter` |
| Chroma | Local | `VectorSearcher`, `VectorInserter`, `VectorDeleter` |

---

## Registry Concurrency

Registries use `sync.RWMutex` — reads (lookups during `Open()`) are lock-free in the
common case. Writes (`Register()` via `init()`) happen once at program startup, before
any goroutines access the registry.

```go
var (
    registry = make(map[string]ChatFactory)
    mu       sync.RWMutex
)

func Register(name string, factory ChatFactory) {
    mu.Lock()
    defer mu.Unlock()
    registry[name] = factory
}

func Open(name string, opts ...Option) (ChatCompleter, error) {
    mu.RLock()
    factory, ok := registry[name]
    mu.RUnlock()
    if !ok {
        return nil, fmt.Errorf("llm: unknown provider %q (did you import the plugin?)", name)
    }
    return factory(buildConfig(opts))
}
```

---

## Registry per-Interface Summary

| Package | Register Signature | Open Signature |
|---------|-------------------|----------------|
| `pkg/llm/` | `Register(name, ChatFactory)` | `Open(name, ...Option) (ChatCompleter, error)` |
| `pkg/embedding/` | `Register(name, EmbedderFactory)` | `Open(name, ...Option) (Embedder, error)` |
| `pkg/vectordb/` | `Register(name, SearcherFactory)` | `Open(name, ...Option) (VectorSearcher, error)` |

Future registries for `Tool`, `Reranker`, etc. will follow the same pattern.
