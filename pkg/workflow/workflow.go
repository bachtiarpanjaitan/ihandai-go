// Package workflow defines a DAG-based workflow engine for orchestrating
// multi-step AI pipelines with parallel execution and conditional branching.
package workflow

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
)

// StepFunc is the function signature for a workflow step.
// It receives the outputs of its dependencies and returns its own output.
type StepFunc func(ctx context.Context, inputs map[string]any) (map[string]any, error)

// Step defines a single node in the workflow DAG.
type Step struct {
	// Name uniquely identifies this step within the workflow.
	Name string `json:"name"`

	// Run is the function to execute.
	Run StepFunc `json:"-"`

	// DependsOn lists the names of steps that must complete before this one.
	DependsOn []string `json:"depends_on,omitempty"`

	// Condition is an optional condition function. If it returns false,
	// this step (and all steps that depend on it) are skipped.
	Condition StepFunc `json:"-"`
}

// StepResult holds the output of a completed step.
type StepResult struct {
	Name   string
	Output map[string]any
	Error  error
	Skipped bool
}

// Workflow is a DAG of steps that execute with dependency ordering.
type Workflow struct {
	name  string
	steps map[string]*Step
}

// New creates a new empty workflow.
func New(name string) *Workflow {
	return &Workflow{name: name, steps: make(map[string]*Step)}
}

// AddStep adds a step to the workflow. Steps with duplicate names replace
// previous definitions.
func (w *Workflow) AddStep(s *Step) {
	w.steps[s.Name] = s
}

// Step returns a step by name, or nil.
func (w *Workflow) Step(name string) *Step { return w.steps[name] }

// Name returns the workflow name.
func (w *Workflow) Name() string { return w.name }

// Steps returns all steps in the workflow (order not guaranteed).
func (w *Workflow) Steps() []*Step {
	result := make([]*Step, 0, len(w.steps))
	for _, s := range w.steps {
		result = append(result, s)
	}
	return result
}

// Validate checks the workflow for common errors: cycles, missing dependencies.
func (w *Workflow) Validate() error {
	if len(w.steps) == 0 {
		return fmt.Errorf("workflow %q has no steps", w.name)
	}
	for _, s := range w.steps {
		for _, dep := range s.DependsOn {
			if _, ok := w.steps[dep]; !ok {
				return fmt.Errorf("step %q depends on unknown step %q", s.Name, dep)
			}
		}
	}
	if _, err := w.topologicalSort(); err != nil {
		return err
	}
	return nil
}

// Save serializes the workflow structure to a JSON file.
// Note: Run and Condition functions are not serialized.
func (w *Workflow) Save(path string) error {
	type stepJSON struct {
		Name      string   `json:"name"`
		DependsOn []string `json:"depends_on,omitempty"`
	}
	type wfJSON struct {
		Name  string      `json:"name"`
		Steps []stepJSON  `json:"steps"`
	}

	j := wfJSON{Name: w.name}
	for _, s := range w.steps {
		j.Steps = append(j.Steps, stepJSON{Name: s.Name, DependsOn: s.DependsOn})
	}

	data, err := json.MarshalIndent(j, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

// Load deserializes a workflow structure from a JSON file.
// Functions must be re-attached via SetFunc before running.
func Load(path string) (*Workflow, error) {
	type stepJSON struct {
		Name      string   `json:"name"`
		DependsOn []string `json:"depends_on,omitempty"`
	}
	type wfJSON struct {
		Name  string      `json:"name"`
		Steps []stepJSON  `json:"steps"`
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var j wfJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return nil, err
	}

	w := New(j.Name)
	for _, s := range j.Steps {
		w.AddStep(&Step{Name: s.Name, DependsOn: s.DependsOn})
	}
	return w, nil
}

// SetFunc sets the Run function for a named step. Use after Load().
func (w *Workflow) SetFunc(name string, fn StepFunc) {
	if s := w.steps[name]; s != nil {
		s.Run = fn
	}
}

// SetCondition sets the Condition function for a named step.
func (w *Workflow) SetCondition(name string, fn StepFunc) {
	if s := w.steps[name]; s != nil {
		s.Condition = fn
	}
}
