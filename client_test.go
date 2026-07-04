package ihandai

import (
	"context"
	"log/slog"
	"os"
	"sync"
	"testing"

	"github.com/bachtiarpanjaitan/ihandai-go/pkg/core"
	"github.com/bachtiarpanjaitan/ihandai-go/pkg/embedding"
	"github.com/bachtiarpanjaitan/ihandai-go/pkg/llm"
	"github.com/bachtiarpanjaitan/ihandai-go/pkg/memory"
	"github.com/bachtiarpanjaitan/ihandai-go/pkg/retriever"
	"github.com/bachtiarpanjaitan/ihandai-go/pkg/vectordb"
)

// ---------------------------------------------------------------------------
// Mock providers (registered via init)
// ---------------------------------------------------------------------------

func init() {
	llm.Register("mock", func(cfg llm.Config) (llm.ChatCompleter, error) {
		return &mockLLM{}, nil
	})
	llm.Register("mock-stream", func(cfg llm.Config) (llm.ChatCompleter, error) {
		return &mockStreamLLM{}, nil
	})
	embedding.Register("mock", func(cfg embedding.Config) (embedding.Embedder, error) {
		return &mockEmbedding{}, nil
	})
	vectordb.Register("mock", func(cfg vectordb.Config) (vectordb.VectorSearcher, error) {
		return &mockStore{docs: make(map[string]core.Document)}, nil
	})
}

// --- LLM mocks ---

type mockLLM struct{}

func (m *mockLLM) Chat(ctx context.Context, msgs []core.Message) (*core.Response, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	return &core.Response{Content: "mock response"}, nil
}

type mockStreamLLM struct{}

func (m *mockStreamLLM) Chat(ctx context.Context, msgs []core.Message) (*core.Response, error) {
	return &core.Response{Content: "mock"}, nil
}

func (m *mockStreamLLM) ChatStream(ctx context.Context, msgs []core.Message) (<-chan llm.Chunk, error) {
	ch := make(chan llm.Chunk, 1)
	go func() {
		ch <- llm.Chunk{Content: "mock stream"}
		close(ch)
	}()
	return ch, nil
}

// --- Embedding mock ---

type mockEmbedding struct{}

func (m *mockEmbedding) Embed(ctx context.Context, text string) ([]float64, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	return []float64{0.1, 0.2, 0.3}, nil
}

func (m *mockEmbedding) EmbedBatch(ctx context.Context, texts []string) ([][]float64, error) {
	result := make([][]float64, len(texts))
	for i := range texts {
		result[i] = []float64{0.1, 0.2, 0.3}
	}
	return result, nil
}

// --- VectorStore mock ---

type mockStore struct {
	docs map[string]core.Document
}

func (m *mockStore) Search(ctx context.Context, vector []float64, opts ...vectordb.SearchOption) ([]core.ScoredDocument, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	var results []core.ScoredDocument
	for _, doc := range m.docs {
		results = append(results, core.ScoredDocument{
			Document: doc,
			Score:    0.95,
		})
	}
	if len(results) > 2 {
		results = results[:2]
	}
	return results, nil
}

func (m *mockStore) Insert(ctx context.Context, documents []core.Document) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	for _, doc := range documents {
		m.docs[doc.ID] = doc
	}
	return nil
}

func (m *mockStore) Delete(ctx context.Context, ids []string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	for _, id := range ids {
		delete(m.docs, id)
	}
	return nil
}

// ---------------------------------------------------------------------------
// Tests
// ---------------------------------------------------------------------------

func TestNew_Defaults(t *testing.T) {
	ai, err := New()
	if err != nil {
		t.Fatalf("New() unexpected error: %v", err)
	}
	defer ai.Close()

	if ai.LLM() != nil {
		t.Error("LLM() should return nil when not configured")
	}
	if ai.Embedding() != nil {
		t.Error("Embedding() should return nil when not configured")
	}
	if ai.VectorStore() != nil {
		t.Error("VectorStore() should return nil when not configured")
	}
}

func TestNew_WithLogger(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))
	ai, _ := New(WithLogger(logger))
	defer ai.Close()
	if ai.cfg.Logger != logger {
		t.Error("WithLogger did not set the logger")
	}
}

func TestNew_UnknownProvider(t *testing.T) {
	_, err := New(WithLLM("nonexistent-xyz"))
	if err == nil {
		t.Error("expected error for unknown provider")
	}
}

