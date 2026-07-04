package ihandai

import (
	"context"
	"fmt"
	"log/slog"
	"sync"

	"github.com/bachtiarpanjaitan/ihandai-go/pkg/core"
	"github.com/bachtiarpanjaitan/ihandai-go/pkg/embedding"
	"github.com/bachtiarpanjaitan/ihandai-go/pkg/llm"
	"github.com/bachtiarpanjaitan/ihandai-go/pkg/loader"
	"github.com/bachtiarpanjaitan/ihandai-go/pkg/prompt"
	"github.com/bachtiarpanjaitan/ihandai-go/pkg/reranker"
	"github.com/bachtiarpanjaitan/ihandai-go/pkg/retriever"
	"github.com/bachtiarpanjaitan/ihandai-go/pkg/splitter"
	"github.com/bachtiarpanjaitan/ihandai-go/pkg/vectordb"
)

// Config holds all configuration for the Client.
type Config struct {
	Logger *slog.Logger

	// Provider configurations
	LLMProvider           string
	LLMOptions            []llm.Option
	EmbeddingProvider     string
	EmbeddingOptions      []embedding.Option
	IndexEmbeddingProvider string
	IndexEmbeddingOptions  []embedding.Option
	VectorStoreProvider   string
	VectorStoreOptions    []vectordb.Option
}

// Option is a functional option for configuring the Client.
type Option func(*Config)

// Client is the main entry point for the ihandai library.
// It is safe for concurrent use.
type Client struct {
	mu  sync.RWMutex
	cfg *Config

	// Providers (immutable after New)
	llm            llm.ChatCompleter
	streamLLM      llm.StreamCompleter
	embedding      embedding.Embedder
	indexEmbedding embedding.Embedder
	vectorStore    vectordb.VectorSearcher

	// Pipeline components (protected by mu)
	promptBuilder prompt.PromptBuilder
	retriever     retriever.Retriever
	reranker      reranker.Reranker
	loader        loader.DocumentLoader
	splitter      splitter.TextSplitter
}

// New creates a new Client with the given options.
func New(opts ...Option) (*Client, error) {
	cfg := &Config{Logger: slog.Default()}
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
		// Also try as StreamCompleter if it implements it
		if sc, ok := chat.(llm.StreamCompleter); ok {
			c.streamLLM = sc
		}
	}

	// Open Embedding provider
	if cfg.EmbeddingProvider != "" {
		embed, err := embedding.Open(cfg.EmbeddingProvider, cfg.EmbeddingOptions...)
		if err != nil {
			return nil, fmt.Errorf("ihandai: %w", err)
		}
		c.embedding = embed
	}

	// Open Index Embedding provider (or reuse query embedding)
	if cfg.IndexEmbeddingProvider != "" {
		embed, err := embedding.Open(cfg.IndexEmbeddingProvider, cfg.IndexEmbeddingOptions...)
		if err != nil {
			return nil, fmt.Errorf("ihandai: %w", err)
		}
		c.indexEmbedding = embed
	} else {
		c.indexEmbedding = c.embedding
	}

	// Open VectorStore provider
	if cfg.VectorStoreProvider != "" {
		store, err := vectordb.Open(cfg.VectorStoreProvider, cfg.VectorStoreOptions...)
		if err != nil {
			return nil, fmt.Errorf("ihandai: %w", err)
		}
		c.vectorStore = store
	}

	// Default pipeline components
	c.loader = loader.NewFile()
	c.splitter = splitter.NewRecursive()
	c.promptBuilder = prompt.NewSimple()

	return c, nil
}

// Close releases any resources held by the Client.
func (c *Client) Close() error { return nil }

// LLM returns the configured LLM provider, or nil.
func (c *Client) LLM() llm.ChatCompleter { return c.llm }

// Embedding returns the configured embedding provider, or nil.
func (c *Client) Embedding() embedding.Embedder { return c.embedding }

