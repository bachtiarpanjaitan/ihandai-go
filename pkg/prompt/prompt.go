// Package prompt defines the interface for building prompts from templates
// and context data.
//
// Prompt builders format retrieved documents, user queries, and system
// instructions into the message format expected by LLMs.
package prompt

import (
	"context"

	"github.com/bachtiarpanjaitan/ihandai-go"
)

// PromptBuilder builds chat messages from templates and context data.
//
// Implementations may use Go templates, string interpolation, or
// more sophisticated prompt engineering techniques.
type PromptBuilder interface {
	// Build constructs chat messages from a template and context data.
	// The context controls cancellation and carries tracing spans.
	//
	// The template is a provider-specific format string (e.g., Go template,
	// Jinja2, or plain text with placeholders).
	// The contextData contains values like retrieved documents, user query,
	// conversation history, and system instructions.
	Build(ctx context.Context, template string, contextData map[string]any) ([]ihandai.Message, error)
}
