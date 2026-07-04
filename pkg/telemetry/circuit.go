package telemetry

import (
	"sync"
	"time"
)

// CircuitState represents the state of a circuit breaker.
type CircuitState int

const (
	// CircuitClosed allows requests (normal operation).
	CircuitClosed CircuitState = iota
	// CircuitOpen rejects requests (failure detected).
	CircuitOpen
	// CircuitHalfOpen allows a single probe request.
	CircuitHalfOpen
)

// String returns a human-readable state name.
func (s CircuitState) String() string {
	switch s {
	case CircuitClosed:
		return "closed"
	case CircuitOpen:
		return "open"
	case CircuitHalfOpen:
		return "half-open"
	default:
		return "unknown"
	}
}

// CircuitBreaker protects a service from cascading failures.
// When consecutive failures exceed the threshold, the circuit opens
// and fast-fails requests for the timeout duration.
type CircuitBreaker struct {
	failureThreshold int
	resetTimeout     time.Duration

	mu              sync.Mutex
	state           CircuitState
	failureCount    int
	lastFailure     time.Time
	lastStateChange time.Time
}

// NewCircuitBreaker creates a new circuit breaker.
// threshold: consecutive failures before opening (default: 5)
// timeout: how long to stay open before half-open probe (default: 30s)
func NewCircuitBreaker(threshold int, timeout time.Duration) *CircuitBreaker {
	if threshold <= 0 {
		threshold = 5
	}
	if timeout <= 0 {
		timeout = 30 * time.Second
	}
	return &CircuitBreaker{
		failureThreshold: threshold,
		resetTimeout:     timeout,
		state:            CircuitClosed,
	}
}

// Allow returns true if the request should be attempted.
func (cb *CircuitBreaker) Allow() bool {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	switch cb.state {
	case CircuitClosed:
		return true
	case CircuitOpen:
		if time.Since(cb.lastStateChange) > cb.resetTimeout {
			cb.state = CircuitHalfOpen
			cb.lastStateChange = time.Now()
			return true // allow probe
		}
		return false
	case CircuitHalfOpen:
		return true // allow probe
	default:
		return false
	}
}

// Success reports a successful request.
func (cb *CircuitBreaker) Success() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.failureCount = 0
	if cb.state == CircuitHalfOpen {
		cb.state = CircuitClosed
		cb.lastStateChange = time.Now()
	}
}

// Failure reports a failed request.
func (cb *CircuitBreaker) Failure() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.failureCount++
	cb.lastFailure = time.Now()

	if cb.state == CircuitHalfOpen || cb.failureCount >= cb.failureThreshold {
		cb.state = CircuitOpen
		cb.lastStateChange = time.Now()
	}
}

// State returns the current circuit state.
func (cb *CircuitBreaker) State() CircuitState {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	return cb.state
}

// Reset resets the circuit breaker to closed state.
func (cb *CircuitBreaker) Reset() {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	cb.state = CircuitClosed
	cb.failureCount = 0
}
