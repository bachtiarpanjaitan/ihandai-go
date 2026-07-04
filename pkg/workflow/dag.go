package workflow

import (
	"context"
	"fmt"
	"sync"
)

// Run executes the workflow with the given initial inputs.
// Steps are executed in dependency order. Independent steps run concurrently.
func (w *Workflow) Run(ctx context.Context, inputs map[string]any) (map[string]any, error) {
	if err := w.Validate(); err != nil {
		return nil, err
	}

	ordered, err := w.topologicalSort()
	if err != nil {
		return nil, err
	}

	// Execute level by level (steps in the same level can run concurrently)
	results := make(map[string]*StepResult)
	finalOutput := make(map[string]any)

	// Copy initial inputs
	for k, v := range inputs {
		finalOutput[k] = v
	}

	level := 0
	remaining := make([]string, len(ordered))
	copy(remaining, ordered)

	for len(remaining) > 0 {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		// Find steps whose dependencies are all resolved
		var ready []string
		var notReady []string
		for _, name := range remaining {
			step := w.steps[name]
			if allDepsResolved(step, results) {
				ready = append(ready, name)
			} else {
				notReady = append(notReady, name)
			}
		}
		remaining = notReady

		if len(ready) == 0 {
			return nil, fmt.Errorf("workflow %q: deadlock at level %d — %d steps waiting", w.name, level, len(remaining))
		}

		// Execute ready steps in parallel
		levelResults := w.executeLevel(ctx, ready, results)
		for name, r := range levelResults {
			results[name] = r
			if r.Error != nil {
				return nil, fmt.Errorf("workflow %q: step %q failed: %w", w.name, name, r.Error)
			}
			if !r.Skipped {
				for k, v := range r.Output {
					finalOutput[k] = v
				}
			}
		}

		// Check for conditional cascading: if a step was skipped, skip dependents
		for _, name := range remaining {
			step := w.steps[name]
			shouldSkip := false
			for _, dep := range step.DependsOn {
				if r, ok := results[dep]; ok && r.Skipped {
					shouldSkip = true
					break
				}
			}
			if shouldSkip {
				results[name] = &StepResult{Name: name, Skipped: true}
				remaining = removeStr(remaining, name)
			}
		}

		level++
	}

	return finalOutput, nil
}

func (w *Workflow) executeLevel(ctx context.Context, names []string, results map[string]*StepResult) map[string]*StepResult {
	if len(names) == 1 {
		// Single step — execute directly
		name := names[0]
		step := w.steps[name]
		r := w.runStep(ctx, step, results)
		return map[string]*StepResult{name: r}
	}

	// Multiple steps — execute in parallel
	var mu sync.Mutex
	var wg sync.WaitGroup
	levelResults := make(map[string]*StepResult)

	for _, name := range names {
		wg.Add(1)
		go func(n string) {
			defer wg.Done()
			step := w.steps[n]
			r := w.runStep(ctx, step, results)
			mu.Lock()
			levelResults[n] = r
			mu.Unlock()
		}(name)
	}
	wg.Wait()

	return levelResults
}

func (w *Workflow) runStep(ctx context.Context, step *Step, results map[string]*StepResult) *StepResult {
	// Check condition
	if step.Condition != nil {
		inputs := collectInputs(step, results)
		condResult, err := step.Condition(ctx, inputs)
		if err != nil {
			return &StepResult{Name: step.Name, Error: fmt.Errorf("condition failed: %w", err)}
		}
		// If condition returns empty map, skip this step
		if len(condResult) == 0 {
			return &StepResult{Name: step.Name, Skipped: true}
		}
	}

	// Build inputs from dependencies
	inputs := collectInputs(step, results)

	output, err := step.Run(ctx, inputs)
	if err != nil {
		return &StepResult{Name: step.Name, Error: err}
	}
	if output == nil {
		output = make(map[string]any)
	}
	return &StepResult{Name: step.Name, Output: output}
}

// collectInputs gathers outputs from all dependency steps.
func collectInputs(step *Step, results map[string]*StepResult) map[string]any {
	inputs := make(map[string]any)
	for _, dep := range step.DependsOn {
		if r, ok := results[dep]; ok && r != nil && !r.Skipped {
			for k, v := range r.Output {
				inputs[k] = v
			}
		}
	}
	return inputs
}

// topologicalSort returns step names in dependency order (Kahn's algorithm).
func (w *Workflow) topologicalSort() ([]string, error) {
	inDegree := make(map[string]int)
	for name := range w.steps {
		inDegree[name] = 0
	}
	for _, s := range w.steps {
		for range s.DependsOn {
			inDegree[s.Name]++
		}
	}

	var queue []string
	for name, deg := range inDegree {
		if deg == 0 {
			queue = append(queue, name)
		}
	}

	var sorted []string
	visited := 0
	for len(queue) > 0 {
		name := queue[0]
		queue = queue[1:]
		sorted = append(sorted, name)
		visited++

		// Decrease in-degree of dependents
		for _, s := range w.steps {
			for _, dep := range s.DependsOn {
				if dep == name {
					inDegree[s.Name]--
					if inDegree[s.Name] == 0 {
						queue = append(queue, s.Name)
					}
				}
			}
		}
	}

	if visited != len(w.steps) {
		return nil, fmt.Errorf("workflow %q: cycle detected", w.name)
	}

	return sorted, nil
}

func allDepsResolved(step *Step, results map[string]*StepResult) bool {
	for _, dep := range step.DependsOn {
		if _, ok := results[dep]; !ok {
			return false
		}
	}
	return true
}

func removeStr(slice []string, item string) []string {
	for i, s := range slice {
		if s == item {
			return append(slice[:i], slice[i+1:]...)
		}
	}
	return slice
}
