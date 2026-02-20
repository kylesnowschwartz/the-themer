package theme

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// minimalPaletteTOML is a valid palette.toml for testing.
const minimalPaletteTOML = `
[theme]
name = "test-theme"
variant = "dark"

[palette]
bg = "#111111"
fg = "#eeeeee"
color0 = "#000000"
color1 = "#aa0000"
color2 = "#00aa00"
color3 = "#aaaa00"
color4 = "#0000aa"
color5 = "#aa00aa"
color6 = "#00aaaa"
color7 = "#aaaaaa"
color8 = "#555555"
color9 = "#ff0000"
color10 = "#00ff00"
color11 = "#ffff00"
color12 = "#0000ff"
color13 = "#ff00ff"
color14 = "#00ffff"
color15 = "#ffffff"

[references]
neovim = "test-scheme"
claude = "dark"
bat = "Dracula"
`

// setupThemeDir creates a minimal theme directory in tmpdir with optional
// app subdirectories. Returns the themes root dir and the theme dir.
func setupThemeDir(t *testing.T, appDirs []string, paletteTOML string) (themesDir, themeDir string) {
	t.Helper()
	tmpDir := t.TempDir()
	themesDir = filepath.Join(tmpDir, "themes")
	themeDir = filepath.Join(themesDir, "test-theme")

	if err := os.MkdirAll(themeDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(themeDir, "palette.toml"), []byte(paletteTOML), 0o644); err != nil {
		t.Fatal(err)
	}

	for _, app := range appDirs {
		appDir := filepath.Join(themeDir, app)
		if err := os.MkdirAll(appDir, 0o755); err != nil {
			t.Fatal(err)
		}
	}
	return themesDir, themeDir
}

// writeFile is a test helper that creates a file with content.
func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestLoadTheme(t *testing.T) {
	themesDir, _ := setupThemeDir(t, nil, minimalPaletteTOML)

	th, err := LoadTheme(themesDir, "test-theme")
	if err != nil {
		t.Fatalf("LoadTheme failed: %v", err)
	}

	if th.Name != "test-theme" {
		t.Errorf("Name = %q, want %q", th.Name, "test-theme")
	}
	if th.Config.Theme.Name != "test-theme" {
		t.Errorf("Config.Theme.Name = %q, want %q", th.Config.Theme.Name, "test-theme")
	}
	if th.Config.References["neovim"] != "test-scheme" {
		t.Errorf("References[neovim] = %q, want %q", th.Config.References["neovim"], "test-scheme")
	}
}

func TestListThemes(t *testing.T) {
	tmpDir := t.TempDir()
	themesDir := filepath.Join(tmpDir, "themes")

	// Create two valid themes and one dir without palette.toml
	for _, name := range []string{"alpha", "beta", "not-a-theme"} {
		dir := filepath.Join(themesDir, name)
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatal(err)
		}
	}
	writeFile(t, filepath.Join(themesDir, "alpha", "palette.toml"), minimalPaletteTOML)
	writeFile(t, filepath.Join(themesDir, "beta", "palette.toml"), minimalPaletteTOML)

	names, err := ListThemes(themesDir)
	if err != nil {
		t.Fatalf("ListThemes failed: %v", err)
	}

	if len(names) != 2 {
		t.Fatalf("ListThemes returned %d, want 2: %v", len(names), names)
	}
	// ReadDir returns sorted order.
	if names[0] != "alpha" || names[1] != "beta" {
		t.Errorf("ListThemes = %v, want [alpha beta]", names)
	}
}

func TestAppDirs(t *testing.T) {
	_, themeDir := setupThemeDir(t, []string{"ghostty", "bat", "delta"}, minimalPaletteTOML)

	apps, err := AppDirs(themeDir)
	if err != nil {
		t.Fatalf("AppDirs failed: %v", err)
	}

	if len(apps) != 3 {
		t.Fatalf("AppDirs returned %d, want 3: %v", len(apps), apps)
	}
}

