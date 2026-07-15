// Package ollama implements the Ollama provider for core.
//
// Ollama is a local-first AI platform that runs models on your own machine.
// It provides both chat completion and embedding capabilities without
// requiring API keys or internet access.
//
// Usage:
//
//	import _ "github.com/bachtiarpanjaitan/ihandai-go/plugins/ollama"
//
//	chat, _ := llm.Open("ollama", llm.WithModel("llama3"))
//	embed, _ := embedding.Open("ollama", embedding.WithModel("nomic-embed-text"))
package ollama

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// DefaultBaseURL is the default Ollama server address.
const DefaultBaseURL = "http://localhost:11434"

// HTTPClient is the interface for HTTP clients used by the provider.
// This allows injecting custom HTTP clients for testing.
type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

// Config holds configuration for the Ollama provider.
type Config struct {
	// BaseURL is the Ollama server URL (default: http://localhost:11434).
	BaseURL string

	// Model is the model name (e.g., "llama3", "nomic-embed-text").
	Model string

	// HTTPClient is the HTTP client to use (default: http.DefaultClient).
	HTTPClient HTTPClient
}

// Option is a functional option for configuring the Ollama provider.
type Option func(*Config)

// WithBaseURL sets the server URL.
func WithBaseURL(url string) Option {
	return func(c *Config) { c.BaseURL = url }
}

// WithModel sets the model name.
func WithModel(model string) Option {
	return func(c *Config) { c.Model = model }
}

// WithHTTPClient sets a custom HTTP client.
func WithHTTPClient(client HTTPClient) Option {
	return func(c *Config) { c.HTTPClient = client }
}

func defaultConfig() Config {
	return Config{
		BaseURL:    DefaultBaseURL,
		HTTPClient: &http.Client{Timeout: 30 * time.Second},
	}
}

func applyOptions(opts []Option) Config {
	cfg := defaultConfig()
	for _, opt := range opts {
		opt(&cfg)
	}
	return cfg
}

// buildURL joins a base URL with an endpoint path, normalizing common Ollama
// path segments that users may accidentally include in their base_url setting.
//
// Examples for baseURL "http://localhost:11434/api" + endpoint "/api/chat":
//
//	→ "http://localhost:11434/api/chat" (not "http://localhost:11434/api/api/chat")
//
// Supported baseURL forms (all produce the correct result):
//
//	http://localhost:11434             → http://localhost:11434/api/chat
//	http://localhost:11434/            → http://localhost:11434/api/chat
//	http://localhost:11434/api         → http://localhost:11434/api/chat
//	http://localhost:11434/api/chat    → http://localhost:11434/api/chat
func buildURL(baseURL, endpoint string) string {
	baseURL = strings.TrimSpace(baseURL)

	u, err := url.Parse(baseURL)
	if err != nil || u.Host == "" {
		// Fallback for malformed URLs: clean slashes and join
		return strings.TrimRight(baseURL, "/") + endpoint
	}

	// Strip known Ollama API path suffixes so users can set base_url to any
	// valid Ollama endpoint without worrying about the exact format.
	// Longest suffixes first to avoid partial matches.
	u.Path = strings.TrimSuffix(u.Path, "/api/chat")
	u.Path = strings.TrimSuffix(u.Path, "/api/embeddings")
	u.Path = strings.TrimSuffix(u.Path, "/api")
	u.Path = strings.TrimSuffix(u.Path, "/")

	// Build the final path
	u.Path += endpoint
	u.RawPath = ""

	return u.String()
}

// chatRequest is the JSON body for POST /api/chat.
type chatRequest struct {
	Model    string        `json:"model"`
	Messages []chatMessage `json:"messages"`
	Stream   bool          `json:"stream"`
}

type chatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// chatResponse is the JSON response from POST /api/chat.
type chatResponse struct {
	Model      string      `json:"model"`
	Message    chatMessage `json:"message"`
	Done       bool        `json:"done"`
	DoneReason string      `json:"done_reason,omitempty"`
}

// chatStreamChunk is a single chunk from a streaming POST /api/chat response.
// Each line is a self-contained JSON object with delta content.
type chatStreamChunk struct {
	Model     string      `json:"model"`
	Message   chatMessage `json:"message"`
	Done      bool        `json:"done"`
	DoneReason string     `json:"done_reason,omitempty"`
}

// embedRequest is the JSON body for POST /api/embeddings.
type embedRequest struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
}

// embedResponse is the JSON response from POST /api/embeddings.
type embedResponse struct {
	Embedding []float64 `json:"embedding"`
}

func doRequest(ctx context.Context, client HTTPClient, url string, body any, result any) error {
	bodyBytes, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("ollama: marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(bodyBytes))
	if err != nil {
		return fmt.Errorf("ollama: create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("ollama: request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("ollama: read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("ollama: HTTP %d: %s", resp.StatusCode, string(respBody))
	}

	if err := json.Unmarshal(respBody, result); err != nil {
		return fmt.Errorf("ollama: unmarshal response: %w", err)
	}

	return nil
}
