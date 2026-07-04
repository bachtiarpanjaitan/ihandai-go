# Contributing

## Rules

- **Keep public APIs backward compatible.** Breaking changes require a major version bump and an ADR.
- **Prefer interfaces over concrete implementations.** Consume interfaces, return structs.
- **Interfaces should have 1-3 methods.** If you need more, split the interface using composition.
- **All public functions that perform I/O must accept `context.Context` as the first parameter.** No exceptions.
- **Add tests for every feature.** Unit tests for logic, integration tests for providers, benchmarks for I/O paths.
- **Document every exported type.** Godoc comments are mandatory. Examples for public APIs.
- **Use structured errors.** Never return `fmt.Errorf("...")` for errors the caller might want to handle programmatically. Use typed errors from the root package.
- **Concurrency-safe by default.** All public types must document their concurrency guarantees. When in doubt, use `sync.RWMutex` or make the type immutable after construction.

## Code Style

- Follow [Effective Go](https://go.dev/doc/effective_go) and [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments).
- Package names: lowercase, single word, no underscores. Avoid generic names (`util`, `common`, `helper`).
- Interface names: single-method interfaces use `-er` suffix (`Reader`, `Embedder`, `Completer`).
- Functional options pattern for constructors with many optional parameters.
- Accept interfaces, return structs.

## Commit Messages

- Conventional commits: `feat:`, `fix:`, `docs:`, `refactor:`, `test:`, `chore:`
- Reference ADR numbers for architecture changes.
