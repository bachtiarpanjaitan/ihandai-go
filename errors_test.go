package ihandai

import (
	"errors"
	"testing"
	"time"
)

func TestRateLimitError(t *testing.T) {
	err := &RateLimitError{
		Provider:   "openai",
		RetryAfter: 30 * time.Second,
	}

	want := "ihandai: rate limited by openai (retry after 30s)"
	if err.Error() != want {
		t.Errorf("got %q, want %q", err.Error(), want)
	}

	var target *RateLimitError
	if !errors.As(err, &target) {
		t.Error("errors.As should match RateLimitError")
	}
}

func TestAuthError(t *testing.T) {
	err := &AuthError{Provider: "anthropic"}

	want := "ihandai: authentication failed for anthropic (check your API key)"
	if err.Error() != want {
		t.Errorf("got %q, want %q", err.Error(), want)
	}

	var target *AuthError
	if !errors.As(err, &target) {
		t.Error("errors.As should match AuthError")
	}
}

func TestTimeoutError(t *testing.T) {
	err := &TimeoutError{
		Provider: "openai",
		Duration: 3 * time.Second,
	}

	want := "ihandai: request to openai timed out after 3s"
	if err.Error() != want {
		t.Errorf("got %q, want %q", err.Error(), want)
	}

	var target *TimeoutError
	if !errors.As(err, &target) {
		t.Error("errors.As should match TimeoutError")
	}
}

func TestProviderError(t *testing.T) {
	tests := []struct {
		name       string
		err        *ProviderError
		wantSubstr string
	}{
		{
			name: "with status code",
			err: &ProviderError{
				Provider:   "openai",
				StatusCode: 500,
				Body:       "internal server error",
			},
			wantSubstr: "HTTP 500",
		},
		{
			name: "without status code",
			err: &ProviderError{
				Provider: "ollama",
				Body:     "connection refused",
			},
			wantSubstr: "connection refused",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.err.Error()
			if !contains(got, tt.wantSubstr) {
				t.Errorf("got %q, want to contain %q", got, tt.wantSubstr)
			}
		})
	}

	var target *ProviderError
	if !errors.As(&ProviderError{Provider: "x", StatusCode: 400, Body: "bad request"}, &target) {
		t.Error("errors.As should match ProviderError")
	}
}

func TestPipelineError(t *testing.T) {
	underlying := errors.New("connection refused")
	err := &PipelineError{
		Step: "embedding",
		Err:  underlying,
	}

	want := `ihandai: pipeline step "embedding" failed: connection refused`
	if err.Error() != want {
		t.Errorf("got %q, want %q", err.Error(), want)
	}

	// Test errors.Is chain
	if !errors.Is(err, underlying) {
		t.Error("errors.Is should find the underlying error")
	}

	// Test Unwrap
	if err.Unwrap() != underlying {
		t.Error("Unwrap should return the underlying error")
	}
}

func TestErrorTypes_AreDistinct(t *testing.T) {
	// Verify each error type is distinct and not caught by the wrong type
	rl := &RateLimitError{Provider: "x", RetryAfter: time.Second}
	auth := &AuthError{Provider: "x"}
	timeout := &TimeoutError{Provider: "x", Duration: time.Second}
	provider := &ProviderError{Provider: "x", StatusCode: 400, Body: "x"}

	var targetAuth *AuthError
	if errors.As(rl, &targetAuth) {
		t.Error("RateLimitError should not match AuthError")
	}

	var targetRL *RateLimitError
	if errors.As(auth, &targetRL) {
		t.Error("AuthError should not match RateLimitError")
	}

	var targetProv *ProviderError
	if errors.As(timeout, &targetProv) {
		t.Error("TimeoutError should not match ProviderError")
	}

	var targetTO *TimeoutError
	if errors.As(provider, &targetTO) {
		t.Error("ProviderError should not match TimeoutError")
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > 0 && len(substr) > 0 && indexSubstr(s, substr) >= 0))
}

func indexSubstr(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}
