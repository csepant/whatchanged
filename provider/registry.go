package provider

import (
	"fmt"
	"sort"
)

var registry = map[string]Provider{}

// Register adds a provider to the global registry.
func Register(p Provider) {
	registry[p.ResourceType()] = p
}

// Get returns a provider by resource type name.
func Get(name string) (Provider, error) {
	p, ok := registry[name]
	if !ok {
		return nil, fmt.Errorf("unknown provider: %s", name)
	}
	return p, nil
}

// All returns all registered providers.
func All() map[string]Provider {
	return registry
}

// Names returns a sorted list of all registered provider names.
func Names() []string {
	names := make([]string, 0, len(registry))
	for name := range registry {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}
