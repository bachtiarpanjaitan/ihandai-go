# Roadmap

## Overview

| Phase | Name | Depends On | Key Deliverables |
|-------|------|------------|------------------|
| 1 | Core | — | `go.mod`, root package, `Client` struct, `Config`, `Option` types, error types (`RateLimitError`, `AuthError`, `TimeoutError`, `ProviderError`) |
| 2 | Interfaces | Phase 1 | All small interfaces: `ChatCompleter`, `StreamCompleter`, `TokenCounter`, `Embedder`, `VectorSearcher`, `VectorInserter`, `VectorDeleter`, `DocumentLoader`, `TextSplitter`, `Retriever`, `Reranker`, `PromptBuilder`, `Tool` |
| 3 | Plugins | Phase 2 | Registry pattern (`llm.Register`, `embedding.Register`, `vectordb.Register`). 2 providers per interface: OpenAI + Ollama for LLM/Embedding, Qdrant + pgvector for VectorDB. `init()` + blank import pattern |
| 4 | Pipeline | Phase 3 | `Client.Ask()` and `Client.Index()` orchestration. Streaming: `Client.AskStream()`. Functional options. Error wrapping with step context |
| 5 | RAG | Phase 4 | Full RAG flow end-to-end. Retrieval strategies: top-K, MMR, hybrid search, multi-query, parent document retrieval, context compression. Metadata filtering |
| 6 | Memory | Phase 5 | Conversation history, user preferences, `ConversationStore` interface. Context window management |
| 7 | Agents | Phase 6 | Agent loop (ReAct), planner-executor, reflection, retry. **Tool Calling** framework: tool registry, execution sandbox. Memory integration |
| 8 | Workflows | Phase 7 | Multi-step AI pipeline orchestration. DAG-based workflow definitions, conditional branching, parallel execution |
| 9 | MCP | Phase 7 | MCP client: stdio + HTTP/SSE transports. Filesystem, GitHub, PostgreSQL servers. Resource discovery |
| 10 | Production | Phase 9 | OpenTelemetry tracing, structured logging (`slog`), Prometheus metrics. Rate limiting, circuit breaker. Security hardening. Stable v1.0 release |

## Phase Detail

### Phase 1: Core
- Initialize Go module: `go mod init github.com/bachtiarpanjaitan/ihandai-go`
- Define root package types: `Client`, `Config`, `Option`
- Define error types: `RateLimitError`, `AuthError`, `TimeoutError`, `ProviderError`
- Define shared types: `Message`, `Response`, `Document`, `Chunk`, `ScoredDocument`
- CI/CD: linting, test coverage, Go version matrix

### Phase 2: Interfaces
- Define all 13 interfaces across their packages
- Each interface: 1-3 methods, `ctx context.Context` as first parameter
- Godoc for every exported type
- Mock implementations for testing

### Phase 3: Plugins
- Implement registry pattern in `pkg/llm/`, `pkg/embedding/`, `pkg/vectordb/`
- Provider plugins (separate repo): `openai`, `ollama`, `qdrant`, `pgvector`
- Provider configuration: API keys, base URLs, model selection
- Integration tests against real providers

### Phase 4: Pipeline
- `Client.Ask()` — full query pipeline
- `Client.Index()` — full indexing pipeline
- `Client.AskStream()` — streaming query pipeline
- Error handling: step-level error wrapping
- Context propagation through all pipeline steps

### Phase 5: RAG
- Retrieval strategies: top-K similarity, MMR (Maximal Marginal Relevance)
- Hybrid search: vector + keyword (BM25)
- Multi-query: generate multiple query variants and merge results
- Parent document retrieval: retrieve parent chunks for context
- Context compression: LLM-based context summarization
- Metadata filtering: filter by date, author, source

### Phase 6: Memory
- `ConversationStore` interface: append, retrieve, trim
- In-memory implementation for development
- Persisted implementation (SQLite, PostgreSQL)
- Context window management: automatic trimming when approaching token limits

### Phase 7: Agents
- Agent loop: plan → execute → observe → reflect → next plan
- Tool calling framework: `Tool` interface, tool registry, execution sandbox
- Built-in tools: `Browser`, `Slack`, `Email`, `HTTP Request`, `Code Executor`
- Retry with backoff for failed tool calls
- Memory integration: agents maintain conversation history

### Phase 8: Workflows
- DAG (Directed Acyclic Graph) based workflow definitions
- Conditional branching: if-then-else based on step outputs
- Parallel execution: independent steps run concurrently
- Workflow serialization: save/load workflow definitions
- Retry and error handling per step

### Phase 9: MCP
- MCP client implementation: stdio and HTTP/SSE transports
- Server registration and discovery
- MCP servers: `filesystem`, `github`, `postgresql`, `slack`
- Resource listing and reading
- Integration with Agent tool calling (agents can use MCP resources)

### Phase 10: Production
- OpenTelemetry tracing: spans for each pipeline step
- Structured logging via `slog`
- Prometheus metrics: request latency, error rates, token usage
- Rate limiting: per-provider rate limit awareness
- Circuit breaker: fail-fast when a provider is unhealthy
- Security: credential management, input sanitization
- Stable API: freeze public interfaces, semver
- Documentation: full godoc, tutorials, migration guides