// VectorStore returns the configured vector store provider, or nil.
func (c *Client) VectorStore() vectordb.VectorSearcher { return c.vectorStore }

// SetRetriever replaces the default retriever. Default wraps VectorStore with top-K.
func (c *Client) SetRetriever(r retriever.Retriever) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.retriever = r
}

// SetReranker sets an optional reranker (nil = skip reranking).
func (c *Client) SetReranker(r reranker.Reranker) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.reranker = r
}

// SetPromptBuilder replaces the default prompt builder.
func (c *Client) SetPromptBuilder(p prompt.PromptBuilder) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.promptBuilder = p
}

// SetLoader replaces the default file loader.
func (c *Client) SetLoader(l loader.DocumentLoader) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.loader = l
}

// SetSplitter replaces the default recursive splitter.
func (c *Client) SetSplitter(s splitter.TextSplitter) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.splitter = s
}

// Config returns a copy of the client's configuration.
func (c *Client) Config() *Config { cp := *c.cfg; return &cp }

// ---------------------------------------------------------------------------
// Pipeline
// ---------------------------------------------------------------------------

// Ask runs the full RAG query pipeline.
//
//	1. Embed query
//	2. Search vector store
//	3. Rerank (optional)
//	4. Build prompt
//	5. Chat completion
func (c *Client) Ask(ctx context.Context, query string) (*core.Response, error) {
	if c.llm == nil {
		return nil, fmt.Errorf("ihandai: LLM provider not configured")
	}
	if c.embedding == nil {
		return nil, fmt.Errorf("ihandai: embedding provider not configured")
	}
	if c.vectorStore == nil {
		return nil, fmt.Errorf("ihandai: vector store not configured")
	}

	// Step 1: Embed query
	queryVec, err := c.embedding.Embed(ctx, query)
	if err != nil {
		return nil, &PipelineError{Step: "embed", Err: err}
	}

	// Step 2: Search vector store
	c.mu.RLock()
	ret := c.retriever
	rerank := c.reranker
	pb := c.promptBuilder
	c.mu.RUnlock()

	if ret == nil {
		ret = retriever.NewTopK(c.vectorStore)
	}
	docs, err := ret.Retrieve(ctx, queryVec, retriever.WithTopK(5))
	if err != nil {
		return nil, &PipelineError{Step: "search", Err: err}
	}

	// Step 3: Rerank (optional)
	if rerank != nil && len(docs) > 0 {
		rawDocs := make([]core.Document, len(docs))
		for i, d := range docs {
			rawDocs[i] = d.Document
		}
		reranked, err := rerank.Rerank(ctx, query, rawDocs)
		if err != nil {
			c.cfg.Logger.Warn("reranking failed, continuing", "error", err)
		} else {
			docs = reranked
		}
	}

	// Step 4: Build prompt
	msgs, err := pb.Build(ctx, "", map[string]any{
		"query":     query,
		"documents": docs,
	})
	if err != nil {
		return nil, &PipelineError{Step: "prompt", Err: err}
	}

	// Step 5: Chat completion
	resp, err := c.llm.Chat(ctx, msgs)
	if err != nil {
		return nil, &PipelineError{Step: "chat", Err: err}
	}

	return resp, nil
}

