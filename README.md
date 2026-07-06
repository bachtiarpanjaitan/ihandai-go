# ihandai — Go AI Framework v1.0.0

> Provider-agnostic Go AI library for building AI-powered applications.

```go
import (
    "github.com/bachtiarpanjaitan/ihandai-go"
    _ "github.com/bachtiarpanjaitan/ihandai-go/plugins/ollama"
)

ai, _ := ihandai.New(
    ihandai.WithLLM("ollama", llm.WithModel("llama3")),
    ihandai.WithEmbedding("ollama", embedding.WithModel("nomic-embed-text")),
    ihandai.WithMemory(memory.NewInMemoryStore()),
)
defer ai.Close()

// Simple LLM chat (no RAG required)
resp, _ := ai.Chat(ctx, "Hello, who are you?")

// RAG query
resp, _ := ai.Ask(ctx, "What is RAG?", ihandai.WithTopK(5))

// Ask() gracefully falls back to Chat() when no vector store is configured

// Multi-turn conversation
resp, _ := ai.AskConversation(ctx, "user-123", "Hello!")

// Document indexing
ai.Index(ctx, "./documents/")

// Autonomous agent
ai.Run(ctx, "Research the latest Go libraries")

// Direct streaming provider access
streamAI := ai.StreamLLM() // nil if provider doesn't support streaming
```

## Features

| Feature | Package | Description |
|---------|---------|-------------|
| **Simple Chat** | root | `Chat()` — LLM call without RAG, no vector store needed |
| **RAG Pipeline** | root | Load → Split → Embed → Search → Rerank → Chat |
| **Retrieval Strategies** | `pkg/retriever` | TopK, MMR, MultiQuery, Hybrid |
| **Agents** | `pkg/agent` | ReAct pattern, tool calling, retry |
| **Memory** | `pkg/memory` | Conversation store, window management |
| **Workflows** | `pkg/workflow` | DAG-based parallel execution, conditional branching |
| **MCP Client** | `pkg/mcp` | JSON-RPC client, filesystem server |
| **Streaming** | root | `AskStream()` + `StreamLLM()` accessor for channel-based response |
| **Observability** | `pkg/telemetry` | Tracing, rate limiter, circuit breaker |

## Packages (18 packages, 100% test pass, race detector clean)

```
pkg/
├── agent/        ReAct agent with tool calling
├── core/         Shared types (Message, Document, etc.)
├── embedding/    Embedder interface + registry
├── llm/          ChatCompleter, StreamCompleter, TokenCounter + registry
├── loader/       DocumentLoader + file loader
├── mcp/          MCP client + filesystem server
├── memory/       ConversationStore + window manager
├── prompt/       PromptBuilder + simple RAG prompt
├── reranker/     Reranker interface
├── retriever/    TopK, MMR, MultiQuery strategies
├── splitter/     TextSplitter + recursive splitter
├── telemetry/    Tracer, RateLimiter, CircuitBreaker
├── tools/        Tool interface + registry
├── vectordb/     VectorSearcher, Inserter, Deleter + registry
└── workflow/     DAG workflow engine
plugins/
└── ollama/       Ollama LLM + Embedding provider
```

## Core Principles
- **Interface-first** — small interfaces (1-3 methods), idiomatic Go
- **Provider-agnostic** — registry pattern, swap providers without code changes
- **Context-aware** — `context.Context` everywhere
- **Concurrency-safe** — `sync.RWMutex` on shared state, race detector clean
- **Structured errors** — typed errors for rate limits, auth, timeouts
- **Pluggable** — `init()` registration + blank import (`database/sql` pattern)
- **Production-ready** — tracing, rate limiting, circuit breaker
