package ihandai

import (
	"encoding/json"
	"testing"
)

func TestMessage_JSON(t *testing.T) {
	msg := Message{
		Role:    "user",
		Content: "Hello, world!",
		Name:    "test-user",
	}

	data, err := json.Marshal(msg)
	if err != nil {
		t.Fatalf("failed to marshal Message: %v", err)
	}

	var decoded Message
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to unmarshal Message: %v", err)
	}

	if decoded.Role != msg.Role {
		t.Errorf("Role: got %q, want %q", decoded.Role, msg.Role)
	}
	if decoded.Content != msg.Content {
		t.Errorf("Content: got %q, want %q", decoded.Content, msg.Content)
	}
	if decoded.Name != msg.Name {
		t.Errorf("Name: got %q, want %q", decoded.Name, msg.Name)
	}
}

func TestResponse_JSON(t *testing.T) {
	resp := Response{
		Content:      "Hello!",
		FinishReason: "stop",
		ToolCalls: []ToolCall{
			{ID: "1", Name: "search", Arguments: []byte(`{"query":"test"}`)},
		},
		Usage: &TokenUsage{
			PromptTokens:     10,
			CompletionTokens: 5,
			TotalTokens:      15,
		},
	}

	data, err := json.Marshal(resp)
	if err != nil {
		t.Fatalf("failed to marshal Response: %v", err)
	}

	var decoded Response
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to unmarshal Response: %v", err)
	}

	if decoded.Content != resp.Content {
		t.Errorf("Content: got %q, want %q", decoded.Content, resp.Content)
	}
	if decoded.FinishReason != resp.FinishReason {
		t.Errorf("FinishReason: got %q, want %q", decoded.FinishReason, resp.FinishReason)
	}
	if len(decoded.ToolCalls) != 1 {
		t.Errorf("ToolCalls len: got %d, want 1", len(decoded.ToolCalls))
	}
	if decoded.Usage.TotalTokens != 15 {
		t.Errorf("TotalTokens: got %d, want 15", decoded.Usage.TotalTokens)
	}
}

func TestDocument(t *testing.T) {
	doc := Document{
		ID:      "doc-1",
		Content: "Document content",
		Metadata: map[string]any{
			"source": "test.txt",
			"page":   1,
		},
	}

	if doc.ID != "doc-1" {
		t.Errorf("ID: got %q, want %q", doc.ID, "doc-1")
	}
	if doc.Content != "Document content" {
		t.Errorf("Content: got %q, want %q", doc.Content, "Document content")
	}
	if doc.Metadata["source"] != "test.txt" {
		t.Errorf("Metadata[source]: got %q, want %q", doc.Metadata["source"], "test.txt")
	}
}

func TestChunk(t *testing.T) {
	chunk := Chunk{
		ID:       "chunk-1",
		Content:  "Chunk content",
		ParentID: "doc-1",
		Metadata: map[string]any{
			"chunk_index": 0,
		},
	}

	if chunk.ID != "chunk-1" {
		t.Errorf("ID: got %q, want %q", chunk.ID, "chunk-1")
	}
	if chunk.Content != "Chunk content" {
		t.Errorf("Content: got %q, want %q", chunk.Content, "Chunk content")
	}
	if chunk.ParentID != "doc-1" {
		t.Errorf("ParentID: got %q, want %q", chunk.ParentID, "doc-1")
	}
	if chunk.Metadata["chunk_index"] != 0 {
		t.Errorf("Metadata[chunk_index]: got %v, want 0", chunk.Metadata["chunk_index"])
	}
}

func TestScoredDocument(t *testing.T) {
	doc := ScoredDocument{
		Document: Document{
			ID:      "doc-1",
			Content: "Content",
		},
		Score: 0.95,
	}

	if doc.Score != 0.95 {
		t.Errorf("Score: got %f, want 0.95", doc.Score)
	}
	if doc.ID != "doc-1" {
		t.Errorf("embedded Document ID should be accessible: got %q", doc.ID)
	}
}

func TestJSONSchema(t *testing.T) {
	schema := JSONSchema{
		Type: "object",
		Properties: map[string]*JSONSchemaProperty{
			"name": {
				Type:        "string",
				Description: "The name of the person",
			},
			"age": {
				Type: "integer",
			},
		},
		Required:    []string{"name"},
		Description: "A person object",
	}

	data, err := json.Marshal(schema)
	if err != nil {
		t.Fatalf("failed to marshal JSONSchema: %v", err)
	}

	var decoded JSONSchema
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to unmarshal JSONSchema: %v", err)
	}

	if decoded.Type != "object" {
		t.Errorf("Type: got %q, want %q", decoded.Type, "object")
	}
	if len(decoded.Required) != 1 || decoded.Required[0] != "name" {
		t.Errorf("Required: got %v, want [name]", decoded.Required)
	}
	if decoded.Properties["name"].Type != "string" {
		t.Errorf("Properties[name].Type: got %q, want %q", decoded.Properties["name"].Type, "string")
	}
}

func TestToolCall(t *testing.T) {
	tc := ToolCall{
		ID:        "call-1",
		Name:      "search",
		Arguments: []byte(`{"query":"test"}`),
	}

	if tc.ID != "call-1" {
		t.Errorf("ID: got %q, want %q", tc.ID, "call-1")
	}
	if tc.Name != "search" {
		t.Errorf("Name: got %q, want %q", tc.Name, "search")
	}
	if string(tc.Arguments) != `{"query":"test"}` {
		t.Errorf("Arguments: got %q, want %q", string(tc.Arguments), `{"query":"test"}`)
	}
}

func TestTokenUsage(t *testing.T) {
	usage := TokenUsage{
		PromptTokens:     100,
		CompletionTokens: 50,
		TotalTokens:      150,
	}

	if usage.TotalTokens != usage.PromptTokens+usage.CompletionTokens {
		t.Error("TotalTokens should equal PromptTokens + CompletionTokens")
	}
}
