# ihandai — Usage Guide

## Installation

```bash
go get github.com/bachtiarpanjaitan/ihandai-go
```

For Ollama (local, free):
```bash
go get github.com/bachtiarpanjaitan/ihandai-go/plugins/ollama
```

---

## Quick Start

```go
package main

import (
    "context"
    "fmt"
    "log"

    "github.com/bachtiarpanjaitan/ihandai-go"
    "github.com/bachtiarpanjaitan/ihandai-go/pkg/llm"
    "github.com/bachtiarpanjaitan/ihandai-go/pkg/embedding"
    _ "github.com/bachtiarpanjaitan/ihandai-go/plugins/ollama"
)

func main() {
    ai, err := ihandai.New(
        ihandai.WithLLM("ollama",
            llm.WithModel("llama3"),
            llm.WithBaseURL("http://localhost:11434"),
        ),
        ihandai.WithEmbedding("ollama",
            embedding.WithModel("nomic-embed-text"),
        ),
    )
    if err != nil {
        log.Fatal(err)
    }
    defer ai.Close()

    ctx := context.Background()
    resp, err := ai.Ask(ctx, "What is the capital of France?")
    if err != nil {
        log.Fatal(err)
    }
    fmt.Println(resp.Content)
}
```

---

## Simple Chat (No RAG)

`Chat()` sends a message directly to the LLM without any RAG pipeline.
Only an LLM provider is required — no embedding or vector store needed.

```go
ai, err := ihandai.New(
    ihandai.WithLLM("ollama",
        llm.WithModel("llama3"),
        llm.WithBaseURL("http://localhost:11434"),
    ),
)
if err != nil {
    log.Fatal(err)
}
defer ai.Close()

resp, err := ai.Chat(ctx, "What is the capital of France?")
if err != nil {
    log.Fatal(err)
}
fmt.Println(resp.Content)
// → "The capital of France is Paris."
```

You can also use `Ask()` without a vector store — it falls back to `Chat()` automatically,
so you can start with simple LLM calls and add RAG later without changing call sites.

---

## Configuration

### Functional Options Pattern

```go
ai, _ := ihandai.New(
    // LLM — chat completion
    ihandai.WithLLM("openai",
        llm.WithModel("gpt-4o"),
        llm.WithAPIKey(os.Getenv("OPENAI_API_KEY")),
        llm.WithBaseURL("https://api.openai.com/v1"),
    ),

    // Embedding — text → vectors
    ihandai.WithEmbedding("ollama",
        embedding.WithModel("nomic-embed-text"),
        embedding.WithBaseURL("http://localhost:11434"),
    ),

    // Optional: separate embedding for bulk indexing
    ihandai.WithIndexEmbedding("ollama",
        embedding.WithModel("nomic-embed-text"),
    ),

    // Vector store
    ihandai.WithVectorStore("qdrant",
        vectordb.WithURL("http://localhost:6333"),
        vectordb.WithCollection("my-documents"),
    ),

    // Conversation memory
    ihandai.WithMemory(memory.NewInMemoryStore()),

    // Agent tools
    ihandai.WithTools(
        agenttools.NewCalculator(),
        agenttools.NewHTTPRequest(),
    ),
)
```

### Direct Provider Access (Advanced)

```go
import (
    "github.com/bachtiarpanjaitan/ihandai-go/pkg/llm"
    "github.com/bachtiarpanjaitan/ihandai-go/pkg/embedding"
    _ "github.com/bachtiarpanjaitan/ihandai-go/plugins/ollama"
)

// Bypass Client — use providers directly
chat, _ := llm.Open("ollama", llm.WithModel("llama3"))
embed, _ := embedding.Open("ollama", embedding.WithModel("nomic-embed-text"))

resp, _ := chat.Chat(ctx, []ihandai.Message{
    {Role: "user", Content: "Hello!"},
})
```

---

## RAG Pipeline

### Document Indexing

```go
// Index a single file
err := ai.Index(ctx, "./documents/report.txt")

// Index a directory (all .txt and .md files)
err := ai.Index(ctx, "./documents/")

// Custom loader and splitter
err := ai.Index(ctx, "./docs/",
    ihandai.WithLoader(myCustomLoader),
    ihandai.WithSplitter(myCustomSplitter),
)
```

### Querying

```go
// Basic query (falls back to Chat() if no vector store is configured)
resp, err := ai.Ask(ctx, "What does the report say about revenue?")

// With options
resp, err := ai.Ask(ctx, "What is RAG?",
    ihandai.WithTopK(10),                                    // more results
    ihandai.WithFilter(map[string]any{"source": "docs/"}),   // metadata filter
)
```