func TestNew_WithLLMConfig(t *testing.T) {
	ai, err := New(
		WithLLM("mock", llm.WithModel("test-model")),
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer ai.Close()
	if ai.LLM() == nil {
		t.Error("LLM() should not be nil")
	}
}

func TestClient_CloseIdempotent(t *testing.T) {
	ai, _ := New()
	for i := range 3 {
		if err := ai.Close(); err != nil {
			t.Errorf("Close #%d: %v", i, err)
		}
	}
}

func TestClient_Concurrency(t *testing.T) {
	ai, _ := New()
	defer ai.Close()

	var wg sync.WaitGroup
	count := 50

	wg.Add(count * 3)
	for range count {
		go func() { defer wg.Done(); _ = ai.Close() }()
		go func() { defer wg.Done(); _ = ai.LLM(); _ = ai.Embedding(); _ = ai.VectorStore() }()
		go func() { defer wg.Done(); ai.SetReranker(nil); ai.SetPromptBuilder(nil) }()
	}
	wg.Wait()
}

func TestAsk_NoProviders(t *testing.T) {
	ai, _ := New()
	defer ai.Close()
	_, err := ai.Ask(context.TODO(), "test")
	if err == nil {
		t.Error("expected error without providers")
	}
}

func TestAsk_FullPipeline(t *testing.T) {
	ai, err := New(
		WithLLM("mock"),
		WithEmbedding("mock"),
		WithVectorStore("mock"),
	)
	if err != nil {
		t.Fatalf("New() unexpected error: %v", err)
	}
	defer ai.Close()

	// Index a document first
	err = ai.Index(context.Background(), "../../README.md")
	if err != nil {
		t.Logf("Index warning (expected if file doesn't exist): %v", err)
	}

	resp, err := ai.Ask(context.Background(), "test query")
	if err != nil {
		t.Fatalf("Ask failed: %v", err)
	}
	if resp.Content != "mock response" {
		t.Errorf("Content: got %q, want %q", resp.Content, "mock response")
	}
}

func TestAsk_ContextCancelled(t *testing.T) {
	ai, _ := New(
		WithLLM("mock"),
		WithEmbedding("mock"),
		WithVectorStore("mock"),
	)
	defer ai.Close()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := ai.Ask(ctx, "test")
	if err == nil {
		t.Error("expected error for cancelled context")
	}
}

func TestAskStream(t *testing.T) {
	ai, _ := New(
		WithLLM("mock-stream"),
		WithEmbedding("mock"),
		WithVectorStore("mock"),
	)
	defer ai.Close()

	ch, err := ai.AskStream(context.Background(), "test")
	if err != nil {
		t.Fatalf("AskStream failed: %v", err)
	}
	chunk := <-ch
	if chunk.Content != "mock stream" {
		t.Errorf("got %q, want %q", chunk.Content, "mock stream")
	}
}

func TestAskStream_NoStreamSupport(t *testing.T) {
	ai, _ := New(
		WithLLM("mock"), // mock doesn't implement StreamCompleter
		WithEmbedding("mock"),
		WithVectorStore("mock"),
	)
	defer ai.Close()

	_, err := ai.AskStream(context.Background(), "test")
	if err == nil {
		t.Error("expected error for non-streaming LLM")
	}
}

func TestIndex_NoProviders(t *testing.T) {
	ai, _ := New()
	defer ai.Close()
	err := ai.Index(context.TODO(), "source")
	if err == nil {
		t.Error("expected error without providers")
	}
}

func TestIndex_Pipeline(t *testing.T) {
	// Create a temp file for indexing
	tmpDir := t.TempDir()
	tmpFile := tmpDir + "/test.txt"
	if err := os.WriteFile(tmpFile, []byte("Hello world. This is a test document."), 0644); err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}

	ai, _ := New(
		WithEmbedding("mock"),
		WithVectorStore("mock"),
	)
	defer ai.Close()

	err := ai.Index(context.Background(), tmpFile)
	if err != nil {
		t.Fatalf("Index failed: %v", err)
	}
}

func TestPipelineError_Wrapping(t *testing.T) {
	ai, _ := New(
		WithLLM("mock"),
		WithEmbedding("mock"),
		WithVectorStore("mock"),
	)
	defer ai.Close()

	_, err := ai.Ask(context.Background(), "test")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	pe := &PipelineError{Step: "chat", Err: context.Canceled}
	if pe.Error() == "" {
		t.Error("PipelineError should have a non-empty message")
	}
}

