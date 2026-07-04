# Vision

Build the Go equivalent of a modern AI application framework — a library you `go get`
and integrate into any backend, with the flexibility to swap providers without
rewriting your application code.

## Core Principles

- **Interface-first** — small, composable interfaces (1-3 methods). Accept interfaces, return structs.
- **Provider-agnostic** — registry pattern. Swap OpenAI for Ollama without application changes.
- **Composable providers via functional options** — no heavy DI framework; just `ihandai.New(ihandai.WithLLM("openai", ...))`.
- **Extensible plugins** — third-party providers register via `init()` + blank import, following `database/sql` conventions.
- **Context-aware** — every I/O operation accepts `context.Context` as its first parameter. Cancellation, timeouts, and tracing work out of the box.
- **Concurrency-safe** — all public APIs are safe for concurrent use. The Client is immutable after creation.
- **Structured errors** — typed errors (`RateLimitError`, `AuthError`, `TimeoutError`, `ProviderError`). No string comparison for error handling.
- **Testability** — small interfaces mean trivial mocking. All I/O surfaces accept interfaces.
- **Production-ready** — OpenTelemetry tracing, structured logging, and metrics built in from day one.
