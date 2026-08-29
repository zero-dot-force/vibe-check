package metrics

import "fmt"

// Registry manages the association between language identifiers and [Adapter]
// implementations. Registry is not safe for concurrent use — it should be
// configured at startup and used read-only thereafter.
// It MUST be passed via dependency injection, never stored as a global variable.
type Registry struct {
	adapters map[string]Adapter
}

// NewRegistry creates an empty Registry ready for adapter registration.
func NewRegistry() *Registry {
	return &Registry{
		adapters: make(map[string]Adapter),
	}
}

// Register adds an adapter to the registry, keyed by its Language() return value.
// Returns an error if an adapter for the same language is already registered.
func (r *Registry) Register(a Adapter) error {
	lang := a.Language()
	if _, exists := r.adapters[lang]; exists {
		return fmt.Errorf("register adapter: language %q is already registered", lang)
	}
	r.adapters[lang] = a
	return nil
}

// Get retrieves the adapter registered for the given language identifier.
// Returns an error if no adapter is registered for the language.
func (r *Registry) Get(language string) (Adapter, error) {
	a, ok := r.adapters[language]
	if !ok {
		return nil, fmt.Errorf("get adapter: no adapter registered for language %q", language)
	}
	return a, nil
}
