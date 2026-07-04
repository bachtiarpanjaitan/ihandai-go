// Package telemetry provides production observability utilities:
// tracing, metrics, rate limiting, and circuit breaking.
package telemetry

import (
	"context"
	"log/slog"
	"time"
)

// Tracer provides lightweight span-based tracing without external dependencies.
// For production, replace with OpenTelemetry integration.
type Tracer struct {
	logger *slog.Logger
}

// Span represents a unit of work in a trace.
type Span struct {
	name      string
	startTime time.Time
	logger    *slog.Logger
	attrs     []slog.Attr
}

// NewTracer creates a new tracer with the given logger.
func NewTracer(logger *slog.Logger) *Tracer {
	if logger == nil {
		logger = slog.Default()
	}
	return &Tracer{logger: logger}
}

// Start begins a new span. The context should be used for child spans.
func (t *Tracer) Start(ctx context.Context, name string, attrs ...slog.Attr) (context.Context, *Span) {
	span := &Span{
		name:      name,
		startTime: time.Now(),
		logger:    t.logger,
		attrs:     attrs,
	}
	return ctx, span
}

// End completes the span and logs its duration.
func (s *Span) End(err error) {
	duration := time.Since(s.startTime)
	attrs := append(s.attrs,
		slog.String("span", s.name),
		slog.Duration("duration_ms", duration),
	)
	if err != nil {
		attrs = append(attrs, slog.String("error", err.Error()))
		s.logger.LogAttrs(context.Background(), slog.LevelError, "span completed", attrs...)
	} else {
		s.logger.LogAttrs(context.Background(), slog.LevelDebug, "span completed", attrs...)
	}
}
