---
title: ihandai — Go AI Framework v1.0
description: Provider-agnostic Go library for RAG, agents, workflows, and production AI
---

# Provider-Agnostic Go AI Library

Build AI-powered Go applications with swappable providers, modular interfaces, and production-ready tooling.

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

resp, _ := ai.Ask(ctx, "What is RAG?", ihandai.WithTopK(5))
```

## Installation

```bash
go get github.com/bachtiarpanjaitan/ihandai-go
```

## Features

| Feature | Description |
|---------|-------------|
| 💬 **Simple Chat** | `Chat()` — LLM call without RAG, no vector store required |
| 🔍 **RAG Pipeline** | Load → Split → Embed → Search → Rerank → Chat |
| 🎯 **Retrieval Strategies** | TopK, MMR (diversity), MultiQuery (expansion) |
| 🤖 **Agents** | ReAct agent loop with tool calling, retry, reflection |
| 🧠 **Memory** | Multi-turn conversations with token-aware window |
| ⚙️ **Workflows** | DAG-based parallel execution, conditional branching |
| 🔌 **MCP** | Model Context Protocol client + filesystem server |
| 📡 **Streaming** | `AskStream()` + `StreamLLM()` accessor for real-time tokens |
| 🛡️ **Production** | Rate limiter, circuit breaker, tracing |

## Quick Links

- [📖 **Usage Guide**](GUIDE) — complete API reference with examples
- [🏗️ **Architecture**](01-ARCHITECTURE) — high-level design
- [📦 **Package Structure**](02-PACKAGE-STRUCTURE) — package layout
- [🧩 **Interfaces**](03-INTERFACES) — core interface definitions
- [🔄 **Pipeline**](04-PIPELINE) — data flow design
- [🔌 **Plugin System**](05-PLUGIN-SYSTEM) — provider registration
- [🗺️ **Roadmap**](10-ROADMAP) — development phases

## Supported Providers

### LLM & Embedding

| Provider | Type | Interface |
|----------|------|-----------|
| Ollama | Local | ChatCompleter, Embedder |
| OpenAI | Cloud | ChatCompleter, Embedder (planned) |
| Anthropic | Cloud | ChatCompleter (planned) |
| Google Gemini | Cloud | ChatCompleter (planned) |

### Vector Stores

| Provider | Type |
|----------|------|
| Qdrant | Self-hosted / Cloud (planned) |
| pgvector | PostgreSQL extension (planned) |
| Milvus | Self-hosted / Cloud (planned) |

## Core Principles

- **Interface-first** — small interfaces (1-3 methods), idiomatic Go
- **Provider-agnostic** — `database/sql`-style registry, swap without code changes
- **Context-aware** — `context.Context` as first parameter in all I/O
- **Concurrency-safe** — `sync.RWMutex`, immutable-after-creation
- **Structured errors** — typed errors: `RateLimitError`, `AuthError`, `TimeoutError`
- **Production-ready** — tracing, rate limiting, circuit breaker

## Quick Example: RAG Chatbot

```go
package main

import (
    "bufio"
    "context"
    "fmt"
    "log"
    "os"

    "github.com/bachtiarpanjaitan/ihandai-go"
    "github.com/bachtiarpanjaitan/ihandai-go/pkg/llm"
    "github.com/bachtiarpanjaitan/ihandai-go/pkg/embedding"
    "github.com/bachtiarpanjaitan/ihandai-go/pkg/memory"
    _ "github.com/bachtiarpanjaitan/ihandai-go/plugins/ollama"
)

func main() {
    ai, _ := ihandai.New(
        ihandai.WithLLM("ollama", llm.WithModel("llama3")),
        ihandai.WithEmbedding("ollama", embedding.WithModel("nomic-embed-text")),
        ihandai.WithVectorStore("mock"),
        ihandai.WithMemory(memory.NewInMemoryStore()),
    )
    defer ai.Close()

    scanner := bufio.NewScanner(os.Stdin)
    session := "default"
    for {
        fmt.Print("> ")
        scanner.Scan()
        resp, _ := ai.AskConversation(context.Background(), session, scanner.Text())
        fmt.Println(resp.Content)
    }
}
```

## License & Contributing

- 📄 [Contributing Guide](https://github.com/bachtiarpanjaitan/ihandai-go/blob/main/.github/CONTRIBUTING.md)
- 📐 [Architecture Decisions](https://github.com/bachtiarpanjaitan/ihandai-go/tree/main/adr)
- 🧪 18 packages, 100% test pass, race detector clean
