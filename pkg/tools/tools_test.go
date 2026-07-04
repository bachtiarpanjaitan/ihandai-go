package tools

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/bachtiarpanjaitan/ihandai-go/pkg/core"
)

// Compile-time interface satisfaction check
var _ Tool = (*mockTool)(nil)

type mockTool struct{}

func (m mockTool) Name() string                  { return "mock_tool" }
func (m mockTool) Description() string           { return "A mock tool for testing" }
func (m mockTool) InputSchema() *core.JSONSchema { return &core.JSONSchema{Type: "object"} }
func (m mockTool) Execute(ctx context.Context, input json.RawMessage) (json.RawMessage, error) {
	return json.RawMessage(`{"status":"ok"}`), nil
}

func TestTool_Mock(t *testing.T) {
	var tool Tool = mockTool{}

	if tool.Name() != "mock_tool" {
		t.Errorf("Name: got %q, want %q", tool.Name(), "mock_tool")
	}
	if tool.Description() != "A mock tool for testing" {
		t.Errorf("Description mismatch")
	}
	if tool.InputSchema() == nil {
		t.Error("InputSchema should not be nil")
	}
	if tool.InputSchema().Type != "object" {
		t.Errorf("Schema Type: got %q, want %q", tool.InputSchema().Type, "object")
	}

	result, err := tool.Execute(context.Background(), json.RawMessage(`{}`))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(result) != `{"status":"ok"}` {
		t.Errorf("got %s, want %s", string(result), `{"status":"ok"}`)
	}
}

func TestRegistry(t *testing.T) {
	r := NewRegistry()
	if r.Len() != 0 {
		t.Errorf("new registry should be empty, got %d tools", r.Len())
	}

	tool := mockTool{}
	r.Register(tool)

	if r.Len() != 1 {
		t.Errorf("expected 1 tool, got %d", r.Len())
	}

	retrieved := r.Get("mock_tool")
	if retrieved == nil {
		t.Fatal("Get returned nil for registered tool")
	}
	if retrieved.Name() != "mock_tool" {
		t.Errorf("got %q, want %q", retrieved.Name(), "mock_tool")
	}

	names := r.List()
	if len(names) != 1 || names[0] != "mock_tool" {
		t.Errorf("List: got %v, want [mock_tool]", names)
	}

	// Get non-existent tool
	if r.Get("nonexistent") != nil {
		t.Error("Get should return nil for unknown tool")
	}
}

func TestRegistry_Duplicate(t *testing.T) {
	r := NewRegistry()
	r.Register(mockTool{})

	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic for duplicate registration")
		}
	}()
	r.Register(mockTool{})
}
