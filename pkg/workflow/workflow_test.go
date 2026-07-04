package workflow

import (
	"context"
	"os"
	"path/filepath"
	"sync/atomic"
	"testing"
	"time"
)

func TestWorkflow_Linear(t *testing.T) {
	w := New("linear-test")
	w.AddStep(&Step{Name: "step1", Run: func(ctx context.Context, inputs map[string]any) (map[string]any, error) {
		return map[string]any{"a": 1}, nil
	}})
	w.AddStep(&Step{Name: "step2", Run: func(ctx context.Context, inputs map[string]any) (map[string]any, error) {
		return map[string]any{"b": inputs["a"].(int) + 1}, nil
	}, DependsOn: []string{"step1"}})

	result, err := w.Run(context.Background(), nil)
	if err != nil {
		t.Fatalf("Run failed: %v", err)
	}
	if result["a"] != 1 {
		t.Errorf("a: got %v, want 1", result["a"])
	}
	if result["b"] != 2 {
		t.Errorf("b: got %v, want 2", result["b"])
	}
}

func TestWorkflow_Parallel(t *testing.T) {
	w := New("parallel-test")
	var order atomic.Int32
	var counter int32

	w.AddStep(&Step{Name: "init", Run: func(ctx context.Context, inputs map[string]any) (map[string]any, error) {
		return map[string]any{"x": 1}, nil
	}})
	w.AddStep(&Step{Name: "branch-a", Run: func(ctx context.Context, inputs map[string]any) (map[string]any, error) {
		time.Sleep(10 * time.Millisecond)
		order.Add(1)
		return map[string]any{"a": "done"}, nil
	}, DependsOn: []string{"init"}})
	w.AddStep(&Step{Name: "branch-b", Run: func(ctx context.Context, inputs map[string]any) (map[string]any, error) {
		time.Sleep(10 * time.Millisecond)
		order.Add(1)
		return map[string]any{"b": "done"}, nil
	}, DependsOn: []string{"init"}})
	w.AddStep(&Step{Name: "merge", Run: func(ctx context.Context, inputs map[string]any) (map[string]any, error) {
		counter = order.Load()
		return map[string]any{"merged": true}, nil
	}, DependsOn: []string{"branch-a", "branch-b"}})

	result, err := w.Run(context.Background(), nil)
	if err != nil {
		t.Fatalf("Run failed: %v", err)
	}
	if result["merged"] != true {
		t.Error("merge step should have run")
	}
	if counter != 2 {
		t.Errorf("both branches should complete before merge, counter=%d", counter)
	}
}

func TestWorkflow_ConditionalSkip(t *testing.T) {
	w := New("conditional-test")
	shouldRun := false

	w.AddStep(&Step{Name: "check", Run: func(ctx context.Context, inputs map[string]any) (map[string]any, error) {
		return map[string]any{"run": shouldRun}, nil
	}})
	w.AddStep(&Step{
		Name: "conditional",
		Condition: func(ctx context.Context, inputs map[string]any) (map[string]any, error) {
			if run, _ := inputs["run"].(bool); !run {
				return nil, nil // skip
			}
			return map[string]any{"ok": true}, nil
		},
		Run: func(ctx context.Context, inputs map[string]any) (map[string]any, error) {
			return map[string]any{"executed": true}, nil
		},
		DependsOn: []string{"check"},
	})
	w.AddStep(&Step{Name: "after", Run: func(ctx context.Context, inputs map[string]any) (map[string]any, error) {
		_, wasExecuted := inputs["executed"]
		return map[string]any{"saw_executed": wasExecuted}, nil
	}, DependsOn: []string{"conditional"}})

	result, err := w.Run(context.Background(), nil)
	if err != nil {
		t.Fatalf("Run failed: %v", err)
	}
	if _, ok := result["saw_executed"]; ok {
		t.Error("after step should not produce output when conditional was skipped")
	}
}