### Retrieval Strategies

```go
import "github.com/bachtiarpanjaitan/ihandai-go/pkg/retriever"

// MMR — diversity-aware retrieval
mmr := retriever.NewMMR(ai.VectorStore(), 0.7) // λ=0.7
resp, err := ai.Ask(ctx, "summarize the findings",
    ihandai.WithRetriever(mmr),
    ihandai.WithTopK(5),
)

// Multi-Query — query expansion
multiQ := retriever.NewMultiQuery(
    ai.VectorStore(),
    ai.LLM(),        // for generating variants
    ai.Embedding(),  // for embedding variants
    3,               // number of variants
)
resp, err := ai.Ask(ctx, "complex question",
    ihandai.WithRetriever(multiQ),
)
```

---

## Streaming

### Built-in RAG Streaming

```go
// Stream tokens as they are generated (requires full RAG setup)
ch, err := ai.AskStream(ctx, "Tell me a story")
if err != nil {
    log.Fatal(err)
}

for chunk := range ch {
    fmt.Print(chunk.Content) // print each token as it arrives
    if chunk.FinishReason == "stop" {
        break
    }
}
fmt.Println()
```

### Direct Streaming Access

For providers that support streaming, access the underlying `StreamCompleter` directly:

```go
if streamer := ai.StreamLLM(); streamer != nil {
    ch, err := streamer.ChatStream(ctx, []core.Message{
        {Role: "user", Content: "Tell me a story"},
    })
    if err != nil {
        log.Fatal(err)
    }
    for chunk := range ch {
        fmt.Print(chunk.Content)
    }
    fmt.Println()
}
```

`StreamLLM()` returns `nil` if the configured LLM provider does not implement
the `StreamCompleter` interface.

---

## Conversations with Memory

```go
import "github.com/bachtiarpanjaitan/ihandai-go/pkg/memory"

// Setup — only LLM + memory needed; no embedding or vector store required
store := memory.NewInMemoryStore()
ai, _ := ihandai.New(
    ihandai.WithLLM("ollama", llm.WithModel("llama3")),
    ihandai.WithMemory(store),
)
defer ai.Close()

// Multi-turn conversation (gracefully skips retrieval if RAG is not configured)
resp1, _ := ai.AskConversation(ctx, "user-123", "My name is Alice.")
resp2, _ := ai.AskConversation(ctx, "user-123", "What is my name?")
// → "Your name is Alice."

// If you add RAG later (embedding + vector store), the same calls
// automatically enrich responses with retrieved context — no code changes needed.

// Conversation history
history, _ := store.History(ctx, "user-123")
for _, msg := range history {
    fmt.Printf("[%s] %s\n", msg.Role, msg.Content)
}
```

### Window Management

```go
// Auto-trim when approaching token limits
wm := memory.NewWindowManager(store, "llama3", 4096, tokenCounter)
msgs, err := wm.Fit(ctx, "user-123", newMessage)
// msgs contains the trimmed conversation that fits within 4096 tokens
```

---

## Agents

### Basic Agent

```go
import (
    "github.com/bachtiarpanjaitan/ihandai-go/pkg/agent"
    agenttools "github.com/bachtiarpanjaitan/ihandai-go/pkg/agent/tools"
)

ai, _ := ihandai.New(
    ihandai.WithLLM("ollama", llm.WithModel("llama3")),
    ihandai.WithTools(
        agenttools.NewCalculator(),
        agenttools.NewHTTPRequest(),
    ),
)

result, err := ai.Run(ctx, "What is 15 * 7 + 22?")
fmt.Println(result.Answer)
// → "15 * 7 = 105, 105 + 22 = 127. The answer is 127."

// See agent's reasoning steps
for _, step := range result.Steps {
    fmt.Printf("Thought: %s\n", step.Thought)
    fmt.Printf("Action: %s(%s)\n", step.Action.Name, step.Action.Input)
    fmt.Printf("Observation: %s\n\n", step.Observation)
}
```

### Custom Tool

```go
type WeatherTool struct{}

func (w WeatherTool) Name() string        { return "weather" }
func (w WeatherTool) Description() string  { return "Get weather for a city" }
func (w WeatherTool) InputSchema() *core.JSONSchema {
    return &core.JSONSchema{
        Type: "object",
        Properties: map[string]*core.JSONSchemaProp{
            "city": {Type: "string", Description: "City name"},
        },
        Required: []string{"city"},
    }
}
func (w WeatherTool) Execute(ctx context.Context, input json.RawMessage) (json.RawMessage, error) {
    var params struct{ City string `json:"city"` }
    json.Unmarshal(input, &params)
    // Call weather API...
    return json.RawMessage(`{"temp":72,"condition":"sunny"}`), nil
}

ai.SetTools(WeatherTool{})
result, _ := ai.Run(ctx, "What is the weather in Jakarta?")
```

