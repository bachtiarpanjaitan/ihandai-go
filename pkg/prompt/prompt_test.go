package prompt

import (
	"context"
	"errors"
	"testing"

	"github.com/bachtiarpanjaitan/ihandai-go"
)

// Compile-time interface satisfaction check
var _ PromptBuilder = (*mockBuilder)(nil)
var _ PromptBuilder = (*failingBuilder)(nil)

type mockBuilder struct{}

func (m mockBuilder) Build(ctx context.Context, template string, contextData map[string]any) ([]ihandai.Message, error) {
	return []ihandai.Message{
		{Role: "system", Content: template},
		{Role: "user", Content: contextData["query"].(string)},
	}, nil
}

type failingBuilder struct{}

func (f failingBuilder) Build(ctx context.Context, template string, contextData map[string]any) ([]ihandai.Message, error) {
	return nil, errors.New("build failed")
}

func TestPromptBuilder_Mock(t *testing.T) {
	var b PromptBuilder = mockBuilder{}
	msgs, err := b.Build(context.Background(), "You are helpful.", map[string]any{
		"query": "hello",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(msgs) != 2 {
		t.Errorf("got %d messages, want 2", len(msgs))
	}
	if msgs[0].Role != "system" {
		t.Errorf("Role[0]: got %q, want %q", msgs[0].Role, "system")
	}
	if msgs[1].Role != "user" {
		t.Errorf("Role[1]: got %q, want %q", msgs[1].Role, "user")
	}
}

func TestPromptBuilder_Failure(t *testing.T) {
	var b PromptBuilder = failingBuilder{}
	_, err := b.Build(context.Background(), "", nil)
	if err == nil {
		t.Error("expected error from failing builder")
	}
}
