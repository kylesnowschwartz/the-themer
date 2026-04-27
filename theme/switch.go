package theme

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
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
		{"tcm", switchTCM},
		{"starship", switchStarship},
		{"eza", switchEza},
		{"gh-dash", switchGhDash},
		{"neovim", switchNeovim},
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

// switchGhostty writes theme.local with the theme filename reference.
// Ghostty matches the `theme` value against filenames in its themes directory,
// so we use the full filename including any .ghostty extension.
func switchGhostty(t Theme, home string) (string, error) {
	ghosttyDir := filepath.Join(t.Dir, "ghostty")
	if _, err := os.Stat(ghosttyDir); os.IsNotExist(err) {
		return "", nil
	}

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
	msg := fmt.Sprintf("theme.local -> %s", themeFile)
	if nudge := nudgeGhosttyReload(); nudge != "" {
		msg += "; " + nudge
	}
	return msg, nil
}

// nudgeGhosttyReload sends Cmd+Shift+R via System Events on macOS so a
// running Ghostty picks up theme.local without manual intervention. Best-
// effort: if Ghostty isn't running, osascript is missing, the platform
// isn't macOS, or TCC denies the action, the function silently returns ""
// and the switch result still reports success.
//
// Why osascript and not an IPC call: Ghostty's macOS build has no CLI->
// running-app channel (`ghostty +new-window` reports "not supported on
// this platform"), so we drive the existing reload_config keybind via
// AppleScript. Requires Ghostty to be granted Automation → System Events
// in System Settings → Privacy & Security; first invocation triggers the
// TCC prompt.
func nudgeGhosttyReload() string {
	if runtime.GOOS != "darwin" {
		return ""
	}
	osa, err := exec.LookPath("osascript")
	if err != nil {
		return ""
	}
	// Don't launch Ghostty just to reload it.
	check := exec.Command(osa, "-e", `application id "com.mitchellh.ghostty" is running`)
	out, err := check.Output()
	if err != nil || strings.TrimSpace(string(out)) != "true" {
		return ""
	}
	// activate brings Ghostty to focus so the keystroke lands in it. In the
	// common case (running `theme` from a Ghostty terminal) Ghostty is
	// already frontmost, so activate is a visual no-op.
	script := `tell application id "com.mitchellh.ghostty" to activate
tell application "System Events" to keystroke "r" using {command down, shift down}`
	if err := exec.Command(osa, "-e", script).Run(); err != nil {
		return fmt.Sprintf("reload nudge failed: %v", err)
	}
	return "reloaded"
}

// switchBat writes the bat theme name to bat-theme.txt.
// bat identifies custom themes by filename (sans .tmTheme extension), so we
// read the actual filename from the theme's bat/ directory.
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
		file, err := firstFile(batDir)
		if err != nil {
			return "", fmt.Errorf("reading bat dir: %w", err)
		}
		if file == "" {
			return "", nil
		}
		themeName = strings.TrimSuffix(file, ".tmTheme")
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

// switchTCM atomically writes ~/.config/tcm/active-theme.json with the
// installed theme JSON's contents. We deliberately do NOT symlink: tcm
// watches the file via fs.watch, and on macOS fs.watch follows symlinks to
// the resolved target — swapping the symlink to a different file (whose
// content didn't change) doesn't fire the watcher. The brief calls this out:
// "Atomic writes (rename trick) are detected." So we write to a sibling .tmp
// file and rename onto active-theme.json; rename is atomic on the same
// filesystem and reliably triggers fs.watch.
func switchTCM(t Theme, home string) (string, error) {
	tcmDir := filepath.Join(t.Dir, "tcm")
	if !dirExists(tcmDir) {
		return "", nil
	}

	srcFile, err := firstFile(tcmDir)
	if err != nil || srcFile == "" {
		return "", err
	}

	installedFile := filepath.Join(home, ".config", "tcm", srcFile)
	dest := filepath.Join(home, ".config", "tcm", "active-theme.json")
	tmp := dest + ".tmp"

	if err := os.MkdirAll(filepath.Dir(dest), 0o755); err != nil {
		return "", err
	}

	// Read the installed source. Required because dest may currently be a
	// symlink (from the previous symlink-based implementation); we want to
	// land a regular file at dest so fs.watch tracks it directly.
	data, err := os.ReadFile(installedFile)
	if err != nil {
		return "", fmt.Errorf("reading %s: %w", installedFile, err)
	}

	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		return "", err
	}
	if err := os.Rename(tmp, dest); err != nil {
		os.Remove(tmp)
		return "", err
	}
	return fmt.Sprintf("tcm/active-theme.json <- %s (atomic write)", srcFile), nil
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
