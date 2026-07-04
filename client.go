package ihandai

import (
	"context"
	"log/slog"
)

// Config holds all configuration for the Client.
// It is created by New() and populated via functional options.
type Config struct {
	// Logger is the structured logger for the client and all providers.
	Logger *slog.Logger
}

// Option is a functional option for configuring the Client.
// Options are applied in order to a Config during New().
type Option func(*Config)

// Client is the main entry point for the ihandai library.
//
// Create one Client per application and reuse it across goroutines.
// The Client is immutable after creation and safe for concurrent use.
//
//	ai, err := ihandai.New(
//	    ihandai.WithLogger(slog.Default()),
//	)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	defer ai.Close()
type Client struct {
	cfg *Config
}

// New creates a new Client with the given options.
//
// Options are applied in order. If no logger is provided,
// slog.Default() is used.
func New(opts ...Option) (*Client, error) {
	cfg := &Config{
		Logger: slog.Default(),
	}
	for _, opt := range opts {
		opt(cfg)
	}
	return &Client{cfg: cfg}, nil
}

// Close releases any resources held by the Client and its providers.
// After Close, the Client must not be used.
//
// It is safe to call Close multiple times.
func (c *Client) Close() error {
	return nil
}

// Ask sends a query through the RAG pipeline and returns the LLM response.
// This is a placeholder that will be implemented in Phase 4 (Pipeline).
func (c *Client) Ask(ctx context.Context, query string) (*Response, error) {
	_ = ctx
	_ = query
	return nil, nil
}

// Index indexes documents into the vector store for later retrieval.
// This is a placeholder that will be implemented in Phase 4 (Pipeline).
func (c *Client) Index(ctx context.Context, documents []Document) error {
	_ = ctx
	_ = documents
	return nil
}

// Config returns a copy of the client's configuration.
// This is useful for debugging and testing.
func (c *Client) Config() *Config {
	cp := *c.cfg
	return &cp
}

// WithLogger sets the structured logger for the client.
// If not set, slog.Default() is used.
func WithLogger(logger *slog.Logger) Option {
	return func(c *Config) {
		c.Logger = logger
	}
}
