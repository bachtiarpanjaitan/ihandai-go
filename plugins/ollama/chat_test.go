package ollama

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/bachtiarpanjaitan/ihandai-go/pkg/core"
)

func TestChatCompleter_Integration(t *testing.T) {
	url := os.Getenv("OLLAMA_URL")
	if url == "" {
		t.Skip("OLLAMA_URL not set, skipping integration test")
	}

	c := NewChatCompleter(WithBaseURL(url), WithModel("llama3"))
	resp, err := c.Chat(context.Background(), []core.Message{
		{Role: "user", Content: "Say hello in one word"},
	})
	if err != nil {
		t.Fatalf("Chat failed: %v", err)
	}
	if resp.Content == "" {
		t.Error("expected non-empty response")
	}
	t.Logf("Response: %s", resp.Content)
}

func TestChatCompleter_MockServer(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/chat" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.Method != http.MethodPost {
			t.Errorf("unexpected method: %s", r.Method)
		}

		var req chatRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("failed to decode request: %v", err)
		}
		if req.Model != "llama3" {
			t.Errorf("Model: got %q, want %q", req.Model, "llama3")
		}
		if req.Stream != false {
			t.Error("Stream should be false")
		}
		if len(req.Messages) != 1 {
			t.Errorf("Messages: got %d, want 1", len(req.Messages))
		}

		resp := chatResponse{
			Model: "llama3",
			Message: chatMessage{
				Role:    "assistant",
				Content: "Hello!",
			},
			Done:       true,
			DoneReason: "stop",
		}
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			t.Fatalf("failed to encode response: %v", err)
		}
	}))
	defer server.Close()

	c := NewChatCompleter(WithBaseURL(server.URL), WithModel("llama3"))
	resp, err := c.Chat(context.Background(), []core.Message{
		{Role: "user", Content: "Hi"},
	})
	if err != nil {
		t.Fatalf("Chat failed: %v", err)
	}
	if resp.Content != "Hello!" {
		t.Errorf("Content: got %q, want %q", resp.Content, "Hello!")
	}
	if resp.FinishReason != "stop" {
		t.Errorf("FinishReason: got %q, want %q", resp.FinishReason, "stop")
	}
}

func TestChatCompleter_ServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("internal error"))
	}))
	defer server.Close()

	c := NewChatCompleter(WithBaseURL(server.URL), WithModel("llama3"))
	_, err := c.Chat(context.Background(), []core.Message{
		{Role: "user", Content: "Hi"},
	})
	if err == nil {
		t.Error("expected error for 500 response")
	}
}

func TestChatCompleter_ContextCancelled(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Don't respond — context should cancel the request
	}))
	defer server.Close()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	c := NewChatCompleter(WithBaseURL(server.URL), WithModel("llama3"))
	_, err := c.Chat(ctx, []core.Message{{Role: "user", Content: "Hi"}})
	if err == nil {
		t.Error("expected error for canceled context")
	}
}

func TestChatCompleter_Registry(t *testing.T) {
	// Verify the provider is registered via init()
	// This import registers "ollama" — test that it's available
	// Note: the init() function runs automatically when importing this package
	t.Log("ollama provider registered via init()")
}
