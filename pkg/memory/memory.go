// Package memory defines interfaces and implementations for conversation memory.
//
// Conversation stores maintain chat history across multiple interactions,
// enabling multi-turn conversations with context retention.
package memory

import (
	"context"

	"github.com/bachtiarpanjaitan/ihandai-go/pkg/core"
)

// TokenCounter estimates token counts for text.
// This matches llm.TokenCounter without importing pkg/llm.
type TokenCounter interface {
	CountTokens(ctx context.Context, model, text string) (int, error)
}

// ConversationStore persists chat messages across interactions.
// Each conversation is identified by a key (e.g., user ID, session ID).
// Implementations must be safe for concurrent use.
type ConversationStore interface {
	// Append adds a message to the conversation identified by key.
	Append(ctx context.Context, key string, msg core.Message) error

	// History returns all messages in the conversation, oldest first.
	History(ctx context.Context, key string) ([]core.Message, error)

	// Trim removes the oldest messages until the total token count
	// is at or below maxTokens. The system message (if any) is preserved.
	// The counter is used to estimate tokens; if nil, a simple character-based
	// heuristic is used.
	Trim(ctx context.Context, key string, maxTokens int, counter TokenCounter) error

	// Clear removes all messages for the given key.
	Clear(ctx context.Context, key string) error
}
