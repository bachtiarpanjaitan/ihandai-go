package llm

import (
	"fmt"
	"sync"
)

// ChatFactory is a function that creates a ChatCompleter from a Config.
type ChatFactory func(cfg Config) (ChatCompleter, error)

var (
	chatRegistry   = make(map[string]ChatFactory)
	chatRegistryMu sync.RWMutex
)

// Register registers a ChatCompleter factory under the given name.
// It is meant to be called from init() functions in provider plugins.
//
//	func init() {
//	    llm.Register("openai", newOpenAIChat)
//	}
//
// Register panics if name is empty or factory is nil.
func Register(name string, factory ChatFactory) {
	if name == "" {
		panic("llm: Register name is empty")
	}
	if factory == nil {
		panic("llm: Register factory is nil")
	}

	chatRegistryMu.Lock()
	defer chatRegistryMu.Unlock()

	if _, dup := chatRegistry[name]; dup {
		panic("llm: Register called twice for provider " + name)
	}
	chatRegistry[name] = factory
}

// Open creates a ChatCompleter from the named provider.
// Options are applied to a default Config before being passed to the factory.
//
// If the named provider has not been registered (via a blank import of its plugin),
// Open returns an error.
func Open(name string, opts ...Option) (ChatCompleter, error) {
	chatRegistryMu.RLock()
	factory, ok := chatRegistry[name]
	chatRegistryMu.RUnlock()

	if !ok {
		return nil, fmt.Errorf("llm: unknown provider %q (forgot to import its plugin?)", name)
	}

	cfg := Config{}
	for _, opt := range opts {
		opt(&cfg)
	}

	return factory(cfg)
}

// Registered returns the names of all registered LLM providers.
func Registered() []string {
	chatRegistryMu.RLock()
	defer chatRegistryMu.RUnlock()

	names := make([]string, 0, len(chatRegistry))
	for name := range chatRegistry {
		names = append(names, name)
	}
	return names
}
