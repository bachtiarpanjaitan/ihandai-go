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

- **Removed `core/`** — was undefined. Its role is now served by the root package.
- **Removed `rag/`** — RAG is a pipeline, not a package. Orchestrated by `Client`.
- **Added `prompt/`** — PromptBuilder was referenced in architecture but had no package.
- **Added `telemetry/`** — structured logging, OpenTelemetry tracing, metrics.

## Plugins (Separate Repository)

Plugins are separate Go modules that import the core library and register providers:

```
github.com/bachtiarpanjaitan/ihandai-go-plugins/
├── openai/       OpenAI LLM + Embedding
├── ollama/       Ollama LLM + Embedding (local)
├── gemini/       Google Gemini LLM
├── claude/       Anthropic Claude LLM
├── qdrant/       Qdrant vector database
├── pgvector/     PostgreSQL pgvector
├── milvus/       Milvus vector database
└── ...
```

Users import plugins via blank import:
```go
import _ "github.com/bachtiarpanjaitan/ihandai-go-plugins/openai"
```