func TestInstall_CopiesFiles(t *testing.T) {
	themesDir, themeDir := setupThemeDir(t, []string{"ghostty", "fzf"}, minimalPaletteTOML)

	// Create source files in the theme.
	writeFile(t, filepath.Join(themeDir, "ghostty", "test-theme.ghostty"), "ghostty config content")
	writeFile(t, filepath.Join(themeDir, "fzf", "test-theme.zsh"), "fzf config content")

	home := t.TempDir()
	th, err := LoadTheme(themesDir, "test-theme")
	if err != nil {
		t.Fatal(err)
	}

	results := Install(th, InstallOpts{HomeDir: home})

	// Check ghostty was installed.
	ghosttyDest := filepath.Join(home, ".config", "ghostty", "themes", "test-theme.ghostty")
	assertFileContains(t, ghosttyDest, "ghostty config content")

	// Check fzf was installed.
	fzfDest := filepath.Join(home, ".config", "the-themer", "fzf", "test-theme.zsh")
	assertFileContains(t, fzfDest, "fzf config content")

	// Check that missing apps were skipped, not errored.
	for _, r := range results {
		if r.Err != nil {
			t.Errorf("unexpected error for %s: %v", r.App, r.Err)
		}
	}
}

func TestSwitch_GhosttyThemeLocal(t *testing.T) {
	themesDir, themeDir := setupThemeDir(t, []string{"ghostty"}, minimalPaletteTOML)
	writeFile(t, filepath.Join(themeDir, "ghostty", "test-theme.ghostty"), "config")

	home := t.TempDir()
	th, err := LoadTheme(themesDir, "test-theme")
	if err != nil {
		t.Fatal(err)
	}

	results := Switch(th, SwitchOpts{HomeDir: home})
	checkNoErrors(t, results)

	themeLocal := filepath.Join(home, ".config", "ghostty", "theme.local")
	assertFileContains(t, themeLocal, "theme = test-theme.ghostty")
}

func TestSwitch_BatThemeTxt(t *testing.T) {
	themesDir, themeDir := setupThemeDir(t, []string{"bat"}, minimalPaletteTOML)
	writeFile(t, filepath.Join(themeDir, "bat", "test-theme.tmTheme"), "bat content")

	home := t.TempDir()
	th, err := LoadTheme(themesDir, "test-theme")
	if err != nil {
		t.Fatal(err)
	}

	results := Switch(th, SwitchOpts{HomeDir: home})
	checkNoErrors(t, results)

	batTheme := filepath.Join(home, ".config", "bat-theme.txt")
	assertFileContains(t, batTheme, "test-theme")
}

func TestSwitch_DeltaThemeTxt(t *testing.T) {
	themesDir, themeDir := setupThemeDir(t, []string{"delta"}, minimalPaletteTOML)
	writeFile(t, filepath.Join(themeDir, "delta", "test-theme.gitconfig"), "delta content")

	home := t.TempDir()
	th, err := LoadTheme(themesDir, "test-theme")
	if err != nil {
		t.Fatal(err)
	}

	results := Switch(th, SwitchOpts{HomeDir: home})
	checkNoErrors(t, results)

	deltaTheme := filepath.Join(home, ".config", "delta-theme.txt")
	assertFileContains(t, deltaTheme, "test-theme")
}

func TestSwitch_FzfSymlink(t *testing.T) {
	themesDir, themeDir := setupThemeDir(t, []string{"fzf"}, minimalPaletteTOML)
	writeFile(t, filepath.Join(themeDir, "fzf", "test-theme.zsh"), "fzf content")

	home := t.TempDir()

	// Install first so the target file exists for the symlink.
	th, err := LoadTheme(themesDir, "test-theme")
	if err != nil {
		t.Fatal(err)
	}
	Install(th, InstallOpts{HomeDir: home})

	results := Switch(th, SwitchOpts{HomeDir: home})
	checkNoErrors(t, results)

	link := filepath.Join(home, ".config", "the-themer", "fzf", "current.zsh")
	target, err := os.Readlink(link)
	if err != nil {
		t.Fatalf("expected symlink at %s: %v", link, err)
	}
	if !strings.HasSuffix(target, "test-theme.zsh") {
		t.Errorf("symlink target = %q, want suffix test-theme.zsh", target)
	}
}

func TestSwitch_StarshipSymlink(t *testing.T) {
	themesDir, themeDir := setupThemeDir(t, []string{"starship"}, minimalPaletteTOML)
	writeFile(t, filepath.Join(themeDir, "starship", "chef-starship.toml"), "starship content")

	home := t.TempDir()

	th, err := LoadTheme(themesDir, "test-theme")
	if err != nil {
		t.Fatal(err)
	}
	Install(th, InstallOpts{HomeDir: home})

	results := Switch(th, SwitchOpts{HomeDir: home})
	checkNoErrors(t, results)

	link := filepath.Join(home, ".config", "starship.toml")
	target, err := os.Readlink(link)
	if err != nil {
		t.Fatalf("expected symlink at %s: %v", link, err)
	}
	if !strings.HasSuffix(target, "chef-starship.toml") {
		t.Errorf("symlink target = %q, want suffix chef-starship.toml", target)
	}
}

