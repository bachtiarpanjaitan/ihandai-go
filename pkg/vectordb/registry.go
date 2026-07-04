package vectordb

import (
	"fmt"
	"sync"
)

// SearcherFactory is a function that creates a VectorSearcher from a Config.
type SearcherFactory func(cfg Config) (VectorSearcher, error)

var (
	searcherRegistry   = make(map[string]SearcherFactory)
	searcherRegistryMu sync.RWMutex
)

// Register registers a VectorSearcher factory under the given name.
// It is meant to be called from init() functions in provider plugins.
//
//	func init() {
//	    vectordb.Register("qdrant", newQdrantSearcher)
//	}
//
// Register panics if name is empty or factory is nil.
func Register(name string, factory SearcherFactory) {
	if name == "" {
		panic("vectordb: Register name is empty")
	}
	if factory == nil {
		panic("vectordb: Register factory is nil")
	}

	searcherRegistryMu.Lock()
	defer searcherRegistryMu.Unlock()

	if _, dup := searcherRegistry[name]; dup {
		panic("vectordb: Register called twice for provider " + name)
	}
	searcherRegistry[name] = factory
}

// Open creates a VectorSearcher from the named provider.
// Options are applied to a default Config before being passed to the factory.
//
// If the named provider has not been registered (via a blank import of its plugin),
// Open returns an error.
func Open(name string, opts ...Option) (VectorSearcher, error) {
	searcherRegistryMu.RLock()
	factory, ok := searcherRegistry[name]
	searcherRegistryMu.RUnlock()

	if !ok {
		return nil, fmt.Errorf("vectordb: unknown provider %q (forgot to import its plugin?)", name)
	}

	cfg := Config{}
	for _, opt := range opts {
		opt(&cfg)
	}

	return factory(cfg)
}

// Registered returns the names of all registered vector store providers.
func Registered() []string {
	searcherRegistryMu.RLock()
	defer searcherRegistryMu.RUnlock()

	names := make([]string, 0, len(searcherRegistry))
	for name := range searcherRegistry {
		names = append(names, name)
	}
	return names
}
