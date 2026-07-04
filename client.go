package ihandai

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/bachtiarpanjaitan/ihandai-go/pkg/embedding"
	"github.com/bachtiarpanjaitan/ihandai-go/pkg/llm"
	"github.com/bachtiarpanjaitan/ihandai-go/pkg/vectordb"
)

// Config holds all configuration for the Client.
// It is created by New() and populated via functional options.
type Config struct {
	// Logger is the structured logger for the client and all providers.
	Logger *slog.Logger

	// LLM holds the LLM provider configuration.
	LLMProvider string
	LLMOptions  []llm.Option

	// Embedding holds the embedding provider configuration.
	EmbeddingProvider string
	EmbeddingOptions  []embedding.Option

	// IndexEmbedding (optional) holds a separate embedding provider for indexing.
	// When empty, Embedding is used for both indexing and querying.
	IndexEmbeddingProvider string
	IndexEmbeddingOptions  []embedding.Option

	// VectorStore holds the vector store provider configuration.
	VectorStoreProvider string
	VectorStoreOptions  []vectordb.Option
}

// Option is a functional option for configuring the Client.
// Options are applied in order to a Config during New().
type Option func(*Config)

// Client is the main entry point for the ihandai library.
//
// Create one Client per application and reuse it across goroutines.
// The Client is immutable after creation and safe for concurrent use.
//
//	import _ "github.com/bachtiarpanjaitan/ihandai-go/plugins/ollama"
//
//	ai, err := ihandai.New(
//	    ihandai.WithLLM("ollama", llm.WithModel("llama3")),
//	    ihandai.WithEmbedding("ollama", embedding.WithModel("nomic-embed-text")),
//	)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	defer ai.Close()
type Client struct {
	cfg *Config

	llm       llm.ChatCompleter
	embedding embedding.Embedder
	vectordb  vectordb.VectorSearcher
}

// New creates a new Client with the given options.
//
// Options are applied in order. If no logger is provided,
// slog.Default() is used.
//
// Providers specified via WithLLM, WithEmbedding, WithVectorStore
// are opened immediately. If a provider name is not registered
// (via a blank import of its plugin), New returns an error.
func New(opts ...Option) (*Client, error) {
	cfg := &Config{
		Logger: slog.Default(),
	}
	for _, opt := range opts {
		opt(cfg)
	}

	c := &Client{cfg: cfg}

	// Open LLM provider
	if cfg.LLMProvider != "" {
		chat, err := llm.Open(cfg.LLMProvider, cfg.LLMOptions...)
		if err != nil {
			return nil, fmt.Errorf("ihandai: %w", err)
		}
		c.llm = chat
	}

	// Open Embedding provider
	if cfg.EmbeddingProvider != "" {
		embed, err := embedding.Open(cfg.EmbeddingProvider, cfg.EmbeddingOptions...)
		if err != nil {
			return nil, fmt.Errorf("ihandai: %w", err)
		}
		c.embedding = embed
	}

	// Open VectorStore provider
	if cfg.VectorStoreProvider != "" {
		store, err := vectordb.Open(cfg.VectorStoreProvider, cfg.VectorStoreOptions...)
		if err != nil {
			return nil, fmt.Errorf("ihandai: %w", err)
		}
		c.vectordb = store
	}

	return c, nil
}

// Close releases any resources held by the Client and its providers.
// After Close, the Client must not be used.
//
// It is safe to call Close multiple times.
func (c *Client) Close() error {
	return nil
}

// LLM returns the configured LLM provider, or nil if not configured.
func (c *Client) LLM() llm.ChatCompleter {
	return c.llm
}

// Embedding returns the configured embedding provider, or nil if not configured.
func (c *Client) Embedding() embedding.Embedder {
	return c.embedding
}

// VectorStore returns the configured vector store provider, or nil if not configured.
func (c *Client) VectorStore() vectordb.VectorSearcher {
	return c.vectordb
}

// Ask sends a query through the RAG pipeline and returns the LLM response.
// This requires LLM, Embedding, and VectorStore providers to be configured.
// This is a placeholder that will be fully implemented in Phase 4 (Pipeline).
func (c *Client) Ask(ctx context.Context, query string) (*Response, error) {
	if c.llm == nil {
		return nil, fmt.Errorf("ihandai: LLM provider not configured")
	}
	if c.embedding == nil {
		return nil, fmt.Errorf("ihandai: embedding provider not configured")
	}
	if c.vectordb == nil {
		return nil, fmt.Errorf("ihandai: vector store not configured")
	}

	// TODO Phase 4: full pipeline — embed → search → rerank → prompt → chat
	// For now, just call LLM directly with the query
	msgs := []Message{
		{Role: "user", Content: query},
	}
	return c.llm.Chat(ctx, msgs)
}

// Index indexes documents into the vector store for later retrieval.
// This requires Embedding and VectorStore providers to be configured.
// This is a placeholder that will be fully implemented in Phase 4 (Pipeline).
func (c *Client) Index(ctx context.Context, documents []Document) error {
	if c.embedding == nil {
		return fmt.Errorf("ihandai: embedding provider not configured")
	}
	if c.vectordb == nil {
		return fmt.Errorf("ihandai: vector store not configured")
	}

	// TODO Phase 4: full pipeline — embed documents → insert into store
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

// WithLLM configures an LLM provider by name.
// The provider must be registered via a blank import of its plugin.
//
//	ihandai.WithLLM("openai", llm.WithModel("gpt-4o"))
func WithLLM(name string, opts ...llm.Option) Option {
	return func(c *Config) {
		c.LLMProvider = name
		c.LLMOptions = opts
	}
}

// WithEmbedding configures an embedding provider by name.
// The provider must be registered via a blank import of its plugin.
//
//	ihandai.WithEmbedding("ollama", embedding.WithModel("nomic-embed-text"))
func WithEmbedding(name string, opts ...embedding.Option) Option {
	return func(c *Config) {
		c.EmbeddingProvider = name
		c.EmbeddingOptions = opts
	}
}

// WithIndexEmbedding configures a separate embedding provider for indexing.
// When not set, the same provider as WithEmbedding is used for both.
// Useful when you want a local model for bulk indexing but a cloud model for queries.
//
//	ihandai.WithIndexEmbedding("ollama", embedding.WithModel("nomic-embed-text"))
func WithIndexEmbedding(name string, opts ...embedding.Option) Option {
	return func(c *Config) {
		c.IndexEmbeddingProvider = name
		c.IndexEmbeddingOptions = opts
	}
}

// WithVectorStore configures a vector store provider by name.
// The provider must be registered via a blank import of its plugin.
//
//	ihandai.WithVectorStore("qdrant", vectordb.WithURL("http://localhost:6333"))
func WithVectorStore(name string, opts ...vectordb.Option) Option {
	return func(c *Config) {
		c.VectorStoreProvider = name
		c.VectorStoreOptions = opts
	}
}