func TestSwitch_ClaudeJSON_Light(t *testing.T) {
	// Use a palette with claude = "light" reference.
	toml := strings.Replace(minimalPaletteTOML, `claude = "dark"`, `claude = "light"`, 1)
	themesDir, _ := setupThemeDir(t, nil, toml)

	home := t.TempDir()
	th, err := LoadTheme(themesDir, "test-theme")
	if err != nil {
		t.Fatal(err)
	}

	results := Switch(th, SwitchOpts{HomeDir: home})
	checkNoErrors(t, results)

	claudeJSON := filepath.Join(home, ".claude.json")
	assertFileContains(t, claudeJSON, `"theme": "light"`)
}

func TestSwitch_ClaudeJSON_DarkRemovesKey(t *testing.T) {
	themesDir, _ := setupThemeDir(t, nil, minimalPaletteTOML)

	home := t.TempDir()
	// Create an existing claude.json with a theme key.
	writeFile(t, filepath.Join(home, ".claude.json"), `{"theme": "light", "other": "value"}`)

	th, err := LoadTheme(themesDir, "test-theme")
	if err != nil {
		t.Fatal(err)
	}

	results := Switch(th, SwitchOpts{HomeDir: home})
	checkNoErrors(t, results)

	claudeJSON := filepath.Join(home, ".claude.json")
	content, err := os.ReadFile(claudeJSON)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(content), "theme") {
		t.Errorf("claude.json should not contain theme key for dark, got: %s", content)
	}
	// Verify other keys are preserved.
	if !strings.Contains(string(content), `"other"`) {
		t.Errorf("claude.json should preserve other keys, got: %s", content)
	}
}

func TestSwitch_ClaudeJSON_PreservesPermissions(t *testing.T) {
	toml := strings.Replace(minimalPaletteTOML, `claude = "dark"`, `claude = "light"`, 1)
	themesDir, _ := setupThemeDir(t, nil, toml)

	home := t.TempDir()
	claudeJSON := filepath.Join(home, ".claude.json")

	// Create claude.json with non-default permissions (0o600).
	writeFile(t, claudeJSON, `{"existing": true}`)
	if err := os.Chmod(claudeJSON, 0o600); err != nil {
		t.Fatal(err)
	}

	th, err := LoadTheme(themesDir, "test-theme")
	if err != nil {
		t.Fatal(err)
	}

	results := Switch(th, SwitchOpts{HomeDir: home})
	checkNoErrors(t, results)

	// Verify permissions were preserved.
	info, err := os.Stat(claudeJSON)
	if err != nil {
		t.Fatal(err)
	}
	if got := info.Mode().Perm(); got != 0o600 {
		t.Errorf("claude.json permissions = %o, want 600", got)
	}
	// Verify content is correct.
	assertFileContains(t, claudeJSON, `"theme": "light"`)
	assertFileContains(t, claudeJSON, `"existing": true`)
}

func TestSwitch_ReferenceFallback_Bat(t *testing.T) {
	// Theme with no bat/ dir but references.bat = "Dracula".
	themesDir, _ := setupThemeDir(t, nil, minimalPaletteTOML)

	home := t.TempDir()
	th, err := LoadTheme(themesDir, "test-theme")
	if err != nil {
		t.Fatal(err)
	}

	results := Switch(th, SwitchOpts{HomeDir: home})
	checkNoErrors(t, results)

	batTheme := filepath.Join(home, ".config", "bat-theme.txt")
	assertFileContains(t, batTheme, "Dracula")
}

func TestSwitch_MissingApps_Skipped(t *testing.T) {
	// Theme with no app dirs and no references for fzf/eza/etc.
	toml := `
[theme]
name = "bare"
variant = "dark"

[palette]
bg = "#111111"
fg = "#eeeeee"
color0 = "#000000"
color1 = "#aa0000"
color2 = "#00aa00"
color3 = "#aaaa00"
color4 = "#0000aa"
color5 = "#aa00aa"
color6 = "#00aaaa"
color7 = "#aaaaaa"
color8 = "#555555"
color9 = "#ff0000"
color10 = "#00ff00"
color11 = "#ffff00"
color12 = "#0000ff"
color13 = "#ff00ff"
color14 = "#00ffff"
color15 = "#ffffff"
`

	themesDir, _ := setupThemeDir(t, nil, toml)
	// Override the theme dir name to match "bare".
	themeDir := filepath.Join(themesDir, "bare")
	os.MkdirAll(themeDir, 0o755)
	writeFile(t, filepath.Join(themeDir, "palette.toml"), toml)

	home := t.TempDir()
	th, err := LoadTheme(themesDir, "bare")
	if err != nil {
		t.Fatal(err)
	}

	results := Switch(th, SwitchOpts{HomeDir: home})

	for _, r := range results {
		if r.Err != nil {
			t.Errorf("unexpected error for %s: %v", r.App, r.Err)
		}
		if !r.Skipped {
			t.Errorf("expected %s to be skipped, got message: %s", r.App, r.Message)
		}
	}
}

