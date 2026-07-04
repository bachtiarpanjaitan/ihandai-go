package prompt

import (
	"context"
	"fmt"
	"strings"

	"github.com/bachtiarpanjaitan/ihandai-go/pkg/core"
)

// Simple is a basic prompt builder that formats retrieved documents
// and the user query into a standard RAG prompt.
type Simple struct{}

// NewSimple creates a new Simple prompt builder.
func NewSimple() *Simple {
	return &Simple{}
}

// Build implements PromptBuilder.
// It formats documents as numbered items and includes the query.
func (s *Simple) Build(ctx context.Context, template string, contextData map[string]any) ([]core.Message, error) {
	_ = ctx

	systemPrompt := template
	if systemPrompt == "" {
		systemPrompt = "You are a helpful assistant. Answer the user's question based on the provided context."
	}

	query, _ := contextData["query"].(string)

	var contextBuilder strings.Builder
	if docs, ok := contextData["documents"].([]core.ScoredDocument); ok && len(docs) > 0 {
		contextBuilder.WriteString("Context:\n")
		for i, doc := range docs {
			contextBuilder.WriteString(fmt.Sprintf("[%d] %s\n", i+1, doc.Content))
		}
		contextBuilder.WriteString("\n")
	}

	userContent := contextBuilder.String()
	if query != "" {
		userContent += fmt.Sprintf("Question: %s", query)
	}

	return []core.Message{
		{Role: "system", Content: systemPrompt},
		{Role: "user", Content: userContent},
	}, nil
}
