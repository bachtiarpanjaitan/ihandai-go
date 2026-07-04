// Package core defines shared types used across all ihandai packages.
//
// These types are defined here to avoid import cycles between the root
// package and the sub-packages (pkg/llm, pkg/embedding, etc.).
package core

// Message represents a chat message in a conversation.
type Message struct {
	Role    string
	Content string
	Name    string
}

// Response represents a response from an LLM chat completion.
type Response struct {
	Content      string
	FinishReason string
	ToolCalls    []ToolCall
	Usage        *TokenUsage
}

// ToolCall represents a tool call requested by an LLM.
type ToolCall struct {
	ID        string
	Name      string
	Arguments []byte
}

// TokenUsage contains token usage information for a request.
type TokenUsage struct {
	PromptTokens     int
	CompletionTokens int
	TotalTokens      int
}

// Document represents a document to be indexed or retrieved.
type Document struct {
	ID       string
	Content  string
	Metadata map[string]any
}

// Chunk represents a split portion of a Document.
type Chunk struct {
	ID       string
	Content  string
	Metadata map[string]any
	ParentID string
}

// ScoredDocument is a Document with a relevance score.
type ScoredDocument struct {
	Document
	Score float64
}

// JSONSchema represents a JSON Schema definition for tool input validation.
type JSONSchema struct {
	Type        string                     `json:"type"`
	Properties  map[string]*JSONSchemaProp `json:"properties,omitempty"`
	Required    []string                   `json:"required,omitempty"`
	Description string                     `json:"description,omitempty"`
}

// JSONSchemaProp represents a property within a JSON Schema.
type JSONSchemaProp struct {
	Type        string          `json:"type"`
	Description string          `json:"description,omitempty"`
	Enum        []string        `json:"enum,omitempty"`
	Items       *JSONSchemaProp `json:"items,omitempty"`
}
