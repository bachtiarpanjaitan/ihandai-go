package ollama

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/bachtiarpanjaitan/ihandai-go/pkg/core"
	"github.com/bachtiarpanjaitan/ihandai-go/pkg/llm"
)

// Compile-time interface satisfaction checks.
var (
	_ llm.ChatCompleter   = (*ChatCompleter)(nil)
	_ llm.StreamCompleter = (*ChatCompleter)(nil)
)

// ChatCompleter implements llm.ChatCompleter using the Ollama chat API.
type ChatCompleter struct {
	cfg Config
}

// NewChatCompleter creates a new Ollama ChatCompleter with the given options.
func NewChatCompleter(opts ...Option) *ChatCompleter {
	return &ChatCompleter{cfg: applyOptions(opts)}
}

// Chat sends messages to the Ollama server and returns the response.
func (c *ChatCompleter) Chat(ctx context.Context, messages []core.Message) (*core.Response, error) {
	msgs := make([]chatMessage, len(messages))
	for i, m := range messages {
		msgs[i] = chatMessage{Role: m.Role, Content: m.Content}
	}

	reqBody := chatRequest{
		Model:    c.cfg.Model,
		Messages: msgs,
		Stream:   false,
	}

	var result chatResponse
	url := buildURL(c.cfg.BaseURL, "/api/chat")
	if err := doRequest(ctx, c.cfg.HTTPClient, url, reqBody, &result); err != nil {
		return nil, fmt.Errorf("ollama: chat: %w", err)
	}

	return &core.Response{
		Content:      result.Message.Content,
		FinishReason: result.DoneReason,
	}, nil
}

// ChatStream sends messages to the Ollama server with streaming enabled.
// Tokens are pushed to the returned channel as they arrive.
func (c *ChatCompleter) ChatStream(ctx context.Context, messages []core.Message) (<-chan llm.Chunk, error) {
	msgs := make([]chatMessage, len(messages))
	for i, m := range messages {
		msgs[i] = chatMessage{Role: m.Role, Content: m.Content}
	}

	reqBody := chatRequest{
		Model:    c.cfg.Model,
		Messages: msgs,
		Stream:   true,
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("ollama: marshal request: %w", err)
	}

	url := buildURL(c.cfg.BaseURL, "/api/chat")
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("ollama: create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.cfg.HTTPClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("ollama: request failed: %w", err)
	}

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		return nil, fmt.Errorf("ollama: HTTP %d: %s", resp.StatusCode, string(body))
	}

	ch := make(chan llm.Chunk)

	go func() {
		defer resp.Body.Close()
		defer close(ch)

		scanner := bufio.NewScanner(resp.Body)
		scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)

		for scanner.Scan() {
			select {
			case <-ctx.Done():
				return
			default:
			}

			line := strings.TrimSpace(scanner.Text())
			if line == "" {
				continue
			}

			var streamChunk chatStreamChunk
			if err := json.Unmarshal([]byte(line), &streamChunk); err != nil {
				continue
			}

			if streamChunk.Message.Content != "" || streamChunk.Done {
				var fr string
				if streamChunk.Done {
					fr = streamChunk.DoneReason
				}
				ch <- llm.Chunk{
					Content:      streamChunk.Message.Content,
					FinishReason: fr,
				}
			}
		}
	}()

	return ch, nil
}

// init registers the Ollama chat provider.
func init() {
	llm.Register("ollama", func(cfg llm.Config) (llm.ChatCompleter, error) {
		opts := []Option{WithModel(cfg.Model)}
		if cfg.BaseURL != "" {
			opts = append(opts, WithBaseURL(cfg.BaseURL))
		}
		return NewChatCompleter(opts...), nil
	})
}
