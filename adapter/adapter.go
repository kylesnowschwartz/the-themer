// Package adapter defines the interface that all theme adapters must implement
// and provides a registry for adapter discovery.
package adapter

import (
	"github.com/kylesnowschwartz/the-themer/palette"
)

// Adapter generates a themed config file for a specific application.
type Adapter interface {
	// Name returns the adapter's identifier (e.g., "ghostty", "starship", "neovim").
	Name() string

	// DirName returns the subdirectory name within the output directory.
	DirName() string

	// FileName returns the output filename for the given theme name.
	FileName(themeName string) string

	// Generate renders the themed config file content from the palette.
	Generate(cfg palette.Config) ([]byte, error)
}

// adapters holds all registered adapters.
var adapters []Adapter

// Register adds an adapter to the global registry.
// Typically called from an adapter package's init() function.
func Register(a Adapter) {
	adapters = append(adapters, a)
}

// All returns all registered adapters.
func All() []Adapter {
	return adapters
}

// ByName returns adapters whose names match the given list.
// If names is empty, returns all adapters.
func ByName(names []string) []Adapter {
	if len(names) == 0 {
		return adapters
	}
	nameSet := make(map[string]bool, len(names))
	for _, n := range names {
		nameSet[n] = true
	}
	var filtered []Adapter
	for _, a := range adapters {
		if nameSet[a.Name()] {
			filtered = append(filtered, a)
		}
	}
	return filtered
}