func TestWorkflow_DiamondShape(t *testing.T) {
	w := Build("diamond").
		Add("start", func(ctx context.Context, inputs map[string]any) (map[string]any, error) {
			return map[string]any{"value": 10}, nil
		}).
		Add("double", func(ctx context.Context, inputs map[string]any) (map[string]any, error) {
			v := inputs["value"].(int)
			return map[string]any{"doubled": v * 2}, nil
		}, "start").
		Add("triple", func(ctx context.Context, inputs map[string]any) (map[string]any, error) {
			v := inputs["value"].(int)
			return map[string]any{"tripled": v * 3}, nil
		}, "start").
		Add("sum", func(ctx context.Context, inputs map[string]any) (map[string]any, error) {
			return map[string]any{
				"doubled": inputs["doubled"],
				"tripled": inputs["tripled"],
			}, nil
		}, "double", "triple").
		MustWorkflow()

	result, err := w.Run(context.Background(), nil)
	if err != nil {
		t.Fatalf("Run failed: %v", err)
	}
	if result["doubled"] != 20 {
		t.Errorf("doubled: got %v, want 20", result["doubled"])
	}
	if result["tripled"] != 30 {
		t.Errorf("tripled: got %v, want 30", result["tripled"])
	}
}

func TestWorkflow_Validation(t *testing.T) {
	t.Run("empty workflow", func(t *testing.T) {
		w := New("empty")
		err := w.Validate()
		if err == nil {
			t.Error("expected error for empty workflow")
		}
	})

	t.Run("missing dependency", func(t *testing.T) {
		w := New("missing-dep")
		w.AddStep(&Step{Name: "s1", DependsOn: []string{"nonexistent"}})
		err := w.Validate()
		if err == nil {
			t.Error("expected error for missing dependency")
		}
	})

	t.Run("cycle detection", func(t *testing.T) {
		w := New("cycle")
		w.AddStep(&Step{Name: "a", DependsOn: []string{"b"}})
		w.AddStep(&Step{Name: "b", DependsOn: []string{"a"}})
		err := w.Validate()
		if err == nil {
			t.Error("expected cycle detection error")
		}
	})
}

func TestWorkflow_Serialization(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "workflow.json")

	// Create and save
	w := Build("test-wf").
		Add("step1", nil).
		Add("step2", nil, "step1").
		MustWorkflow()

	if err := w.Save(path); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	// Load and verify
	loaded, err := Load(path)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}
	if loaded.Name() != "test-wf" {
		t.Errorf("Name: got %q, want %q", loaded.Name(), "test-wf")
	}
	if len(loaded.Steps()) != 2 {
		t.Errorf("got %d steps, want 2", len(loaded.Steps()))
	}

	// Set functions and run
	executed := false
	loaded.SetFunc("step1", func(ctx context.Context, inputs map[string]any) (map[string]any, error) {
		executed = true
		return map[string]any{"ok": true}, nil
	})
	loaded.SetFunc("step2", func(ctx context.Context, inputs map[string]any) (map[string]any, error) {
		return map[string]any{"done": true}, nil
	})

	_, err = loaded.Run(context.Background(), nil)
	if err != nil {
		t.Fatalf("Run after load failed: %v", err)
	}
	if !executed {
		t.Error("step1 should have executed")
	}
}

func TestWorkflow_ContextCancellation(t *testing.T) {
	w := New("cancel-test")
	w.AddStep(&Step{Name: "respects-ctx", Run: func(ctx context.Context, inputs map[string]any) (map[string]any, error) {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(500 * time.Millisecond):
			return map[string]any{}, nil
		}
	}})

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	_, err := w.Run(ctx, nil)
	if err == nil {
		t.Error("expected timeout/cancellation error")
	}
}

func TestWorkflow_Error(t *testing.T) {
	w := New("error-test")
	w.AddStep(&Step{Name: "failing", Run: func(ctx context.Context, inputs map[string]any) (map[string]any, error) {
		return nil, os.ErrNotExist
	}})

	_, err := w.Run(context.Background(), nil)
	if err == nil {
		t.Error("expected error from failing step")
	}
}

func TestWorkflow_EmptyInputs(t *testing.T) {
	w := New("nil-test")
	w.AddStep(&Step{Name: "returns-nil", Run: func(ctx context.Context, inputs map[string]any) (map[string]any, error) {
		return nil, nil
	}})

	result, err := w.Run(context.Background(), nil)
	if err != nil {
		t.Fatalf("Run failed: %v", err)
	}
	if result == nil {
		t.Error("result should not be nil")
	}
}