---

## Workflows

### Linear Pipeline

```go
import "github.com/bachtiarpanjaitan/ihandai-go/pkg/workflow"

w := workflow.Build("data-pipeline").
    Add("fetch", func(ctx context.Context, inputs map[string]any) (map[string]any, error) {
        // Fetch data from API
        return map[string]any{"data": "raw data"}, nil
    }).
    Add("transform", func(ctx context.Context, inputs map[string]any) (map[string]any, error) {
        data := inputs["data"].(string)
        return map[string]any{"transformed": strings.ToUpper(data)}, nil
    }, "fetch").
    Add("store", func(ctx context.Context, inputs map[string]any) (map[string]any, error) {
        transformed := inputs["transformed"].(string)
        // Store in database
        return map[string]any{"stored": true}, nil
    }, "transform").
    MustWorkflow()

result, _ := w.Run(ctx, nil)
```

### Parallel Execution

```go
w := workflow.Build("parallel-analysis").
    Add("load", loadData).
    Add("sentiment", analyzeSentiment, "load").  // runs in parallel
    Add("entities", extractEntities, "load").    // runs in parallel
    Add("keywords", extractKeywords, "load").    // runs in parallel
    Add("merge", mergeResults, "sentiment", "entities", "keywords").
    MustWorkflow()

result, _ := w.Run(ctx, map[string]any{"source": "data.csv"})
```

### Conditional Branching

```go
w := workflow.New("conditional-example")

w.AddStep(&workflow.Step{Name: "check", Run: checkData})
w.AddStep(&workflow.Step{
    Name: "process",
    Condition: func(ctx context.Context, inputs map[string]any) (map[string]any, error) {
        if valid, _ := inputs["valid"].(bool); !valid {
            return nil, nil // skip this branch
        }
        return map[string]any{"proceed": true}, nil
    },
    Run:       processData,
    DependsOn: []string{"check"},
})
```

### Save & Load

```go
// Save workflow structure
w.Save("pipeline.json")

// Load and re-attach functions
loaded, _ := workflow.Load("pipeline.json")
loaded.SetFunc("fetch", fetchData)
loaded.SetFunc("transform", transformData)
loaded.SetFunc("store", storeData)

loaded.Run(ctx, nil)
```

---

## MCP (Model Context Protocol)

### Connect to Filesystem Server

```go
import "github.com/bachtiarpanjaitan/ihandai-go/pkg/mcp"

// In-process filesystem server
srv := mcp.NewFilesystemServer("./project/")
transport := mcp.NewInMemoryTransport(srv)
client, _ := mcp.Connect(transport)
defer client.Close()

// List files
resources, _ := client.ListResources(ctx)
for _, r := range resources {
    fmt.Printf("%s — %s\n", r.Name, r.Description)
}

// Read a file
result, _ := client.ReadResource(ctx, "file://README.md")
fmt.Println(result.Contents[0].Text)

// Attach to ihandai client
ai.SetMCP(client)
allResources := ai.MCPResources()
```

### Connect to External MCP Server

```go
// Connect via stdio (subprocess)
client, err := mcp.ConnectStdio("npx", "@modelcontextprotocol/server-filesystem", "/path/to/root")
```

---

## Error Handling

```go
resp, err := ai.Ask(ctx, "query")
if err != nil {
    var rateLimitErr *ihandai.RateLimitError
    var authErr *ihandai.AuthError
    var timeoutErr *ihandai.TimeoutError
    var providerErr *ihandai.ProviderError
    var pipelineErr *ihandai.PipelineError

    switch {
    case errors.As(err, &rateLimitErr):
        fmt.Printf("Rate limited by %s. Retry after %v\n",
            rateLimitErr.Provider, rateLimitErr.RetryAfter)
        time.Sleep(rateLimitErr.RetryAfter)
        // retry

    case errors.As(err, &authErr):
        log.Fatalf("Auth failed for %s. Check API key.", authErr.Provider)

    case errors.As(err, &timeoutErr):
        log.Printf("Timeout after %v with %s", timeoutErr.Duration, timeoutErr.Provider)

    case errors.As(err, &providerErr):
        log.Printf("Provider %s returned HTTP %d: %s",
            providerErr.Provider, providerErr.StatusCode, providerErr.Body)

    case errors.As(err, &pipelineErr):
        log.Printf("Pipeline step %q failed: %v", pipelineErr.Step, pipelineErr.Err)

    default:
        log.Printf("Unexpected error: %v", err)
    }
}
```

