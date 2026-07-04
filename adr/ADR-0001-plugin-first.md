# ADR-0001: Plugin-First Architecture

## Status
Accepted

## Decision
Adopt plugin-first architecture — interface-driven design with provider registration.

## What "Plugin-First" Means

"Plugin-first" in this project refers to **interface-first design with a registry
pattern**, not Go's native `plugin` package (which has platform and versioning limitations).

This is the same pattern used by Go's `database/sql`:

```go
import _ "github.com/lib/pq"          // registers "postgres" driver
db, _ := sql.Open("postgres", connStr) // uses the registered driver
```

Applied to ihandai:

```go
import _ "github.com/bachtiarpanjaitan/ihandai-go-plugins/openai"  // registers "openai" provider
chat, _ := llm.Open("openai", ...)                        // uses the registered provider
```

## Why

1. **Avoid coupling to any single AI provider.** Applications swap OpenAI for Ollama
   without changing business logic.
2. **Enable third-party extensions.** Anyone can write and distribute a provider plugin.
3. **Testability.** Mock implementations replace real providers in tests.
4. **Progressive disclosure.** Users start with `ihandai.New(...)` and only dive into
   individual registries when they need advanced control.

## Consequences

- Each interface needs its own registry (`llm.Register`, `embedding.Register`, `vectordb.Register`).
- Providers must be distributed as separate Go modules (not in the core library).
- Breaking changes to interfaces affect all provider implementations — interfaces must
  be designed carefully from the start.
- `init()` functions run at program startup; registration must be lightweight and
  never fail.
