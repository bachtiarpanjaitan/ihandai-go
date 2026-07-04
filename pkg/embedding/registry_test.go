package embedding

import (
	"context"
	"testing"
)

func TestRegister_Valid(t *testing.T) {
	name := "test-register-embed"
	Register(name, func(cfg Config) (Embedder, error) {
		return mockEmbedder{}, nil
	})
	if !containsStr(Registered(), name) {
		t.Errorf("expected %q to be registered", name)
	}
}

func TestRegister_EmptyName(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic for empty name")
		}
	}()
	Register("", func(cfg Config) (Embedder, error) { return nil, nil })
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
	name := "test-dup-embed"
	f := func(cfg Config) (Embedder, error) { return mockEmbedder{}, nil }
	Register(name, f)

	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic for duplicate registration")
		}
	}()
	Register(name, f)
}

func TestOpen_Valid(t *testing.T) {
	name := "test-open-valid-embed"
	Register(name, func(cfg Config) (Embedder, error) {
		return mockEmbedder{}, nil
	})

	e, err := Open(name)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if e == nil {
		t.Fatal("Open returned nil Embedder")
	}

	vec, err := e.Embed(context.Background(), "test")
	if err != nil {
		t.Fatalf("Embed failed: %v", err)
	}
	if len(vec) != 3 {
		t.Errorf("got %d dimensions, want 3", len(vec))
	}
}

func TestOpen_Unknown(t *testing.T) {
	_, err := Open("nonexistent-embedder-xyz")
	if err == nil {
		t.Error("expected error for unknown provider")
	}
}

func TestOpen_WithOptions(t *testing.T) {
	name := "test-open-opts-embed"
	Register(name, func(cfg Config) (Embedder, error) {
		if cfg.Model != "test-model" {
			t.Errorf("Model: got %q, want %q", cfg.Model, "test-model")
		}
		return mockEmbedder{}, nil
	})

	_, err := Open(name, WithModel("test-model"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRegistered_NotNil(t *testing.T) {
	if Registered() == nil {
		t.Error("Registered() should not return nil")
	}
}

func containsStr(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
