// Package ihandai is a provider-agnostic AI library for Go.
// It provides composable interfaces for LLM, embedding, vector stores,
// and orchestrates them into RAG pipelines.
package ihandai

import "github.com/bachtiarpanjaitan/ihandai-go/pkg/core"

// Re-export shared types from pkg/core for convenience.
// Users can either use ihandai.Message or core.Message.

// Message represents a chat message.
type Message = core.Message

// Response represents an LLM chat response.
type Response = core.Response

// ToolCall represents a tool call requested by an LLM.
type ToolCall = core.ToolCall

// TokenUsage contains token usage information.
type TokenUsage = core.TokenUsage

// Document represents a document to be indexed or retrieved.
type Document = core.Document

// Chunk represents a split portion of a Document.
type Chunk = core.Chunk

// ScoredDocument is a Document with a relevance score.
type ScoredDocument = core.ScoredDocument

// JSONSchema represents a JSON Schema definition.
type JSONSchema = core.JSONSchema

// JSONSchemaProp represents a property within a JSON Schema.
type JSONSchemaProp = core.JSONSchemaProp
