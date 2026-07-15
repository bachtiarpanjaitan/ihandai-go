package ollama

import (
	"context"
	"fmt"

	"github.com/bachtiarpanjaitan/ihandai-go/pkg/core"
	"github.com/bachtiarpanjaitan/ihandai-go/pkg/llm"
)

// Compile-time interface satisfaction check.
var _ llm.ChatCompleter = (*ChatCompleter)(nil)

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
