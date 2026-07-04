package ihandai

import (
	"context"
	"log/slog"
	"os"
	"sync"
	"testing"
)

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

	// Last option wins
	ai, err := New(WithLogger(logger1), WithLogger(logger2))
	if err != nil {
		t.Fatalf("New() unexpected error: %v", err)
	}
	if ai.cfg.Logger != logger2 {
		t.Error("last WithLogger should win")
	}

	ai.Close()
}

func TestClient_Config(t *testing.T) {
	ai, _ := New()
	cfg := ai.Config()
	if cfg == nil {
		t.Fatal("Config() returned nil")
	}
	// Verify it's a copy, not the original
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
	if err := ai.Close(); err != nil {
		t.Errorf("third Close() should not error: %v", err)
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

	wg.Wait()
}

func TestClient_Ask_Placeholder(t *testing.T) {
	ai, _ := New()
	defer ai.Close()

	resp, err := ai.Ask(context.TODO(), "test")
	if err != nil {
		t.Errorf("placeholder Ask should not error: %v", err)
	}
	if resp != nil {
		t.Error("placeholder Ask should return nil response")
	}
}

func TestClient_Index_Placeholder(t *testing.T) {
	ai, _ := New()
	defer ai.Close()

	err := ai.Index(context.TODO(), []Document{{ID: "1", Content: "test"}})
	if err != nil {
		t.Errorf("placeholder Index should not error: %v", err)
	}
}
