package llm

import (
	"context"
	"errors"
	"testing"

	"github.com/bachtiarpanjaitan/ihandai-go/pkg/core"
)

// Compile-time interface satisfaction check
var _ ChatCompleter = (*mockChat)(nil)
var _ StreamCompleter = (*mockStream)(nil)
var _ TokenCounter = (*mockCounter)(nil)

type mockChat struct{}

func (m mockChat) Chat(ctx context.Context, msgs []core.Message) (*core.Response, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	return &core.Response{Content: "mock"}, nil
}

type mockStream struct{}

func (m mockStream) ChatStream(ctx context.Context, msgs []core.Message) (<-chan Chunk, error) {
	ch := make(chan Chunk)
	close(ch)
	return ch, nil
}

type mockCounter struct{}

func (m mockCounter) CountTokens(ctx context.Context, model, text string) (int, error) {
	return len(text), nil
}

func TestChatCompleter_Mock(t *testing.T) {
	var c ChatCompleter = mockChat{}
	resp, err := c.Chat(context.Background(), []core.Message{{Role: "user", Content: "hello"}})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Content != "mock" {
		t.Errorf("got %q, want %q", resp.Content, "mock")
	}
}

func TestStreamCompleter_Mock(t *testing.T) {
	var s StreamCompleter = mockStream{}
	ch, err := s.ChatStream(context.Background(), nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	_, ok := <-ch
	if ok {
		t.Error("channel should be closed immediately")
	}
}

func TestTokenCounter_Mock(t *testing.T) {
	var tc TokenCounter = mockCounter{}
	count, err := tc.CountTokens(context.Background(), "gpt-4", "hello world")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if count != 11 {
		t.Errorf("got %d, want 11", count)
	}
}

func TestChatCompleter_ContextCancelled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err := (&mockChat{}).Chat(ctx, nil)
	if !errors.Is(err, context.Canceled) {
		t.Errorf("expected context.Canceled, got %v", err)
	}
}

func TestConfig_Options(t *testing.T) {
	cfg := Config{}
	WithModel("gpt-4o")(&cfg)
	WithAPIKey("sk-123")(&cfg)
	WithBaseURL("https://api.example.com")(&cfg)

	if cfg.Model != "gpt-4o" {
		t.Errorf("Model: got %q, want %q", cfg.Model, "gpt-4o")
	}
	if cfg.APIKey != "sk-123" {
		t.Errorf("APIKey: got %q, want %q", cfg.APIKey, "sk-123")
	}
	if cfg.BaseURL != "https://api.example.com" {
		t.Errorf("BaseURL: got %q, want %q", cfg.BaseURL, "https://api.example.com")
	}
}

func TestChunk(t *testing.T) {
	c := Chunk{Content: "hello", FinishReason: "stop"}
	if c.Content != "hello" {
		t.Errorf("Content: got %q, want %q", c.Content, "hello")
	}
	if c.FinishReason != "stop" {
		t.Errorf("FinishReason: got %q, want %q", c.FinishReason, "stop")
	}
}
