# Package Structure

## Root Package (`github.com/bachtiarpanjaitan/ihandai-go`)

The root package is the primary entry point for users. It contains:

- `Client` — the main facade
- `New()` — constructor with functional options
- `Option`, `Config` — configuration types
- Error types: `RateLimitError`, `AuthError`, `TimeoutError`, `ProviderError`

```go
import "github.com/bachtiarpanjaitan/ihandai-go"

ai := ihandai.New(
    ihandai.WithLLM("openai", llm.WithModel("gpt-4o")),
    ihandai.WithEmbedding("ollama", embedding.WithModel("nomic-embed-text")),
)
```

## Sub-Packages (`pkg/`)

Sub-packages provide the interfaces that providers implement, plus the registry
machinery for provider registration. Users only import sub-packages for **advanced** use (direct provider access, custom pipeline logic).

```
pkg/
├── core/          Message, Response, Document, Chunk, ScoredDocument,
│                  JSONSchema, ToolCall, TokenUsage (shared types)
├── llm/           ChatCompleter, StreamCompleter, TokenCounter interfaces
│                  Registry: llm.Register(), llm.Open()
│
├── embedding/     Embedder interface
│                  Registry: embedding.Register(), embedding.Open()
│
├── vectordb/      VectorSearcher, VectorInserter, VectorDeleter interfaces
│                  Registry: vectordb.Register(), vectordb.Open()
│
├── loader/        DocumentLoader interface
├── splitter/      TextSplitter interface
├── retriever/     Retriever interface
├── reranker/      Reranker interface
├── prompt/        PromptBuilder interface
├── tools/         Tool interface (for LLM function calling)
├── mcp/           MCP client implementation (future — phase 9)
├── memory/        Memory interfaces (future — phase 6)
├── agent/         Agent interfaces (future — phase 7)
├── workflow/      Workflow interfaces (future — phase 8)
└── telemetry/     Observability: tracing, metrics, structured logging
```

## What Changed

- **Added `core/`** — shared types needed by both root and sub-packages, avoiding import cycles.
- **Removed old `core/`** — was undefined; now has a concrete purpose.
- **Removed `rag/`** — RAG is a pipeline, not a package. Orchestrated by `Client`.
- **Added `prompt/`** — PromptBuilder was referenced in architecture but had no package.
- **Added `telemetry/`** — structured logging, OpenTelemetry tracing, metrics.

## Plugins (`plugins/`)

Reference provider implementations live in the `plugins/` directory:

```
plugins/
└── ollama/        Ollama LLM + Embedding (local, no API key)
    ├── ollama.go   shared HTTP client, config, helpers
    ├── chat.go     ChatCompleter implementation
    └── embedding.go Embedder implementation
```

Future providers (openai, qdrant, etc.) may be extracted to a separate repository
(`github.com/bachtiarpanjaitan/ihandai-go-plugins`).

Users activate plugins via blank import:
```go
import _ "github.com/bachtiarpanjaitan/ihandai-go/plugins/ollama"

chat, _ := llm.Open("ollama", llm.WithModel("llama3"))
embed, _ := embedding.Open("ollama", embedding.WithModel("nomic-embed-text"))
```
