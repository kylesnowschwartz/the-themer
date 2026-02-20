package theme

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// stateDir returns the path to the-themer's state directory.
func stateDir(home string) string {
	return filepath.Join(home, ".config", "the-themer")
}

// statePath returns the path to the current theme state file.
func statePath(home string) string {
	return filepath.Join(stateDir(home), "current")
}

// ReadState reads the currently active theme name from the state file.
// Returns an empty string (not an error) if no theme has been set.
func ReadState(home string) (string, error) {
	data, err := os.ReadFile(statePath(home))
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil
		}
		return "", fmt.Errorf("reading state file: %w", err)
	}
	return strings.TrimSpace(string(data)), nil
}

// WriteState records the active theme name to the state file.
func WriteState(home, themeName string) error {
	dir := stateDir(home)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("creating state directory: %w", err)
	}
	if err := os.WriteFile(statePath(home), []byte(themeName+"\n"), 0o644); err != nil {
		return fmt.Errorf("writing state file: %w", err)
	}
	return nil
}

// defaultPath returns the path to a variant default file (default-dark or default-light).
func defaultPath(home, variant string) string {
	return filepath.Join(stateDir(home), "default-"+variant)
}

// ReadDefault reads the configured default theme for a variant ("dark" or "light").
// Returns an empty string (not an error) if no default has been set.
func ReadDefault(home, variant string) (string, error) {
	data, err := os.ReadFile(defaultPath(home, variant))
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil
		}
		return "", fmt.Errorf("reading default-%s: %w", variant, err)
	}
	return strings.TrimSpace(string(data)), nil
}

// WriteDefault records a theme name as the default for a variant ("dark" or "light").
func WriteDefault(home, variant, themeName string) error {
	dir := stateDir(home)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("creating state directory: %w", err)
	}
	if err := os.WriteFile(defaultPath(home, variant), []byte(themeName+"\n"), 0o644); err != nil {
		return fmt.Errorf("writing default-%s: %w", variant, err)
	}
	return nil
}
