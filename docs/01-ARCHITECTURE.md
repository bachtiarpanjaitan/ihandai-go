# High-Level Architecture

## Overview

```
Application (your Go backend)
  │
  └── ihandai.Client (root package, your entry point)
        │
        ├── Orchestration (convenience)
        │   ├── Index()   — document indexing pipeline
        │   └── Ask()     — query pipeline
        │
        └── Direct Access (advanced)
            ├── .LLM()         → llm.ChatCompleter
            ├── .Embedding()   → embedding.Embedder
            └── .VectorStore() → vectordb.VectorSearcher
```

## Client

The `Client` is the root package's main type. It is:

- **Created once** via `ihandai.New(...)` with functional options
- **Immutable** after creation — safe for concurrent use across goroutines
- **A facade** that wires providers into pipelines
- **An accessor** that exposes individual providers for advanced/custom use

```
ihandai.New(
    ihandai.WithLLM("openai", llm.WithModel("gpt-4o")),
    ihandai.WithEmbedding("ollama", embedding.WithModel("nomic-embed-text")),
    ihandai.WithVectorStore("qdrant", vectordb.WithURL("http://localhost:6333")),
)
```

## Pipeline (Orchestration Layer)

Two built-in pipelines orchestrate the modules:

### Index Pipeline
```
DocumentLoader → TextSplitter → Embedder → VectorInserter
```

### Ask Pipeline
```
Embedder(Query) → VectorSearcher → Reranker → PromptBuilder → ChatCompleter
```

Each step accepts `context.Context` for cancellation and tracing.
Embedding in Index and Embedding in Ask can use **different providers**.

## Modules

| Module | Interface(s) | Package | Description |
|--------|-------------|---------|-------------|
| Document Loader | `DocumentLoader` | `pkg/loader/` | Load documents from files, URLs, databases |
| Text Splitter | `TextSplitter` | `pkg/splitter/` | Split documents into chunks |
| Embedding | `Embedder` | `pkg/embedding/` | Convert text to vectors |
| Vector Store | `VectorSearcher`, `VectorInserter`, `VectorDeleter` | `pkg/vectordb/` | Store and search vectors |
| Retriever | `Retriever` | `pkg/retriever/` | Retrieval strategies (top-K, MMR, hybrid) |
| Reranker | `Reranker` | `pkg/reranker/` | Re-rank retrieved documents |
| Prompt Builder | `PromptBuilder` | `pkg/prompt/` | Build prompts from templates + context |
| LLM | `ChatCompleter`, `StreamCompleter`, `TokenCounter` | `pkg/llm/` | Chat completion, streaming, token counting |
| Tools | `Tool` | `pkg/tools/` | Function definitions for LLM tool calling |
| MCP | (client) | `pkg/mcp/` | Model Context Protocol client |
| Memory | (future — phase 6) | `pkg/memory/` | Conversation history, user preferences |
| Agent | (future — phase 7) | `pkg/agent/` | Agent loop, planner-executor, reflection |
| Workflow | (future — phase 8) | `pkg/workflow/` | Multi-step AI workflow orchestration |
| Telemetry | (observability) | `pkg/telemetry/` | OpenTelemetry tracing, metrics, structured logging |

## Cross-Cutting Concerns

- **`context.Context`** — every I/O method accepts it as first parameter
- **Structured errors** — `RateLimitError`, `AuthError`, `TimeoutError`, `ProviderError` in root package
- **Concurrency** — Client immutable after creation; providers must be goroutine-safe
- **Configuration** — functional options pattern (`ihandai.WithLLM(...)`, `llm.WithModel(...)`)
