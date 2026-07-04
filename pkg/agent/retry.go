package agent

import (
	"context"
	"math"
	"time"

	"github.com/bachtiarpanjaitan/ihandai-go/pkg/core"
	"github.com/bachtiarpanjaitan/ihandai-go/pkg/llm"
)

// RetryPolicy defines how LLM requests are retried on failure.
type RetryPolicy struct {
	MaxRetries        int
	InitialBackoff    time.Duration
	MaxBackoff        time.Duration
	BackoffMultiplier float64
}

// DefaultRetryPolicy returns a sensible default.
func DefaultRetryPolicy() RetryPolicy {
	return RetryPolicy{
		MaxRetries:        3,
		InitialBackoff:    1 * time.Second,
		MaxBackoff:        30 * time.Second,
		BackoffMultiplier: 2.0,
	}
}

// RetryChat wraps a ChatCompleter with retry logic.
func RetryChat(ctx context.Context, chat llm.ChatCompleter, policy RetryPolicy, messages []core.Message) (*core.Response, error) {
	var lastErr error
	for attempt := 0; attempt <= policy.MaxRetries; attempt++ {
		if attempt > 0 {
			delay := time.Duration(float64(policy.InitialBackoff) * math.Pow(policy.BackoffMultiplier, float64(attempt-1)))
			if delay > policy.MaxBackoff {
				delay = policy.MaxBackoff
			}
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(delay):
			}
		}

		resp, err := chat.Chat(ctx, messages)
		if err == nil {
			return resp, nil
		}
		lastErr = err
	}
	return nil, lastErr
}
