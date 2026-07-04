package tools

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/bachtiarpanjaitan/ihandai-go/pkg/tools"
)

// Compile-time check
var _ tools.Tool = (*Calculator)(nil)

func TestCalculator_Addition(t *testing.T) {
	c := NewCalculator()
	result, err := c.Execute(context.Background(), json.RawMessage(`{"expression":"2 + 3"}`))
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}
	if string(result) != `{"result":5}` {
		t.Errorf("got %s, want %s", string(result), `{"result":5}`)
	}
}

func TestCalculator_Multiplication(t *testing.T) {
	c := NewCalculator()
	result, err := c.Execute(context.Background(), json.RawMessage(`{"expression":"4 * 5"}`))
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}
	if string(result) != `{"result":20}` {
		t.Errorf("got %s, want %s", string(result), `{"result":20}`)
	}
}

func TestCalculator_Division(t *testing.T) {
	c := NewCalculator()
	result, err := c.Execute(context.Background(), json.RawMessage(`{"expression":"10/2"}`))
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}
	if string(result) != `{"result":5}` {
		t.Errorf("got %s, want %s", string(result), `{"result":5}`)
	}
}

func TestCalculator_DivisionByZero(t *testing.T) {
	c := NewCalculator()
	result, err := c.Execute(context.Background(), json.RawMessage(`{"expression":"5/0"}`))
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}
	if !contains(string(result), "error") {
		t.Errorf("expected error, got %s", string(result))
	}
}

func TestCalculator_Parentheses(t *testing.T) {
	c := NewCalculator()
	result, err := c.Execute(context.Background(), json.RawMessage(`{"expression":"(2+3)*4"}`))
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}
	if string(result) != `{"result":20}` {
		t.Errorf("got %s, want %s", string(result), `{"result":20}`)
	}
}

func TestCalculator_ToolInterface(t *testing.T) {
	c := NewCalculator()
	if c.Name() != "calculator" {
		t.Errorf("Name: got %q", c.Name())
	}
	if c.Description() == "" {
		t.Error("Description should not be empty")
	}
	if c.InputSchema() == nil {
		t.Error("InputSchema should not be nil")
	}
}

func contains(s, sub string) bool {
	return len(s) >= len(sub) && searchStr(s, sub)
}

func searchStr(s, sub string) bool {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
