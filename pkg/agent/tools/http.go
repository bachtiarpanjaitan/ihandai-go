package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/bachtiarpanjaitan/ihandai-go/pkg/core"
)

// HTTPRequest is a tool that makes HTTP GET requests.
type HTTPRequest struct {
	client *http.Client
}

// NewHTTPRequest creates a new HTTP request tool.
func NewHTTPRequest() *HTTPRequest {
	return &HTTPRequest{client: &http.Client{Timeout: 10 * time.Second}}
}

func (h *HTTPRequest) Name() string { return "http_get" }
func (h *HTTPRequest) Description() string {
	return "Make an HTTP GET request to fetch content from a URL."
}
func (h *HTTPRequest) InputSchema() *core.JSONSchema {
	return &core.JSONSchema{
		Type: "object",
		Properties: map[string]*core.JSONSchemaProp{
			"url": {Type: "string", Description: "The URL to fetch"},
		},
		Required: []string{"url"},
	}
}

func (h *HTTPRequest) Execute(ctx context.Context, input json.RawMessage) (json.RawMessage, error) {
	var params struct {
		URL string `json:"url"`
	}
	if err := json.Unmarshal(input, &params); err != nil {
		return nil, fmt.Errorf("http_get: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, params.URL, nil)
	if err != nil {
		return nil, fmt.Errorf("http_get: %w", err)
	}

	resp, err := h.client.Do(req)
	if err != nil {
		return json.RawMessage(fmt.Sprintf(`{"error":"%s"}`, err.Error())), nil
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 100*1024)) // 100KB limit
	if err != nil {
		return nil, fmt.Errorf("http_get: read: %w", err)
	}

	// Truncate response for readability
	content := string(body)
	if len(content) > 5000 {
		content = content[:5000] + "..."
	}
	content = strings.ReplaceAll(content, `"`, `\"`)
	content = strings.ReplaceAll(content, "\n", "\\n")

	return json.RawMessage(fmt.Sprintf(`{"status":%d,"content":"%s"}`, resp.StatusCode, content)), nil
}
