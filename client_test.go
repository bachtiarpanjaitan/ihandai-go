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
	"github.com/bachtiarpanjaitan/ihandai-go/pkg/vectordb"
)

// Register mock providers for testing.
func init() {
	llm.Register("mock", func(cfg llm.Config) (llm.ChatCompleter, error) {
		return &mockLLM{}, nil
	})
	embedding.Register("mock", func(cfg embedding.Config) (embedding.Embedder, error) {
		return &mockEmbedding{}, nil
	})
	vectordb.Register("mock", func(cfg vectordb.Config) (vectordb.VectorSearcher, error) {
		return &mockVectorStore{}, nil
	})
}

type mockLLM struct{}

func (m *mockLLM) Chat(ctx context.Context, msgs []core.Message) (*core.Response, error) {
	return &core.Response{Content: "mock"}, nil
}

type mockEmbedding struct{}

func (m *mockEmbedding) Embed(ctx context.Context, text string) ([]float64, error) {
	return []float64{0.1}, nil
}

func (m *mockEmbedding) EmbedBatch(ctx context.Context, texts []string) ([][]float64, error) {
	result := make([][]float64, len(texts))
	for i := range texts {
		result[i] = []float64{0.1}
	}
	return result, nil
}

type mockVectorStore struct{}

func (m *mockVectorStore) Search(ctx context.Context, vector []float64, opts ...vectordb.SearchOption) ([]core.ScoredDocument, error) {
	return nil, nil
}

func TestNew_Defaults(t *testing.T) {
	ai, err := New()
	if err != nil {
		t.Fatalf("New() unexpected error: %v", err)
	}
	if ai == nil {
		t.Fatal("New() returned nil Client")
	}
	if ai.cfg == nil {
		t.Fatal("Client has nil config")
	}
	if ai.cfg.Logger == nil {
		t.Fatal("Client has nil logger (should default to slog.Default())")
	}
	if ai.LLM() != nil {
		t.Error("LLM() should return nil when not configured")
	}
	if ai.Embedding() != nil {
		t.Error("Embedding() should return nil when not configured")
	}
	if ai.VectorStore() != nil {
		t.Error("VectorStore() should return nil when not configured")
	}

	ai.Close()
}

func TestNew_WithLogger(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))
	ai, err := New(WithLogger(logger))
	if err != nil {
		t.Fatalf("New(WithLogger(...)) unexpected error: %v", err)
	}
	if ai.cfg.Logger != logger {
		t.Error("WithLogger did not set the logger")
	}
	ai.Close()
}

func TestNew_MultipleOptions(t *testing.T) {
	logger1 := slog.New(slog.NewTextHandler(os.Stderr, nil))
	logger2 := slog.New(slog.NewJSONHandler(os.Stderr, nil))

	ai, err := New(WithLogger(logger1), WithLogger(logger2))
	if err != nil {
		t.Fatalf("New() unexpected error: %v", err)
	}
	if ai.cfg.Logger != logger2 {
		t.Error("last WithLogger should win")
	}
	ai.Close()
}

func TestNew_UnknownProvider(t *testing.T) {
	_, err := New(
		WithLLM("nonexistent-provider-xyz"),
	)
	if err == nil {
		t.Error("expected error for unknown LLM provider")
	}
}

func TestNew_WithLLMConfig(t *testing.T) {
	ai, err := New(
		WithLLM("mock", llm.WithModel("test-model"), llm.WithBaseURL("http://test")),
	)
	if err != nil {
		t.Fatalf("New() unexpected error: %v", err)
	}
	defer ai.Close()

	if ai.cfg.LLMProvider != "mock" {
		t.Errorf("LLMProvider: got %q, want %q", ai.cfg.LLMProvider, "mock")
	}
	if len(ai.cfg.LLMOptions) != 2 {
		t.Errorf("LLMOptions: got %d, want 2", len(ai.cfg.LLMOptions))
	}
}

func TestNew_WithEmbeddingConfig(t *testing.T) {
	ai, err := New(
		WithEmbedding("mock"),
	)
	if err != nil {
		t.Fatalf("New() unexpected error: %v", err)
	}
	defer ai.Close()

	if ai.cfg.EmbeddingProvider != "mock" {
		t.Errorf("EmbeddingProvider: got %q, want %q", ai.cfg.EmbeddingProvider, "mock")
	}
}

func TestNew_WithIndexEmbeddingConfig(t *testing.T) {
	ai, err := New(
		WithIndexEmbedding("mock", nil...),
	)
	if err != nil {
		t.Fatalf("New() unexpected error: %v", err)
	}
	defer ai.Close()

	if ai.cfg.IndexEmbeddingProvider != "mock" {
		t.Errorf("IndexEmbeddingProvider: got %q, want %q", ai.cfg.IndexEmbeddingProvider, "mock")
	}
}

func TestNew_WithVectorStoreConfig(t *testing.T) {
	ai, err := New(
		WithVectorStore("mock"),
	)
	if err != nil {
		t.Fatalf("New() unexpected error: %v", err)
	}
	defer ai.Close()

	if ai.cfg.VectorStoreProvider != "mock" {
		t.Errorf("VectorStoreProvider: got %q, want %q", ai.cfg.VectorStoreProvider, "mock")
	}
}

func TestClient_Config(t *testing.T) {
	ai, _ := New()
	cfg := ai.Config()
	if cfg == nil {
		t.Fatal("Config() returned nil")
	}
	cfg.Logger = nil
	if ai.cfg.Logger == nil {
		t.Error("Config() should return a copy, not the original")
	}
	ai.Close()
}

func TestClient_Close_IsIdempotent(t *testing.T) {
	ai, _ := New()
	if err := ai.Close(); err != nil {
		t.Errorf("first Close() should not error: %v", err)
	}
	if err := ai.Close(); err != nil {
		t.Errorf("second Close() should not error: %v", err)
	}
}

func TestClient_Concurrency(t *testing.T) {
	ai, _ := New()
	defer ai.Close()

	var wg sync.WaitGroup
	count := 100

	wg.Add(count)
	for range count {
		go func() {
			defer wg.Done()
			_ = ai.Close()
		}()
	}

	wg.Add(count)
	for range count {
		go func() {
			defer wg.Done()
			_ = ai.Config()
		}()
	}

	wg.Add(count)
	for range count {
		go func() {
			defer wg.Done()
			_ = ai.LLM()
			_ = ai.Embedding()
			_ = ai.VectorStore()
		}()
	}

	wg.Wait()
}

func TestClient_Ask_WithoutProviders(t *testing.T) {
	ai, _ := New()
	defer ai.Close()

	_, err := ai.Ask(context.TODO(), "test")
	if err == nil {
		t.Error("expected error when LLM is not configured")
	}
}

func TestClient_Index_WithoutProviders(t *testing.T) {
	ai, _ := New()
	defer ai.Close()

	err := ai.Index(context.TODO(), []Document{{ID: "1", Content: "test"}})
	if err == nil {
		t.Error("expected error when embedding is not configured")
	}
}
