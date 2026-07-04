package embedding

import (
	"fmt"
	"sync"
)

// EmbedderFactory is a function that creates an Embedder from a Config.
type EmbedderFactory func(cfg Config) (Embedder, error)

var (
	embedRegistry   = make(map[string]EmbedderFactory)
	embedRegistryMu sync.RWMutex
)

// Register registers an Embedder factory under the given name.
// It is meant to be called from init() functions in provider plugins.
//
//	func init() {
//	    embedding.Register("openai", newOpenAIEmbedder)
//	}
//
// Register panics if name is empty or factory is nil.
func Register(name string, factory EmbedderFactory) {
	if name == "" {
		panic("embedding: Register name is empty")
	}
	if factory == nil {
		panic("embedding: Register factory is nil")
	}

	embedRegistryMu.Lock()
	defer embedRegistryMu.Unlock()

	if _, dup := embedRegistry[name]; dup {
		panic("embedding: Register called twice for provider " + name)
	}
	embedRegistry[name] = factory
}

// Open creates an Embedder from the named provider.
// Options are applied to a default Config before being passed to the factory.
//
// If the named provider has not been registered (via a blank import of its plugin),
// Open returns an error.
func Open(name string, opts ...Option) (Embedder, error) {
	embedRegistryMu.RLock()
	factory, ok := embedRegistry[name]
	embedRegistryMu.RUnlock()

	if !ok {
		return nil, fmt.Errorf("embedding: unknown provider %q (forgot to import its plugin?)", name)
	}

	cfg := Config{}
	for _, opt := range opts {
		opt(&cfg)
	}

	return factory(cfg)
}

// Registered returns the names of all registered embedding providers.
func Registered() []string {
	embedRegistryMu.RLock()
	defer embedRegistryMu.RUnlock()

	names := make([]string, 0, len(embedRegistry))
	for name := range embedRegistry {
		names = append(names, name)
	}
	return names
}
