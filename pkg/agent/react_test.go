package agent

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/bachtiarpanjaitan/ihandai-go/pkg/core"
	"github.com/bachtiarpanjaitan/ihandai-go/pkg/llm"
	"github.com/bachtiarpanjaitan/ihandai-go/pkg/tools"
)

// mockLLM simulates tool-calling behavior for testing.
type mockAgentLLM struct {
	responses []string
	idx       int
}

func (m *mockAgentLLM) Chat(ctx context.Context, msgs []core.Message) (*core.Response, error) {
	if m.idx >= len(m.responses) {
		return &core.Response{Content: "Final Answer: Task complete."}, nil
	}
	resp := m.responses[m.idx]
	m.idx++
	return &core.Response{Content: resp}, nil
}

type mockTool struct{}

func (m mockTool) Name() string                  { return "echo" }
func (m mockTool) Description() string            { return "Echoes back the input" }
func (m mockTool) InputSchema() *core.JSONSchema   { return &core.JSONSchema{Type: "object", Properties: map[string]*core.JSONSchemaProp{"message": {Type: "string"}}} }
func (m mockTool) Execute(ctx context.Context, input json.RawMessage) (json.RawMessage, error) {
	return input, nil
}

func TestReAct_SingleToolCall(t *testing.T) {
	llmMock := &mockAgentLLM{
		responses: []string{
			"Thought: I need to echo a message.\nAction: echo({\"message\":\"hello\"})",
			"Thought: I received the echo.\nFinal Answer: The echo returned hello.",
		},
	}

	a := NewReAct(Config{
		LLM:   llmMock,
		Tools: []tools.Tool{mockTool{}},
	})

	result, err := a.Run(context.Background(), "Echo hello")
	if err != nil {
		t.Fatalf("Run failed: %v", err)
	}
	if result.Answer != "The echo returned hello." {
		t.Errorf("Answer: got %q, want %q", result.Answer, "The echo returned hello.")
	}
	if len(result.Steps) < 2 {
		t.Errorf("expected at least 2 steps, got %d", len(result.Steps))
	}
}

func TestReAct_DirectAnswer(t *testing.T) {
	llmMock := &mockAgentLLM{
		responses: []string{
			"Thought: The answer is obvious.\nFinal Answer: 42",
		},
	}

	a := NewReAct(Config{LLM: llmMock})
	result, err := a.Run(context.Background(), "What is the answer?")
	if err != nil {
		t.Fatalf("Run failed: %v", err)
	}
	if result.Answer != "42" {
		t.Errorf("Answer: got %q, want %q", result.Answer, "42")
	}
}

func TestReAct_MaxSteps(t *testing.T) {
	// Never gives final answer
	llmMock := &mockAgentLLM{
		responses: []string{
			"Thought: Hmm.\nAction: echo({\"message\":\"loop\"})",
			"Thought: Again.\nAction: echo({\"message\":\"loop\"})",
			"Thought: More.\nAction: echo({\"message\":\"loop\"})",
			"Thought: Still.\nAction: echo({\"message\":\"loop\"})",
			"Thought: ...\nAction: echo({\"message\":\"loop\"})",
		},
	}

	a := NewReAct(Config{
		LLM:      llmMock,
		Tools:    []tools.Tool{mockTool{}},
		MaxSteps: 3,
	})

	result, err := a.Run(context.Background(), "Test max steps")
	if err != nil {
		t.Fatalf("Run failed: %v", err)
	}
	if result.Answer == "" {
		t.Error("expected non-empty fallback answer")
	}
}

func TestParseResponse_FinalAnswer(t *testing.T) {
	r := &ReAct{}
	step, isFinal := r.parseResponse("Thought: I know it.\nFinal Answer: The capital is Paris.")
	if !isFinal {
		t.Error("expected final answer")
	}
	if step.Observation != "The capital is Paris." {
		t.Errorf("got %q, want %q", step.Observation, "The capital is Paris.")
	}
}

func TestParseResponse_Action(t *testing.T) {
	r := &ReAct{}
	step, isFinal := r.parseResponse(`Thought: I need to search.
Action: calculator({"expression":"2+2"})`)
	if isFinal {
		t.Error("expected action, not final")
	}
	if step.Action.Name != "calculator" {
		t.Errorf("Action name: got %q, want %q", step.Action.Name, "calculator")
	}
	if step.Action.Input != `{"expression":"2+2"}` {
		t.Errorf("Action input: got %q", step.Action.Input)
	}
}

func TestReAct_ContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	a := NewReAct(Config{
		LLM: &mockAgentLLM{},
	})
	_, err := a.Run(ctx, "test")
	if err == nil {
		t.Error("expected error for canceled context")
	}
}

func TestNewReAct_Defaults(t *testing.T) {
	a := NewReAct(Config{LLM: &mockAgentLLM{}})
	if a.cfg.MaxSteps != 10 {
		t.Errorf("MaxSteps: got %d, want 10", a.cfg.MaxSteps)
	}
	if a.cfg.SystemPrompt != defaultSystemPrompt {
		t.Error("SystemPrompt should default")
	}
}

// Compile-time check
var _ Agent = (*ReAct)(nil)
var _ llm.ChatCompleter = (*mockAgentLLM)(nil)
