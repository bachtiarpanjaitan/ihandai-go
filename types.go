// Package ihandai is a provider-agnostic AI library for Go.
// It provides composable interfaces for LLM, embedding, vector stores,
// and orchestrates them into RAG pipelines.
package ihandai

// Message represents a chat message in a conversation.
type Message struct {
	// Role is the role of the message author.
	// Must be one of: "system", "user", "assistant", "tool".
	Role string

	// Content is the text content of the message.
	Content string

	// Name is an optional name for the message author.
	// Used for tool calls to identify which tool was called.
	Name string
}

// Response represents a response from an LLM chat completion.
type Response struct {
	// Content is the text content of the response.
	Content string

	// FinishReason indicates why the model stopped generating.
	// Common values: "stop", "length", "tool_calls", "content_filter".
	FinishReason string

	// ToolCalls contains any tool calls the model requested.
	ToolCalls []ToolCall

	// Usage contains token usage information, if provided by the provider.
	Usage *TokenUsage
}

// ToolCall represents a tool call requested by an LLM.
type ToolCall struct {
	// ID is the unique identifier for this tool call.
	ID string

	// Name is the name of the tool to call.
	Name string

	// Arguments is the JSON-encoded arguments for the tool.
	Arguments []byte
}

// TokenUsage contains token usage information for a request.
type TokenUsage struct {
	// PromptTokens is the number of tokens in the prompt.
	PromptTokens int

	// CompletionTokens is the number of tokens in the completion.
	CompletionTokens int

	// TotalTokens is the total number of tokens used.
	TotalTokens int
}

// Document represents a document to be indexed or retrieved.
type Document struct {
	// ID is a unique identifier for the document.
	ID string

	// Content is the text content of the document.
	Content string

	// Metadata contains arbitrary metadata about the document.
	Metadata map[string]any
}

// Chunk represents a split portion of a Document.
type Chunk struct {
	// ID is a unique identifier for the chunk.
	ID string

	// Content is the text content of the chunk.
	Content string

	// Metadata contains arbitrary metadata about the chunk.
	Metadata map[string]any

	// ParentID is the ID of the parent Document this chunk was split from.
	ParentID string
}

// ScoredDocument is a Document with a relevance score from a vector search.
type ScoredDocument struct {
	Document

	// Score is the relevance score (higher = more relevant).
	// For cosine similarity, range is typically [-1, 1].
	Score float64
}

// JSONSchema represents a JSON Schema definition for tool input validation.
type JSONSchema struct {
	// Type is the root type of the schema (usually "object").
	Type string `json:"type"`

	// Properties defines the properties of the object.
	Properties map[string]*JSONSchemaProperty `json:"properties,omitempty"`

	// Required lists the required property names.
	Required []string `json:"required,omitempty"`

	// Description provides a description of the schema.
	Description string `json:"description,omitempty"`
}

// JSONSchemaProperty represents a property within a JSON Schema.
type JSONSchemaProperty struct {
	// Type is the type of the property.
	Type string `json:"type"`

	// Description provides a description of the property.
	Description string `json:"description,omitempty"`

	// Enum lists allowed values for the property.
	Enum []string `json:"enum,omitempty"`

	// Items defines the schema for array items (when Type is "array").
	Items *JSONSchemaProperty `json:"items,omitempty"`
}