func TestState_RoundTrip(t *testing.T) {
	home := t.TempDir()

	// Initially no state.
	name, err := ReadState(home)
	if err != nil {
		t.Fatalf("ReadState failed: %v", err)
	}
	if name != "" {
		t.Errorf("ReadState on empty = %q, want empty", name)
	}

	// Write state.
	if err := WriteState(home, "cobalt-next-neon"); err != nil {
		t.Fatalf("WriteState failed: %v", err)
	}

	// Read back.
	name, err = ReadState(home)
	if err != nil {
		t.Fatalf("ReadState failed: %v", err)
	}
	if name != "cobalt-next-neon" {
		t.Errorf("ReadState = %q, want %q", name, "cobalt-next-neon")
	}

	// Overwrite.
	if err := WriteState(home, "dayfox"); err != nil {
		t.Fatalf("WriteState failed: %v", err)
	}
	name, err = ReadState(home)
	if err != nil {
		t.Fatalf("ReadState failed: %v", err)
	}
	if name != "dayfox" {
		t.Errorf("ReadState = %q, want %q", name, "dayfox")
	}
}

func TestDefault_RoundTrip(t *testing.T) {
	home := t.TempDir()

	// Initially no default.
	name, err := ReadDefault(home, "dark")
	if err != nil {
		t.Fatalf("ReadDefault failed: %v", err)
	}
	if name != "" {
		t.Errorf("ReadDefault on empty = %q, want empty", name)
	}

	// Write dark default.
	if err := WriteDefault(home, "dark", "cobalt-next-neon"); err != nil {
		t.Fatalf("WriteDefault dark failed: %v", err)
	}
	name, err = ReadDefault(home, "dark")
	if err != nil {
		t.Fatalf("ReadDefault dark failed: %v", err)
	}
	if name != "cobalt-next-neon" {
		t.Errorf("ReadDefault dark = %q, want %q", name, "cobalt-next-neon")
	}

	// Write light default — independent from dark.
	if err := WriteDefault(home, "light", "dayfox"); err != nil {
		t.Fatalf("WriteDefault light failed: %v", err)
	}
	name, err = ReadDefault(home, "light")
	if err != nil {
		t.Fatalf("ReadDefault light failed: %v", err)
	}
	if name != "dayfox" {
		t.Errorf("ReadDefault light = %q, want %q", name, "dayfox")
	}

	// Dark still intact.
	name, err = ReadDefault(home, "dark")
	if err != nil {
		t.Fatalf("ReadDefault dark after light write failed: %v", err)
	}
	if name != "cobalt-next-neon" {
		t.Errorf("ReadDefault dark = %q, want %q after light write", name, "cobalt-next-neon")
	}

	// Overwrite dark default.
	if err := WriteDefault(home, "dark", "tekapo-sunset-dark"); err != nil {
		t.Fatalf("WriteDefault overwrite failed: %v", err)
	}
	name, err = ReadDefault(home, "dark")
	if err != nil {
		t.Fatalf("ReadDefault after overwrite failed: %v", err)
	}
	if name != "tekapo-sunset-dark" {
		t.Errorf("ReadDefault dark = %q, want %q", name, "tekapo-sunset-dark")
	}
}

// assertFileContains reads a file and checks it contains the expected substring.
func assertFileContains(t *testing.T, path, want string) {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("reading %s: %v", path, err)
	}
	if !strings.Contains(string(data), want) {
		t.Errorf("%s does not contain %q, got:\n%s", path, want, data)
	}
}

// checkNoErrors verifies no results have errors (skipped results are fine).
func checkNoErrors(t *testing.T, results []SwitchResult) {
	t.Helper()
	for _, r := range results {
		if r.Err != nil {
			t.Errorf("unexpected error for %s: %v", r.App, r.Err)
		}
	}
}
