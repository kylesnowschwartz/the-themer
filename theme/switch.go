package theme

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// SwitchOpts configures the switch operation.
type SwitchOpts struct {
	HomeDir string // injectable for testing; defaults to os.UserHomeDir()
}

// resolveHome returns opts.HomeDir if set, otherwise os.UserHomeDir().
func (o SwitchOpts) resolveHome() (string, error) {
	if o.HomeDir != "" {
		return o.HomeDir, nil
	}
	return os.UserHomeDir()
}

// SwitchResult holds the outcome of switching one app.
type SwitchResult struct {
	App     string
	Skipped bool
	Message string
	Err     error
}

// Switch activates a theme across all supported apps. Each app is handled
// independently — errors are collected best-effort.
func Switch(t Theme, opts SwitchOpts) []SwitchResult {
	home, err := opts.resolveHome()
	if err != nil {
		return []SwitchResult{{App: "home", Err: fmt.Errorf("resolving home directory: %w", err)}}
	}

	handlers := []struct {
		app     string
		switch_ func(t Theme, home string) (string, error)
	}{
		{"ghostty", switchGhostty},
		{"bat", switchBat},
		{"delta", switchDelta},
		{"fzf", switchFzf},
		{"starship", switchStarship},
		{"eza", switchEza},
		{"gh-dash", switchGhDash},
		{"neovim", switchNeovim},
		{"claude", switchClaude},
	}

	var results []SwitchResult
	for _, h := range handlers {
		msg, err := h.switch_(t, home)
		if msg == "" && err == nil {
			// Handler signaled nothing to do.
			results = append(results, SwitchResult{App: h.app, Skipped: true, Message: "not configured for this theme"})
			continue
		}
		results = append(results, SwitchResult{App: h.app, Message: msg, Err: err})
	}
	return results
}

// switchGhostty writes theme.local with the theme file reference.
// Ghostty reads custom themes from ~/.config/ghostty/themes/ by filename.
func switchGhostty(t Theme, home string) (string, error) {
	ghosttyDir := filepath.Join(t.Dir, "ghostty")
	if _, err := os.Stat(ghosttyDir); os.IsNotExist(err) {
		return "", nil
	}

	// Find the theme file in the ghostty dir.
	themeFile, err := firstFile(ghosttyDir)
	if err != nil {
		return "", fmt.Errorf("reading ghostty dir: %w", err)
	}
	if themeFile == "" {
		return "", nil
	}

	content := fmt.Sprintf("# Managed by the-themer — do not edit\ntheme = %s\n", themeFile)
	dest := filepath.Join(home, ".config", "ghostty", "theme.local")

	if err := os.MkdirAll(filepath.Dir(dest), 0o755); err != nil {
		return "", err
	}
	if err := os.WriteFile(dest, []byte(content), 0o644); err != nil {
		return "", err
	}
	return fmt.Sprintf("theme.local -> %s", themeFile), nil
}

// switchBat writes the bat theme name to bat-theme.txt.
// Generated themes use titleCase of the theme name (matching .tmTheme <name>).
// If only a reference is set, use that directly.
func switchBat(t Theme, home string) (string, error) {
	batDir := filepath.Join(t.Dir, "bat")
	hasBatDir := dirExists(batDir)
	refName := t.Config.References["bat"]

	if !hasBatDir && refName == "" {
		return "", nil
	}

	var themeName string
	if hasBatDir {
		// Generated bat theme name matches titleCase in the .tmTheme template.
		themeName = titleCase(t.Config.Theme.Name)
	} else {
		themeName = refName
	}

	dest := filepath.Join(home, ".config", "bat-theme.txt")
	if err := os.MkdirAll(filepath.Dir(dest), 0o755); err != nil {
		return "", err
	}
	if err := os.WriteFile(dest, []byte(themeName+"\n"), 0o644); err != nil {
		return "", err
	}
	return fmt.Sprintf("bat-theme.txt -> %s", themeName), nil
}

// switchDelta writes the delta feature name to delta-theme.txt.
// The feature name is derived from the gitconfig filename (without extension).
func switchDelta(t Theme, home string) (string, error) {
	deltaDir := filepath.Join(t.Dir, "delta")
	hasDeltaDir := dirExists(deltaDir)
	refName := t.Config.References["delta"]

	if !hasDeltaDir && refName == "" {
		return "", nil
	}

	var featureName string
	if hasDeltaDir {
		// The gitconfig filename (without .gitconfig) is the delta feature name.
		file, err := firstFile(deltaDir)
		if err != nil {
			return "", err
		}
		featureName = strings.TrimSuffix(file, ".gitconfig")
	} else {
		featureName = refName
	}

	dest := filepath.Join(home, ".config", "delta-theme.txt")
	if err := os.MkdirAll(filepath.Dir(dest), 0o755); err != nil {
		return "", err
	}
	if err := os.WriteFile(dest, []byte(featureName+"\n"), 0o644); err != nil {
		return "", err
	}
	return fmt.Sprintf("delta-theme.txt -> %s", featureName), nil
}

// switchFzf creates a symlink current.zsh pointing to the installed fzf config.
func switchFzf(t Theme, home string) (string, error) {
	fzfDir := filepath.Join(t.Dir, "fzf")
	if !dirExists(fzfDir) {
		return "", nil
	}

	srcFile, err := firstFile(fzfDir)
	if err != nil || srcFile == "" {
		return "", err
	}

	installedDir := filepath.Join(home, ".config", "the-themer", "fzf")
	installedFile := filepath.Join(installedDir, srcFile)
	link := filepath.Join(installedDir, "current.zsh")

	// Remove existing symlink before creating new one.
	os.Remove(link)
	if err := os.Symlink(installedFile, link); err != nil {
		return "", err
	}
	return fmt.Sprintf("fzf/current.zsh -> %s", srcFile), nil
}

