# Client

The `Client` is the root package's main type. It is your application's single entry
point to all AI capabilities.

```go
import "github.com/bachtiarpanjaitan/ihandai-go"

ai := ihandai.New(
    ihandai.WithLLM("openai", ...),
    ihandai.WithEmbedding("ollama", ...),
)
```

## Design

- **Immutable** — all configuration happens in `New()`. The Client is safe for concurrent use.
- **Facade** — wires providers into pipelines (`Ask`, `Index`, `Chat`)
- **Accessor** — exposes providers for advanced/custom use (`.LLM()`, `.StreamLLM()`, `.Embedding()`, `.VectorStore()`)
- **Closable** — `Close()` shuts down connections, drains streams

## API

```go
// Constructor
func New(opts ...Option) *Client

// Convenience — built-in pipelines
func (c *Client) Chat(ctx context.Context, query string) (*Response, error)       // LLM-only, no RAG
func (c *Client) Ask(ctx context.Context, query string) (*Response, error)         // full RAG pipeline
func (c *Client) Index(ctx context.Context, documents []Document) error

// Advanced — direct provider access
func (c *Client) LLM() llm.ChatCompleter
func (c *Client) StreamLLM() llm.StreamCompleter       // nil if provider doesn't support streaming
func (c *Client) Embedding() embedding.Embedder
func (c *Client) VectorStore() vectordb.VectorSearcher

// Lifecycle
func (c *Client) Close() error
```

### Chat — Simple LLM Without RAG

`Chat()` sends a message directly to the LLM without running the RAG pipeline.
No embedding provider or vector store is required — just an LLM.

```go
resp, err := ai.Chat(ctx, "What is the capital of France?")
// → "The capital of France is Paris."
```

`Chat()` is also used internally as the fallback when `Ask()` is called without a
configured vector store (see [Graceful Fallback](#graceful-fallback)).

### StreamLLM — Direct Streaming Access

`StreamLLM()` gives you direct access to the streaming LLM provider if it implements
`llm.StreamCompleter`. Returns `nil` otherwise.

```go
if streamer := ai.StreamLLM(); streamer != nil {
    ch, _ := streamer.ChatStream(ctx, messages)
    for chunk := range ch {
        fmt.Print(chunk.Content)
    }
}
```

For the built-in streaming RAG pipeline, use `AskStream()` instead.

### Graceful Fallback

`Ask()` and `AskConversation()` work even without RAG infrastructure:

- **`Ask()`** — when no embedding or vector store is configured, falls back to `Chat()`.
- **`AskConversation()`** — when no embedding or vector store is configured, loads history
  from memory and sends the conversation to the LLM directly, skipping retrieval.

This lets you start with simple LLM interactions and add RAG later without changing
your application code.

## Functional Options

```go
type Option func(*Config)

func WithLLM(name string, opts ...llm.Option) Option
func WithEmbedding(name string, opts ...embedding.Option) Option
func WithIndexEmbedding(name string, opts ...embedding.Option) Option  // optional: separate embedding for indexing
func WithVectorStore(name string, opts ...vectordb.Option) Option
func WithLogger(logger *slog.Logger) Option
func WithTracer(tracer trace.Tracer) Option
```

### Option Chaining

Each provider's options are nested:

```go
ihandai.New(
    // Top-level: which provider?
    ihandai.WithLLM("openai",
        // Provider-level: how to configure it?
        llm.WithModel("gpt-4o"),
        llm.WithAPIKey(os.Getenv("OPENAI_API_KEY")),
        llm.WithTimeout(30 * time.Second),
        llm.WithMaxRetries(3),
        llm.WithBaseURL("https://custom-endpoint/v1"),
    ),
    ihandai.WithEmbedding("ollama",
        embedding.WithModel("nomic-embed-text"),
        embedding.WithBaseURL("http://localhost:11434"),
    ),
    ihandai.WithVectorStore("qdrant",
        vectordb.WithURL("http://localhost:6333"),
        vectordb.WithCollection("my-documents"),
    ),
)
```

## Concurrency

```go
// Safe: one Client, many goroutines
go func() { ai.Ask(ctx, "query 1") }()
go func() { ai.Ask(ctx, "query 2") }()
go func() { ai.Index(ctx, docs) }()
```

The Client holds no mutable state after `New()`. All providers must be goroutine-safe.

## Lifecycle

```go
func main() {
    ai, err := ihandai.New(
        ihandai.WithLLM("openai", ...),
        ihandai.WithVectorStore("qdrant", ...),
    )
    if err != nil {
        log.Fatal(err)
    }
    defer ai.Close()  // drain streams, close connections

    // Use ai...
}
```

## Error Types

All errors from the Client are **typed**. Never compare error strings:

```go
resp, err := ai.Ask(ctx, "query")
if err != nil {
    var rateLimitErr *ihandai.RateLimitError
    var authErr *ihandai.AuthError
    var timeoutErr *ihandai.TimeoutError
    var providerErr *ihandai.ProviderError

    switch {
    case errors.As(err, &rateLimitErr):
        log.Printf("rate limited by %s, retry after %v", rateLimitErr.Provider, rateLimitErr.RetryAfter)
        time.Sleep(rateLimitErr.RetryAfter)
        // retry
    case errors.As(err, &authErr):
        log.Fatalf("auth failed for %s: check API key", authErr.Provider)
    case errors.As(err, &timeoutErr):
        log.Printf("timeout after %v with %s", timeoutErr.Duration, timeoutErr.Provider)
    case errors.As(err, &providerErr):
        log.Printf("provider %s returned HTTP %d: %s", providerErr.Provider, providerErr.StatusCode, providerErr.Body)
    default:
        log.Printf("unexpected error: %v", err)
    }
}
```

## Error Types Definition

```go
// root package errors.go

type RateLimitError struct {
    Provider   string
    RetryAfter time.Duration
}
func (e *RateLimitError) Error() string { ... }

type AuthError struct {
    Provider string
}
func (e *AuthError) Error() string { ... }

type TimeoutError struct {
    Provider string
    Duration time.Duration
}
func (e *TimeoutError) Error() string { ... }

type ProviderError struct {
    Provider   string
    StatusCode int
    Body       string
}
func (e *ProviderError) Error() string { ... }
```
