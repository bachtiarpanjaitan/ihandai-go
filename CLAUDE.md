# CLAUDE.md

## Project Context

ihandai is a provider-agnostic Go AI library designed for RAG-first, incremental delivery.
It provides interfaces for LLM, embedding, vector stores, and orchestrates them into RAG pipelines.

## Core Principles

- **Small interfaces** — 1-3 methods per interface, idiomatic Go
- **context.Context everywhere** — all I/O functions accept context as first parameter
- **Structured errors** — typed errors, never string comparison
- **Concurrency-safe** — all public APIs must be safe for concurrent use
- **Provider-agnostic** — interface-first, registry pattern (see ADR-0001)

## AI Development Roles

When working on this project, adopt the following roles as needed:

- **System Architect** — design interfaces, architecture decisions, trade-off analysis
- **API Designer** — public API surface, developer experience, naming conventions
- **Core Engineer** — implementation of core interfaces and pipelines
- **Plugin Engineer** — implementation of provider adapters (OpenAI, Ollama, Qdrant, etc.)
- **Documentation Engineer** — godoc comments, examples, README updates
- **QA Engineer** — test coverage, edge cases, benchmarking
- **Release Engineer** — versioning, changelog, backwards compatibility checks

## References

- Architecture Decision Records in `adr/`
- Full specification in `docs/`
- No code yet — this project is in the design phase
