package telemetry

import (
	"context"
	"log/slog"
	"os"
	"testing"
	"time"
)

func TestRateLimiter_Allow(t *testing.T) {
	rl := NewRateLimiter(10, 5) // 10 tokens/sec, burst 5

	// Initial burst
	for range 5 {
		if !rl.Allow() {
			t.Error("expected burst of 5 to be allowed")
		}
	}
	// 6th should be denied
	if rl.Allow() {
		t.Error("expected 6th request to be denied")
	}
}

func TestRateLimiter_Refill(t *testing.T) {
	rl := NewRateLimiter(100, 1)
	rl.Allow() // consume the burst

	if rl.Allow() {
		t.Error("should be empty")
	}

	time.Sleep(15 * time.Millisecond) // ~1.5 tokens refilled
	if !rl.Allow() {
		t.Error("should have refilled enough")
	}
}

func TestRateLimiter_Wait(t *testing.T) {
	rl := NewRateLimiter(100, 1)
	rl.Allow()

	if rl.Wait(200 * time.Millisecond) {
		t.Log("wait succeeded")
	} else {
		t.Error("wait should succeed within timeout")
	}
}

func TestCircuitBreaker_ClosedToOpen(t *testing.T) {
	cb := NewCircuitBreaker(3, 100*time.Millisecond)

	if cb.State() != CircuitClosed {
		t.Error("initial state should be closed")
	}

	// Fail 3 times
	cb.Failure()
	cb.Failure()
	if cb.State() != CircuitClosed {
		t.Error("should still be closed after 2 failures")
	}

	cb.Failure() // 3rd failure → open
	if cb.State() != CircuitOpen {
		t.Error("should be open after 3 failures")
	}

	if cb.Allow() {
		t.Error("should not allow in open state")
	}
}

func TestCircuitBreaker_HalfOpen(t *testing.T) {
	cb := NewCircuitBreaker(2, 10*time.Millisecond)

	// Trigger open
	cb.Failure()
	cb.Failure()

	if cb.State() != CircuitOpen {
		t.Fatal("should be open")
	}

	time.Sleep(15 * time.Millisecond)

	if !cb.Allow() {
		t.Error("should allow probe in half-open state")
	}
	if cb.State() != CircuitHalfOpen {
		t.Error("should be half-open")
	}

	// Probe succeeds
	cb.Success()
	if cb.State() != CircuitClosed {
		t.Error("should return to closed after success")
	}
}

func TestCircuitBreaker_Reset(t *testing.T) {
	cb := NewCircuitBreaker(1, time.Minute)
	cb.Failure() // open
	cb.Reset()
	if cb.State() != CircuitClosed {
		t.Error("reset should close the circuit")
	}
}

func TestTracer(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))
	tracer := NewTracer(logger)
	ctx, span := tracer.Start(context.TODO(), "test-span", slog.String("key", "val"))
	if ctx == nil {
		t.Error("context should not be nil")
	}
	span.End(nil)
	t.Log("span completed")
}
