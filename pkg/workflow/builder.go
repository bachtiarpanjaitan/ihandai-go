package workflow

// Builder provides a fluent API for constructing workflows.
type Builder struct {
	wf  *Workflow
	err error
}

// Build creates a new workflow builder.
func Build(name string) *Builder {
	return &Builder{wf: New(name)}
}

// Add adds a step to the workflow.
// name is the step identifier.
// fn is the step function.
// deps are the names of prerequisite steps.
func (b *Builder) Add(name string, fn StepFunc, deps ...string) *Builder {
	if b.err != nil {
		return b
	}
	b.wf.AddStep(&Step{
		Name:      name,
		Run:       fn,
		DependsOn: deps,
	})
	return b
}

// AddConditional adds a conditional step. If the condition returns no output,
// the step and its dependents are skipped.
func (b *Builder) AddConditional(name string, condition StepFunc, fn StepFunc, deps ...string) *Builder {
	if b.err != nil {
		return b
	}
	b.wf.AddStep(&Step{
		Name:      name,
		Run:       fn,
		Condition: condition,
		DependsOn: deps,
	})
	return b
}

// Workflow returns the constructed workflow.
func (b *Builder) Workflow() (*Workflow, error) {
	if b.err != nil {
		return nil, b.err
	}
	return b.wf, b.wf.Validate()
}

// MustWorkflow returns the workflow or panics if validation fails.
func (b *Builder) MustWorkflow() *Workflow {
	w, err := b.Workflow()
	if err != nil {
		panic(err)
	}
	return w
}