// switchStarship symlinks ~/.config/starship.toml to the installed starship config.
func switchStarship(t Theme, home string) (string, error) {
	starshipDir := filepath.Join(t.Dir, "starship")
	if !dirExists(starshipDir) {
		return "", nil
	}

	srcFile, err := firstFile(starshipDir)
	if err != nil || srcFile == "" {
		return "", err
	}

	installedFile := filepath.Join(home, ".config", "the-themer", "starship", srcFile)
	link := filepath.Join(home, ".config", "starship.toml")

	os.Remove(link)
	if err := os.Symlink(installedFile, link); err != nil {
		return "", err
	}
	return fmt.Sprintf("starship.toml -> %s", srcFile), nil
}

// switchEza symlinks ~/.config/eza/theme.yml to the installed eza theme.
func switchEza(t Theme, home string) (string, error) {
	ezaDir := filepath.Join(t.Dir, "eza")
	if !dirExists(ezaDir) {
		return "", nil
	}

	srcFile, err := firstFile(ezaDir)
	if err != nil || srcFile == "" {
		return "", err
	}

	installedFile := filepath.Join(home, ".config", "eza", "themes", srcFile)
	link := filepath.Join(home, ".config", "eza", "theme.yml")

	if err := os.MkdirAll(filepath.Dir(link), 0o755); err != nil {
		return "", err
	}
	os.Remove(link)
	if err := os.Symlink(installedFile, link); err != nil {
		return "", err
	}
	return fmt.Sprintf("eza/theme.yml -> %s", srcFile), nil
}

// switchGhDash copies the installed gh-dash config to ~/.config/gh-dash/config.yml.
func switchGhDash(t Theme, home string) (string, error) {
	ghDashDir := filepath.Join(t.Dir, "gh-dash")
	if !dirExists(ghDashDir) {
		return "", nil
	}

	srcFile, err := firstFile(ghDashDir)
	if err != nil || srcFile == "" {
		return "", err
	}

	src := filepath.Join(home, ".config", "the-themer", "gh-dash", srcFile)
	dest := filepath.Join(home, ".config", "gh-dash", "config.yml")

	if err := os.MkdirAll(filepath.Dir(dest), 0o755); err != nil {
		return "", err
	}
	if err := copyFile(src, dest); err != nil {
		return "", err
	}
	return fmt.Sprintf("gh-dash/config.yml -> %s", srcFile), nil
}

// switchNeovim uses headless nvim to set the colorscheme via Themery.
func switchNeovim(t Theme, home string) (string, error) {
	name := t.Config.References["neovim"]
	if name == "" {
		return "", nil
	}

	nvimPath, err := exec.LookPath("nvim")
	if err != nil {
		return "nvim not on PATH, skipped", nil
	}

	luaCmd := fmt.Sprintf(`pcall(function() require('themery').setThemeByName('%s', true) end)`, name)
	cmd := exec.Command(nvimPath, "--headless", "-c", fmt.Sprintf("lua %s", luaCmd), "-c", "qa")
	if out, err := cmd.CombinedOutput(); err != nil {
		return "", fmt.Errorf("nvim themery switch failed: %s: %w", strings.TrimSpace(string(out)), err)
	}
	return fmt.Sprintf("neovim -> %s", name), nil
}

// switchClaude edits ~/.claude.json to set the theme key.
// "dark" deletes the key (dark is the default). Pure Go, no jq dependency.
func switchClaude(t Theme, home string) (string, error) {
	value := t.Config.References["claude"]
	if value == "" {
		return "", nil
	}

	claudePath := filepath.Join(home, ".claude.json")

	var data map[string]any

	raw, err := os.ReadFile(claudePath)
	if err != nil {
		if !os.IsNotExist(err) {
			return "", fmt.Errorf("reading claude.json: %w", err)
		}
		// File doesn't exist — start fresh.
		data = make(map[string]any)
	} else {
		if err := json.Unmarshal(raw, &data); err != nil {
			return "", fmt.Errorf("parsing claude.json: %w", err)
		}
	}

	if value == "dark" {
		delete(data, "theme")
	} else {
		data["theme"] = value
	}

	out, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return "", fmt.Errorf("marshaling claude.json: %w", err)
	}
	// Append newline for clean file endings.
	out = append(out, '\n')

	if err := os.WriteFile(claudePath, out, 0o644); err != nil {
		return "", err
	}

	if value == "dark" {
		return "claude.json -> removed theme key (dark is default)", nil
	}
	return fmt.Sprintf("claude.json -> %s", value), nil
}

// titleCase capitalizes the first letter of each hyphen-separated word.
// "cobalt-next-neon" -> "Cobalt-Next-Neon". Used for bat theme names.
func titleCase(s string) string {
	parts := strings.Split(s, "-")
	for i, part := range parts {
		if len(part) > 0 {
			r := []rune(part)
			if r[0] >= 'a' && r[0] <= 'z' {
				r[0] -= 32
			}
			parts[i] = string(r)
		}
	}
	return strings.Join(parts, "-")
}

// firstFile returns the name of the first regular file in dir, or "" if empty.
func firstFile(dir string) (string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return "", err
	}
	for _, e := range entries {
		if !e.IsDir() {
			return e.Name(), nil
		}
	}
	return "", nil
}

// dirExists returns true if path exists and is a directory.
func dirExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}
