package telemetry

import (
	"sync"
	"time"
)

// RateLimiter implements a token bucket rate limiter.
// It controls the rate of operations (e.g., API calls to LLM providers).
type RateLimiter struct {
	rate       float64 // tokens per second
	burst      int     // max tokens
	tokens     float64
	lastUpdate time.Time
	mu         sync.Mutex
}

// NewRateLimiter creates a token bucket rate limiter.
// rate: tokens per second
// burst: maximum burst size
func NewRateLimiter(rate float64, burst int) *RateLimiter {
	return &RateLimiter{
		rate:       rate,
		burst:      burst,
		tokens:     float64(burst),
		lastUpdate: time.Now(),
	}
}

// Allow checks if a request is allowed. Returns true if within limits.
func (rl *RateLimiter) Allow() bool {
	return rl.AllowN(1)
}

// AllowN checks if n requests are allowed.
func (rl *RateLimiter) AllowN(n int) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	elapsed := now.Sub(rl.lastUpdate).Seconds()
	rl.tokens += elapsed * rl.rate
	if rl.tokens > float64(rl.burst) {
		rl.tokens = float64(rl.burst)
	}
	rl.lastUpdate = now

	if rl.tokens >= float64(n) {
		rl.tokens -= float64(n)
		return true
	}
	return false
}

// Wait blocks until n tokens are available or the timeout expires.
func (rl *RateLimiter) Wait(timeout time.Duration) bool {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if rl.Allow() {
			return true
		}
		time.Sleep(50 * time.Millisecond)
	}
	return false
}
