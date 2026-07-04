# Testing

## Test Categories

### Unit Tests
Every exported function and method must have unit tests. The small-interface design
makes mocking trivial:

```go
type mockChat struct{}

func (m mockChat) Chat(ctx context.Context, messages []ihandai.Message) (*ihandai.Response, error) {
    return &ihandai.Response{Content: "mock response"}, nil
}

func TestClient_Ask(t *testing.T) {
    ai := ihandai.New(
        ihandai.WithLLMFromInstance(mockChat{}),  // inject mock directly
    )
    resp, err := ai.Ask(context.Background(), "test")
    assert.NoError(t, err)
    assert.Equal(t, "mock response", resp.Content)
}
```

### Table-Driven Tests
Go's standard pattern. All provider interface implementations should be tested with
a shared table-driven test suite:

```go
func TestChatCompleter(t *testing.T) {
    tests := []struct {
        name     string
        messages []ihandai.Message
        want     string
        wantErr  bool
    }{
        {"simple query", simpleMessages, "expected", false},
        {"empty input", nil, "", true},
        {"cancelled context", simpleMessages, "", true},
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // test each provider against the same cases
        })
    }
}
```

### Integration Tests
Test against real provider instances (requires API keys or local services):

```go
func TestOpenAI_Integration(t *testing.T) {
    if os.Getenv("OPENAI_API_KEY") == "" {
        t.Skip("OPENAI_API_KEY not set")
    }
    chat, err := llm.Open("openai", llm.WithModel("gpt-3.5-turbo"))
    require.NoError(t, err)
    resp, err := chat.Chat(context.Background(), []ihandai.Message{
        {Role: "user", Content: "Say hello"},
    })
    assert.NoError(t, err)
    assert.NotEmpty(t, resp.Content)
}
```

### Compatibility Tests
Verify the library works across:

- **Go versions**: latest 2 stable releases (e.g., Go 1.22, Go 1.23)
- **Provider versions**: each provider plugin pinned to a known-compatible API version
- **Architectures**: `linux/amd64`, `linux/arm64`, `darwin/amd64`, `darwin/arm64`

Run via CI matrix:
```yaml
# .github/workflows/compat.yml
strategy:
  matrix:
    go-version: ["1.22", "1.23"]
    platform: [ubuntu-latest, macos-latest]
```

### Benchmarks
Every I/O path must have benchmarks:

```go
func BenchmarkChatCompleter_Chat(b *testing.B) { ... }
func BenchmarkEmbedder_Embed(b *testing.B) { ... }
func BenchmarkVectorSearcher_Search(b *testing.B) { ... }
```

### Load Tests
End-to-end pipeline benchmarks simulating concurrent users:

```go
func BenchmarkPipeline_Concurrent(b *testing.B) {
    b.RunParallel(func(pb *testing.PB) {
        for pb.Next() {
            ai.Ask(ctx, "query")
        }
    })
}
```

## Context Cancellation Tests

```go
func TestClient_Ask_Cancelled(t *testing.T) {
    ctx, cancel := context.WithCancel(context.Background())
    cancel() // immediately cancel
    _, err := ai.Ask(ctx, "query")
    assert.Error(t, err)
    assert.ErrorIs(t, err, context.Canceled)
}

func TestClient_Ask_Timeout(t *testing.T) {
    ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
    defer cancel()
    time.Sleep(10 * time.Millisecond) // exceed timeout
    _, err := ai.Ask(ctx, "query")
    assert.Error(t, err)
    assert.ErrorIs(t, err, context.DeadlineExceeded)
}
```

## Mocking with Small Interfaces

Small interfaces enable targeted mocking:

```go
// Only mock what you need — 1 method, not 5
type failingEmbedder struct{}
func (f failingEmbedder) Embed(ctx context.Context, text string) ([]float64, error) {
    return nil, errors.New("embedding service down")
}
func (f failingEmbedder) EmbedBatch(ctx context.Context, texts []string) ([][]float64, error) {
    return nil, errors.New("embedding service down")
}

func TestPipeline_EmbeddingFails(t *testing.T) {
    ai := ihandai.New(
        ihandai.WithLLMFromInstance(mockChat{}),
        ihandai.WithEmbeddingFromInstance(failingEmbedder{}), // test failure path
    )
    _, err := ai.Ask(context.Background(), "query")
    assert.Error(t, err)
    assert.Contains(t, err.Error(), "embedding service down")
}
```
