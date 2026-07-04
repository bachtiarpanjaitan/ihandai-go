// Package tools defines the interface for LLM-callable tools (function calling).
//
// Tools are functions that an LLM can request to execute. The LLM decides
// when to call a tool based on the user's request and the tool's description.
//
// This is distinct from MCP (Model Context Protocol), which provides
// resources and context to LLMs rather than functions they can invoke.
package tools

import (
	"context"
	"encoding/json"

	"github.com/bachtiarpanjaitan/ihandai-go/pkg/core"
)

// Tool is a function that an LLM can request to call.
//
// Each tool has a name, description, and JSON Schema for its input.
// The LLM sees the name, description, and schema, and decides whether
// to call the tool based on the user's request.
type Tool interface {
	// Name returns the unique name of the tool (e.g., "send_email", "web_search").
	Name() string

	// Description explains what the tool does. The LLM uses this to decide
	// when to call the tool. Be specific and include example usage.
	Description() string

	// InputSchema returns the JSON Schema describing the tool's input parameters.
	InputSchema() *core.JSONSchema

	// Execute runs the tool with the given input and returns the result.
	// The input is the JSON-encoded arguments from the LLM.
	// The output is JSON-encoded result to return to the LLM.
	// The context controls cancellation and carries tracing spans.
	Execute(ctx context.Context, input json.RawMessage) (json.RawMessage, error)
}

// Registry manages a collection of named tools.
// It is safe for concurrent use.
type Registry struct {
	tools map[string]Tool
}

// NewRegistry creates an empty tool registry.
func NewRegistry() *Registry {
	return &Registry{tools: make(map[string]Tool)}
}

// Register adds a tool to the registry.
// It panics if a tool with the same name is already registered.
func (r *Registry) Register(t Tool) {
	name := t.Name()
	if _, exists := r.tools[name]; exists {
		panic("tools: duplicate registration for tool " + name)
	}
	r.tools[name] = t
}

// Get returns a tool by name, or nil if not found.
func (r *Registry) Get(name string) Tool {
	return r.tools[name]
}

// List returns all registered tool names.
func (r *Registry) List() []string {
	names := make([]string, 0, len(r.tools))
	for name := range r.tools {
		names = append(names, name)
	}
	return names
}

// Len returns the number of registered tools.
func (r *Registry) Len() int {
	return len(r.tools)
}