---

## Production Features

### Rate Limiting

```go
import "github.com/bachtiarpanjaitan/ihandai-go/pkg/telemetry"

limiter := telemetry.NewRateLimiter(10, 5) // 10 req/sec, burst 5

for _, query := range queries {
    if !limiter.Allow() {
        limiter.Wait(5 * time.Second) // wait up to 5s
    }
    ai.Ask(ctx, query)
}
```

### Circuit Breaker

```go
breaker := telemetry.NewCircuitBreaker(5, 30*time.Second) // 5 failures, 30s reset

if breaker.Allow() {
    resp, err := ai.Ask(ctx, query)
    if err != nil {
        breaker.Failure()
        return
    }
    breaker.Success()
} else {
    return fmt.Errorf("LLM service unavailable (circuit open)")
}
```

### Tracing

```go
tracer := telemetry.NewTracer(slog.Default())

ctx, span := tracer.Start(ctx, "rag-query",
    slog.String("query", query),
    slog.Int("top_k", 5),
)
defer span.End(nil)

resp, err := ai.Ask(ctx, query)
if err != nil {
    span.End(err)
}
```

---

## Testing Your Integration

```go
import (
    "github.com/bachtiarpanjaitan/ihandai-go/pkg/llm"
    "github.com/bachtiarpanjaitan/ihandai-go/pkg/embedding"
)

// Register mock providers in your tests
func init() {
    llm.Register("mock", func(cfg llm.Config) (llm.ChatCompleter, error) {
        return &myMockLLM{}, nil
    })
    embedding.Register("mock", func(cfg embedding.Config) (embedding.Embedder, error) {
        return &myMockEmbedding{}, nil
    })
}

func TestMyApp(t *testing.T) {
    ai, _ := ihandai.New(
        ihandai.WithLLM("mock"),
        ihandai.WithEmbedding("mock"),
    )
    resp, _ := ai.Ask(context.Background(), "test")
    assert.Equal(t, "expected response", resp.Content)
}
```

---

## Context & Timeout

```go
// All methods accept context for cancellation and timeouts
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()

resp, err := ai.Ask(ctx, "complex query")

// Cancel mid-request
ctx, cancel = context.WithCancel(context.Background())
go func() {
    time.Sleep(5 * time.Second)
    cancel() // cancel if taking too long
}()
resp, err = ai.Ask(ctx, "query")
```

---

## Concurrency

```go
// Client is safe for concurrent use
ai, _ := ihandai.New(...)

var wg sync.WaitGroup
for _, query := range queries {
    wg.Add(1)
    go func(q string) {
        defer wg.Done()
        resp, _ := ai.Ask(context.Background(), q)
        fmt.Println(resp.Content)
    }(query)
}
wg.Wait()
```

---

## Complete Example: RAG Chatbot

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
    ai, err := ihandai.New(
        ihandai.WithLLM("ollama",
            llm.WithModel("llama3"),
            llm.WithBaseURL("http://localhost:11434"),
        ),
        ihandai.WithEmbedding("ollama",
            embedding.WithModel("nomic-embed-text"),
        ),
        ihandai.WithVectorStore("mock"),
        ihandai.WithMemory(memory.NewInMemoryStore()),
    )
    if err != nil {
        log.Fatal(err)
    }
    defer ai.Close()

    // Index documents
    if len(os.Args) > 1 {
        fmt.Println("Indexing documents...")
        if err := ai.Index(context.Background(), os.Args[1]); err != nil {
            log.Printf("Index warning: %v", err)
        }
    }

    // Interactive chat
    sessionID := "default"
    scanner := bufio.NewScanner(os.Stdin)
    fmt.Println("Chat started. Type /exit to quit.")

    for {
        fmt.Print("\n> ")
        if !scanner.Scan() {
            break
        }
        input := scanner.Text()
        if input == "/exit" {
            break
        }

        resp, err := ai.AskConversation(context.Background(), sessionID, input,
            ihandai.WithTopK(3),
        )
        if err != nil {
            fmt.Printf("Error: %v\n", err)
            continue
        }
        fmt.Println(resp.Content)
    }
}
```

---

## Next Steps

- Add more providers: `openai`, `gemini`, `claude`, `qdrant`, `pgvector`
- Implement custom tools for your domain
- Build complex workflows for multi-step AI tasks
- Use MCP to connect external data sources
- Add production telemetry for monitoring
