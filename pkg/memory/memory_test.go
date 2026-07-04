package memory

import (
	"context"
	"sync"
	"testing"

	"github.com/bachtiarpanjaitan/ihandai-go/pkg/core"
)

func TestInMemoryStore_Append(t *testing.T) {
	s := NewInMemoryStore()
	ctx := context.Background()
	key := "test-convo"

	err := s.Append(ctx, key, core.Message{Role: "user", Content: "hello"})
	if err != nil {
		t.Fatalf("Append failed: %v", err)
	}
	if s.Len() != 1 {
		t.Errorf("Len: got %d, want 1", s.Len())
	}
}

func TestInMemoryStore_History(t *testing.T) {
	s := NewInMemoryStore()
	ctx := context.Background()
	key := "test-convo"

	s.Append(ctx, key, core.Message{Role: "system", Content: "You are helpful."})
	s.Append(ctx, key, core.Message{Role: "user", Content: "hello"})
	s.Append(ctx, key, core.Message{Role: "assistant", Content: "Hi!"})

	history, err := s.History(ctx, key)
	if err != nil {
		t.Fatalf("History failed: %v", err)
	}
	if len(history) != 3 {
		t.Errorf("got %d messages, want 3", len(history))
	}
	if history[0].Role != "system" {
		t.Errorf("first message role: got %q, want %q", history[0].Role, "system")
	}
	if history[2].Role != "assistant" {
		t.Errorf("last message role: got %q, want %q", history[2].Role, "assistant")
	}
}

func TestInMemoryStore_History_Empty(t *testing.T) {
	s := NewInMemoryStore()
	history, err := s.History(context.Background(), "nonexistent")
	if err != nil {
		t.Fatalf("History failed: %v", err)
	}
	if len(history) != 0 {
		t.Errorf("got %d messages, want 0", len(history))
	}
	if history == nil {
		t.Error("History should return empty slice, not nil")
	}
}

func TestInMemoryStore_History_ReturnsCopy(t *testing.T) {
	s := NewInMemoryStore()
	ctx := context.Background()
	s.Append(ctx, "key", core.Message{Role: "user", Content: "original"})

	history, _ := s.History(ctx, "key")
	history[0].Content = "modified" // mutate the copy

	history2, _ := s.History(ctx, "key")
	if history2[0].Content != "original" {
		t.Error("History should return a copy, mutation should not affect store")
	}
}

func TestInMemoryStore_Trim(t *testing.T) {
	s := NewInMemoryStore()
	ctx := context.Background()
	key := "trim-test"

	// System message should be preserved
	s.Append(ctx, key, core.Message{Role: "system", Content: "You are helpful."})
	s.Append(ctx, key, core.Message{Role: "user", Content: "msg1"})
	s.Append(ctx, key, core.Message{Role: "assistant", Content: "reply1"})
	s.Append(ctx, key, core.Message{Role: "user", Content: "msg2"})
	s.Append(ctx, key, core.Message{Role: "assistant", Content: "reply2"})
	s.Append(ctx, key, core.Message{Role: "user", Content: "msg3"})

	// Trim to ~20 tokens (should keep system + last few messages)
	err := s.Trim(ctx, key, 20, nil)
	if err != nil {
		t.Fatalf("Trim failed: %v", err)
	}

	history, _ := s.History(ctx, key)
	if len(history) == 0 {
		t.Fatal("history is empty after trim")
	}
	if history[0].Role != "system" {
		t.Error("system message should be preserved")
	}
}

func TestInMemoryStore_Clear(t *testing.T) {
	s := NewInMemoryStore()
	ctx := context.Background()
	key := "clear-test"

	s.Append(ctx, key, core.Message{Role: "user", Content: "hello"})
	if s.Len() != 1 {
		t.Fatal("expected 1 conversation")
	}

	err := s.Clear(ctx, key)
	if err != nil {
		t.Fatalf("Clear failed: %v", err)
	}
	if s.Len() != 0 {
		t.Errorf("Len: got %d, want 0 after clear", s.Len())
	}

	history, _ := s.History(ctx, key)
	if len(history) != 0 {
		t.Errorf("got %d messages after clear, want 0", len(history))
	}
}

func TestInMemoryStore_Concurrency(t *testing.T) {
	s := NewInMemoryStore()
	ctx := context.Background()
	key := "concurrent"

	var wg sync.WaitGroup
	n := 100

	for i := range n {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			s.Append(ctx, key, core.Message{Role: "user", Content: "msg " + itoa(i)})
		}(i)
	}

	for range n / 2 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			s.History(ctx, key)
		}()
	}

	wg.Wait()

	history, _ := s.History(ctx, key)
	if len(history) != n {
		t.Errorf("got %d messages after concurrent writes, want %d", len(history), n)
	}
}

func TestWindowManager_Fit(t *testing.T) {
	s := NewInMemoryStore()
	key := "window-test"

	wm := NewWindowManager(s, "gpt-4", 50, nil)

	// Add a large message
	largeMsg := core.Message{Role: "user", Content: "This is a very long message that should consume many tokens"}
	msgs, err := wm.Fit(context.Background(), key, largeMsg)
	if err != nil {
		t.Fatalf("Fit failed: %v", err)
	}
	if len(msgs) != 1 {
		t.Errorf("got %d messages, want 1", len(msgs))
	}

	// Add many more messages to trigger trimming
	for range 10 {
		wm.Fit(context.Background(), key, core.Message{Role: "assistant", Content: "short reply"})
	}

	history, _ := s.History(context.Background(), key)
	t.Logf("History after 11 messages: %d messages", len(history))
	// Some trimming should have occurred
}

func TestEstimateTokens_Fallback(t *testing.T) {
	tokens := estimateTokens(nil, []core.Message{
		{Role: "user", Content: "hello world"}, // 11 chars → ~2 tokens
	}, nil)
	if tokens < 1 {
		t.Error("expected at least 1 token")
	}
}

func itoa(n int) string {
	digits := ""
	for n > 0 {
		digits = string(rune('0'+n%10)) + digits
		n /= 10
	}
	if digits == "" {
		return "0"
	}
	return digits
}
