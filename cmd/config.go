package cmd

import (
	"os"
	"path/filepath"
)

// defaultThemesDir returns the themes directory path.
// Checks THE_THEMER_THEMES_DIR env var first, then resolves "themes/"
// relative to the executable's real path (follows symlinks).
func defaultThemesDir() string {
	if dir := os.Getenv("THE_THEMER_THEMES_DIR"); dir != "" {
		return dir
	}

	exe, err := os.Executable()
	if err != nil {
		return "./themes/"
	}

	real, err := filepath.EvalSymlinks(exe)
	if err != nil {
		return "./themes/"
	}

	return filepath.Join(filepath.Dir(real), "themes")
}