// Index runs the document indexing pipeline.
//
//	1. Load documents from source
//	2. Split into chunks
//	3. Embed chunks
//	4. Insert into vector store
func (c *Client) Index(ctx context.Context, source string) error {
	if c.indexEmbedding == nil {
		return fmt.Errorf("ihandai: embedding provider not configured")
	}
	if c.vectorStore == nil {
		return fmt.Errorf("ihandai: vector store not configured")
	}

	// Step 1: Load documents
	c.mu.RLock()
	ld := c.loader
	sp := c.splitter
	c.mu.RUnlock()

	docs, err := ld.Load(ctx, source)
	if err != nil {
		return &PipelineError{Step: "load", Err: err}
	}
	if len(docs) == 0 {
		return nil
	}

	// Step 2: Split into chunks
	chunks, err := sp.Split(ctx, docs)
	if err != nil {
		return &PipelineError{Step: "split", Err: err}
	}
	if len(chunks) == 0 {
		return nil
	}

	// Step 3: Embed chunks
	texts := make([]string, len(chunks))
	for i, ch := range chunks {
		texts[i] = ch.Content
	}
	_, err = c.indexEmbedding.EmbedBatch(ctx, texts)
	if err != nil {
		return &PipelineError{Step: "embed", Err: err}
	}

	// Step 4: Insert into vector store
	// Convert chunks to documents with embedding metadata
	storeDocs := make([]core.Document, len(chunks))
	for i, ch := range chunks {
		storeDocs[i] = core.Document{
			ID:      ch.ID,
			Content: ch.Content,
			Metadata: map[string]any{
				"parent_id": ch.ParentID,
			},
		}
	}

	inserter, ok := c.vectorStore.(vectordb.VectorInserter)
	if !ok {
		return fmt.Errorf("ihandai: vector store does not support insertion")
	}
	if err := inserter.Insert(ctx, storeDocs); err != nil {
		return &PipelineError{Step: "insert", Err: err}
	}

	return nil
}

// AskStream runs the Ask pipeline with streaming response.
// Returns a channel that receives response chunks as they are generated.
func (c *Client) AskStream(ctx context.Context, query string) (<-chan llm.Chunk, error) {
	if c.streamLLM == nil {
		return nil, fmt.Errorf("ihandai: LLM provider does not support streaming")
	}
	if c.embedding == nil {
		return nil, fmt.Errorf("ihandai: embedding provider not configured")
	}
	if c.vectorStore == nil {
		return nil, fmt.Errorf("ihandai: vector store not configured")
	}

	// Run steps 1-4 synchronously, then stream step 5
	queryVec, err := c.embedding.Embed(ctx, query)
	if err != nil {
		return nil, &PipelineError{Step: "embed", Err: err}
	}

	c.mu.RLock()
	ret := c.retriever
	pb := c.promptBuilder
	c.mu.RUnlock()

	if ret == nil {
		ret = retriever.NewTopK(c.vectorStore)
	}
	docs, err := ret.Retrieve(ctx, queryVec, retriever.WithTopK(5))
	if err != nil {
		return nil, &PipelineError{Step: "search", Err: err}
	}

	msgs, err := pb.Build(ctx, "", map[string]any{
		"query":     query,
		"documents": docs,
	})
	if err != nil {
		return nil, &PipelineError{Step: "prompt", Err: err}
	}

	return c.streamLLM.ChatStream(ctx, msgs)
}

// ---------------------------------------------------------------------------
// Options
// ---------------------------------------------------------------------------

// WithLogger sets the structured logger.
func WithLogger(logger *slog.Logger) Option {
	return func(c *Config) { c.Logger = logger }
}

// WithLLM configures an LLM provider.
func WithLLM(name string, opts ...llm.Option) Option {
	return func(c *Config) { c.LLMProvider = name; c.LLMOptions = opts }
}

// WithEmbedding configures an embedding provider.
func WithEmbedding(name string, opts ...embedding.Option) Option {
	return func(c *Config) { c.EmbeddingProvider = name; c.EmbeddingOptions = opts }
}

// WithIndexEmbedding configures a separate embedding provider for indexing.
func WithIndexEmbedding(name string, opts ...embedding.Option) Option {
	return func(c *Config) { c.IndexEmbeddingProvider = name; c.IndexEmbeddingOptions = opts }
}

// WithVectorStore configures a vector store provider.
func WithVectorStore(name string, opts ...vectordb.Option) Option {
	return func(c *Config) { c.VectorStoreProvider = name; c.VectorStoreOptions = opts }
}
