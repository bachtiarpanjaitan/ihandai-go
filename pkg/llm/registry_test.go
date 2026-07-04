package llm

import (
	"context"
	"testing"

	"github.com/bachtiarpanjaitan/ihandai-go/pkg/core"
)

func TestRegister(t *testing.T) {
	// Use a unique name to avoid conflicts with other tests
	name := "test-register-llm"
	Register(name, func(cfg Config) (ChatCompleter, error) {
		return mockChat{}, nil
	})

	if !contains(Registered(), name) {
		t.Errorf("expected %q to be registered", name)
	}
}

func TestRegister_EmptyName(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic for empty name")
		}
	}()
	Register("", func(cfg Config) (ChatCompleter, error) { return nil, nil })
}

func TestRegister_NilFactory(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic for nil factory")
		}
	}()
	Register("nil-factory", nil)
}

func TestRegister_Duplicate(t *testing.T) {
	name := "test-dup-llm"
	f := func(cfg Config) (ChatCompleter, error) { return mockChat{}, nil }
	Register(name, f)

	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic for duplicate registration")
		}
	}()
	Register(name, f)
}

func TestOpen_Valid(t *testing.T) {
	name := "test-open-valid-llm"
	Register(name, func(cfg Config) (ChatCompleter, error) {
		return mockChat{}, nil
	})

	c, err := Open(name)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if c == nil {
		t.Fatal("Open returned nil ChatCompleter")
	}

	// Verify it works
	resp, err := c.Chat(context.Background(), []core.Message{{Role: "user", Content: "test"}})
	if err != nil {
		t.Fatalf("Chat failed: %v", err)
	}
	if resp.Content != "mock" {
		t.Errorf("got %q, want %q", resp.Content, "mock")
	}
}

func TestOpen_Unknown(t *testing.T) {
	_, err := Open("nonexistent-provider-xyz")
	if err == nil {
		t.Error("expected error for unknown provider")
	}
}

func TestOpen_WithOptions(t *testing.T) {
	name := "test-open-opts-llm"
	Register(name, func(cfg Config) (ChatCompleter, error) {
		if cfg.Model != "test-model" {
			t.Errorf("Model: got %q, want %q", cfg.Model, "test-model")
		}
		return mockChat{}, nil
	})

	_, err := Open(name, WithModel("test-model"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRegistered(t *testing.T) {
	// At minimum, the providers registered in other tests should appear
	names := Registered()
	if names == nil {
		t.Error("Registered() should not return nil")
	}
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
