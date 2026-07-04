package vectordb

import (
	"context"
	"testing"
)

func TestRegister_Valid(t *testing.T) {
	name := "test-register-vectordb"
	Register(name, func(cfg Config) (VectorSearcher, error) {
		return newMockStore(), nil
	})
	if !sliceHas(Registered(), name) {
		t.Errorf("expected %q to be registered", name)
	}
}

func TestRegister_EmptyName(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic for empty name")
		}
	}()
	Register("", func(cfg Config) (VectorSearcher, error) { return nil, nil })
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
	name := "test-dup-vectordb"
	f := func(cfg Config) (VectorSearcher, error) { return newMockStore(), nil }
	Register(name, f)

	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic for duplicate registration")
		}
	}()
	Register(name, f)
}

func TestOpen_Valid(t *testing.T) {
	name := "test-open-valid-vectordb"
	Register(name, func(cfg Config) (VectorSearcher, error) {
		return newMockStore(), nil
	})

	store, err := Open(name)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if store == nil {
		t.Fatal("Open returned nil VectorSearcher")
	}

	results, err := store.Search(context.Background(), []float64{0.1})
	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}
	if results == nil {
		t.Error("Search returned nil results")
	}
}

func TestOpen_Unknown(t *testing.T) {
	_, err := Open("nonexistent-store-xyz")
	if err == nil {
		t.Error("expected error for unknown provider")
	}
}

func TestOpen_WithOptions(t *testing.T) {
	name := "test-open-opts-vectordb"
	Register(name, func(cfg Config) (VectorSearcher, error) {
		if cfg.URL != "http://test:6333" {
			t.Errorf("URL: got %q, want %q", cfg.URL, "http://test:6333")
		}
		return newMockStore(), nil
	})

	_, err := Open(name, WithURL("http://test:6333"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRegistered_NotNil(t *testing.T) {
	if Registered() == nil {
		t.Error("Registered() should not return nil")
	}
}

func sliceHas(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
