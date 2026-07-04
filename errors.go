package ihandai

import (
	"fmt"
	"time"
)

// RateLimitError indicates that a provider returned a rate limit response.
// Callers can inspect RetryAfter to determine when to retry.
type RateLimitError struct {
	// Provider is the name of the provider that rate-limited the request.
	Provider string

	// RetryAfter is the duration to wait before retrying.
	RetryAfter time.Duration
}

// Error implements the error interface.
func (e *RateLimitError) Error() string {
	return fmt.Sprintf("ihandai: rate limited by %s (retry after %v)", e.Provider, e.RetryAfter)
}

// AuthError indicates that authentication with a provider failed.
// This typically means the API key is invalid, expired, or missing.
type AuthError struct {
	// Provider is the name of the provider that rejected the authentication.
	Provider string
}

// Error implements the error interface.
func (e *AuthError) Error() string {
	return fmt.Sprintf("ihandai: authentication failed for %s (check your API key)", e.Provider)
}

// TimeoutError indicates that a request to a provider timed out.
type TimeoutError struct {
	// Provider is the name of the provider that timed out.
	Provider string

	// Duration is how long the request waited before timing out.
	Duration time.Duration
}

// Error implements the error interface.
func (e *TimeoutError) Error() string {
	return fmt.Sprintf("ihandai: request to %s timed out after %v", e.Provider, e.Duration)
}

// ProviderError indicates a generic error returned by a provider's API.
// This covers HTTP errors, unexpected response formats, etc.
type ProviderError struct {
	// Provider is the name of the provider that returned the error.
	Provider string

	// StatusCode is the HTTP status code, or 0 if not applicable.
	StatusCode int

	// Body is the response body, if available.
	Body string
}

// Error implements the error interface.
func (e *ProviderError) Error() string {
	if e.StatusCode > 0 {
		return fmt.Sprintf("ihandai: %s returned HTTP %d: %s", e.Provider, e.StatusCode, e.Body)
	}
	return fmt.Sprintf("ihandai: %s error: %s", e.Provider, e.Body)
}

// PipelineError wraps an error that occurred at a specific step in a pipeline.
type PipelineError struct {
	// Step is the name of the pipeline step that failed.
	Step string

	// Err is the underlying error.
	Err error
}

// Error implements the error interface.
func (e *PipelineError) Error() string {
	return fmt.Sprintf("ihandai: pipeline step %q failed: %v", e.Step, e.Err)
}

// Unwrap returns the underlying error for errors.Is / errors.As support.
func (e *PipelineError) Unwrap() error {
	return e.Err
}
