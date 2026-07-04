package memory

import (
	"context"
	"sync"

	"github.com/bachtiarpanjaitan/ihandai-go/pkg/core"
)

// InMemoryStore is a thread-safe, in-memory conversation store.
// Conversations are lost when the process exits. Use for development and testing.
type InMemoryStore struct {
	mu     sync.RWMutex
	convos map[string][]core.Message
}

// NewInMemoryStore creates a new in-memory conversation store.
func NewInMemoryStore() *InMemoryStore {
	return &InMemoryStore{convos: make(map[string][]core.Message)}
}

// Append adds a message to the conversation.
func (s *InMemoryStore) Append(ctx context.Context, key string, msg core.Message) error {
	_ = ctx
	s.mu.Lock()
	defer s.mu.Unlock()
	s.convos[key] = append(s.convos[key], msg)
	return nil
}

// History returns all messages in the conversation, oldest first.
// Returns an empty slice (not nil) if the key does not exist.
func (s *InMemoryStore) History(ctx context.Context, key string) ([]core.Message, error) {
	_ = ctx
	s.mu.RLock()
	defer s.mu.RUnlock()
	msgs := s.convos[key]
	if msgs == nil {
		return []core.Message{}, nil
	}
	// Return a copy
	result := make([]core.Message, len(msgs))
	copy(result, msgs)
	return result, nil
}

// Trim removes oldest messages until total tokens ≤ maxTokens.
// The system message (role="system") is always preserved.
func (s *InMemoryStore) Trim(ctx context.Context, key string, maxTokens int, counter TokenCounter) error {
	_ = ctx
	s.mu.Lock()
	defer s.mu.Unlock()

	msgs := s.convos[key]
	if len(msgs) <= 1 {
		return nil
	}

	// Separate system message(s) from the rest
	var system []core.Message
	var rest []core.Message
	for _, m := range msgs {
		if m.Role == "system" {
			system = append(system, m)
		} else {
			rest = append(rest, m)
		}
	}

	// Trim from the front (oldest first), keeping newest messages
	for len(rest) > 0 {
		total := estimateTokens(system, rest, counter)
		if total <= maxTokens {
			break
		}
		rest = rest[1:]
	}

	s.convos[key] = append(system, rest...)
	return nil
}

// Clear removes all messages for the key.
func (s *InMemoryStore) Clear(ctx context.Context, key string) error {
	_ = ctx
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.convos, key)
	return nil
}

// Len returns the number of conversations in the store.
func (s *InMemoryStore) Len() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.convos)
}