func TestAsk_WithTopK(t *testing.T) {
	ai, _ := New(
		WithLLM("mock"),
		WithEmbedding("mock"),
		WithVectorStore("mock"),
	)
	defer ai.Close()

	resp, err := ai.Ask(context.Background(), "test",
		WithTopK(3),
	)
	if err != nil {
		t.Fatalf("Ask with TopK failed: %v", err)
	}
	if resp.Content != "mock response" {
		t.Errorf("got %q, want %q", resp.Content, "mock response")
	}
}

func TestAsk_WithFilter(t *testing.T) {
	ai, _ := New(
		WithLLM("mock"),
		WithEmbedding("mock"),
		WithVectorStore("mock"),
	)
	defer ai.Close()

	resp, err := ai.Ask(context.Background(), "test",
		WithFilter(map[string]any{"source": "docs/"}),
	)
	if err != nil {
		t.Fatalf("Ask with filter failed: %v", err)
	}
	_ = resp
}

func TestAsk_WithCustomRetriever(t *testing.T) {
	ai, _ := New(
		WithLLM("mock"),
		WithEmbedding("mock"),
		WithVectorStore("mock"),
	)
	defer ai.Close()

	// Use MMR retriever
	mmr := retriever.NewMMR(ai.VectorStore(), 0.7)
	resp, err := ai.Ask(context.Background(), "test",
		WithRetriever(mmr),
		WithTopK(3),
	)
	if err != nil {
		t.Fatalf("Ask with custom retriever failed: %v", err)
	}
	if resp.Content != "mock response" {
		t.Errorf("got %q, want %q", resp.Content, "mock response")
	}
}

func TestIndex_WithCustomLoader(t *testing.T) {
	ai, _ := New(
		WithEmbedding("mock"),
		WithVectorStore("mock"),
	)
	defer ai.Close()

	// Use custom loader
	customLoader := &mockLoader{docs: []core.Document{
		{ID: "custom", Content: "custom doc"},
	}}
	err := ai.Index(context.Background(), "test-source",
		WithLoader(customLoader),
	)
	if err != nil {
		t.Fatalf("Index with custom loader failed: %v", err)
	}
}

func TestAsk_Defaults(t *testing.T) {
	ai, _ := New(
		WithLLM("mock"),
		WithEmbedding("mock"),
		WithVectorStore("mock"),
	)
	defer ai.Close()

	// Verify defaults work
	resp, err := ai.Ask(context.Background(), "query")
	if err != nil {
		t.Fatalf("Ask with defaults failed: %v", err)
	}
	if resp.Content == "" {
		t.Error("expected non-empty response")
	}
}

// mockLoader for testing Index with custom loader
type mockLoader struct {
	docs []core.Document
}

func (m *mockLoader) Load(ctx context.Context, source string) ([]core.Document, error) {
	return m.docs, nil
}

func TestAskConversation(t *testing.T) {
	store := memory.NewInMemoryStore()
	key := "test-session"

	ai, _ := New(
		WithLLM("mock"),
		WithEmbedding("mock"),
		WithVectorStore("mock"),
		WithMemory(store),
	)
	defer ai.Close()

	// First message
	resp, err := ai.AskConversation(context.Background(), key, "hello")
	if err != nil {
		t.Fatalf("AskConversation failed: %v", err)
	}
	if resp.Content == "" {
		t.Error("expected non-empty response")
	}

	// Verify history was saved
	history, _ := store.History(context.Background(), key)
	if len(history) < 2 {
		t.Errorf("expected at least 2 messages in history (query + response), got %d", len(history))
	}

	// Second message should include history
	resp2, err := ai.AskConversation(context.Background(), key, "follow up")
	if err != nil {
		t.Fatalf("second AskConversation failed: %v", err)
	}
	if resp2.Content == "" {
		t.Error("expected non-empty response")
	}
}

func TestAskConversation_NoMemory(t *testing.T) {
	ai, _ := New(
		WithLLM("mock"),
		WithEmbedding("mock"),
		WithVectorStore("mock"),
	)
	defer ai.Close()

	// Should fall back to stateless Ask
	resp, err := ai.AskConversation(context.Background(), "key", "query")
	if err != nil {
		t.Fatalf("AskConversation without memory should fallback to Ask: %v", err)
	}
	if resp.Content == "" {
		t.Error("expected non-empty response")
	}
}

func TestWithMemory(t *testing.T) {
	store := memory.NewInMemoryStore()
	ai, _ := New(WithMemory(store))
	defer ai.Close()

	if ai.Memory() != store {
		t.Error("Memory() should return the configured store")
	}
}
