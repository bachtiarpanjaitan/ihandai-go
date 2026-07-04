package memory

import (
	"context"

	"github.com/bachtiarpanjaitan/ihandai-go/pkg/core"
)

// WindowManager provides context window management for conversations.
// It estimates token usage and can trim history to fit within model limits.
type WindowManager struct {
	store       ConversationStore
	maxTokens   int
	model       string
	counter     TokenCounter
}

// NewWindowManager creates a WindowManager for the given store and model.
// maxTokens is the target maximum token count. For most models, 4096-128000.
func NewWindowManager(store ConversationStore, model string, maxTokens int, counter TokenCounter) *WindowManager {
	return &WindowManager{
		store:     store,
		maxTokens: maxTokens,
		model:     model,
		counter:   counter,
	}
}

// Fit ensures the conversation fits within the token limit.
// It trims the conversation if it exceeds maxTokens, and returns the
// updated message list that fits.
func (w *WindowManager) Fit(ctx context.Context, key string, newMsg core.Message) ([]core.Message, error) {
	// Append the new message first
	if err := w.store.Append(ctx, key, newMsg); err != nil {
		return nil, err
	}

	// Trim if necessary
	if err := w.store.Trim(ctx, key, w.maxTokens, w.counter); err != nil {
		return nil, err
	}

	return w.store.History(ctx, key)
}

// MaxTokens returns the configured maximum token count.
func (w *WindowManager) MaxTokens() int { return w.maxTokens }

// estimateTokens computes total tokens for a set of messages.
// Uses the counter if provided, otherwise falls back to character-based heuristic
// (4 chars ≈ 1 token for English text).
func estimateTokens(system, rest []core.Message, counter TokenCounter) int {
	total := 0
	all := append(system, rest...)
	for _, m := range all {
		if counter != nil {
			n, err := counter.CountTokens(context.TODO(), "", m.Content)
			if err == nil {
				total += n
				continue
			}
		}
		// Fallback: ~4 characters per token for English
		total += len(m.Content) / 4
	}
	return total
}
