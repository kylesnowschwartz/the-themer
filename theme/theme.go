// Package theme provides theme discovery, installation, and switching.
// A theme is a directory under themes/ containing a palette.toml and
// per-app config subdirectories. Directory presence determines app support.
package theme

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/kylesnowschwartz/the-themer/palette"
)

// Theme holds a loaded theme with its parsed palette config and directory path.
type Theme struct {
	Name   string
	Dir    string // absolute path to the theme directory
	Config palette.Config
}

// LoadTheme reads a theme directory, parses its palette.toml, and returns
// a fully resolved Theme. The themesDir is the parent directory containing
// all theme directories (e.g., "./themes/").
func LoadTheme(themesDir, name string) (Theme, error) {
	dir, err := filepath.Abs(filepath.Join(themesDir, name))
	if err != nil {
		return Theme{}, fmt.Errorf("resolving theme path: %w", err)
	}

	palettePath := filepath.Join(dir, "palette.toml")
	cfg, err := palette.Load(palettePath)
	if err != nil {
		return Theme{}, fmt.Errorf("loading theme %q: %w", name, err)
	}

	return Theme{
		Name:   name,
		Dir:    dir,
		Config: cfg,
	}, nil
}

// ListThemes scans themesDir for directories containing a palette.toml
// and returns their names in sorted order.
func ListThemes(themesDir string) ([]string, error) {
	entries, err := os.ReadDir(themesDir)
	if err != nil {
		return nil, fmt.Errorf("reading themes directory: %w", err)
	}

	var names []string
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		palettePath := filepath.Join(themesDir, e.Name(), "palette.toml")
		if _, err := os.Stat(palettePath); err == nil {
			names = append(names, e.Name())
		}
	}
	return names, nil
}

// AppDirs lists the app subdirectories present in a theme directory.
// Each subdirectory name corresponds to an app (ghostty, bat, delta, etc.).
// Only directories are returned, not files like palette.toml.
func AppDirs(themeDir string) ([]string, error) {
	entries, err := os.ReadDir(themeDir)
	if err != nil {
		return nil, fmt.Errorf("reading theme directory: %w", err)
	}

	var apps []string
	for _, e := range entries {
		if e.IsDir() {
			apps = append(apps, e.Name())
		}
	}
	return apps, nil
}
