package palette_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/kylesnowschwartz/the-themer/palette"
)

func TestParseValidTOML(t *testing.T) {
	cfg, err := palette.Load("../testdata/bleu.toml")
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	t.Run("theme metadata", func(t *testing.T) {
		assertEqual(t, "theme.name", cfg.Theme.Name, "bleu")
		assertEqual(t, "theme.author", cfg.Theme.Author, "bnema")
		assertEqual(t, "theme.variant", cfg.Theme.Variant, "dark")
	})

	t.Run("palette basics", func(t *testing.T) {
		assertEqual(t, "bg", cfg.Palette.BG, "#050a14")
		assertEqual(t, "fg", cfg.Palette.FG, "#e0ecf4")
		assertEqual(t, "cursor", cfg.Palette.Cursor, "#5588cc")
		assertEqual(t, "selection_bg", cfg.Palette.SelectionBG, "#2d4a6b")
		assertEqual(t, "selection_fg", cfg.Palette.SelectionFG, "#e0ecf4")
	})

	t.Run("ANSI colors", func(t *testing.T) {
		assertEqual(t, "color0", cfg.Palette.Color0, "#050a14")
		assertEqual(t, "color1", cfg.Palette.Color1, "#A167A5")
		assertEqual(t, "color2", cfg.Palette.Color2, "#99FFE4")
		assertEqual(t, "color3", cfg.Palette.Color3, "#FDBD85")
		assertEqual(t, "color4", cfg.Palette.Color4, "#5588cc")
		assertEqual(t, "color5", cfg.Palette.Color5, "#87ceeb")
		assertEqual(t, "color6", cfg.Palette.Color6, "#6bb6d6")
		assertEqual(t, "color7", cfg.Palette.Color7, "#e0ecf4")
		assertEqual(t, "color8", cfg.Palette.Color8, "#2d4a6b")
		assertEqual(t, "color9", cfg.Palette.Color9, "#A167A5")
		assertEqual(t, "color10", cfg.Palette.Color10, "#99FFE4")
		assertEqual(t, "color11", cfg.Palette.Color11, "#FDBD85")
		assertEqual(t, "color12", cfg.Palette.Color12, "#5588cc")
		assertEqual(t, "color13", cfg.Palette.Color13, "#87ceeb")
		assertEqual(t, "color14", cfg.Palette.Color14, "#6bb6d6")
		assertEqual(t, "color15", cfg.Palette.Color15, "#fefefe")
	})

	t.Run("UI overrides", func(t *testing.T) {
		assertEqual(t, "ui.border", cfg.Palette.UI.Border, "#2d4a6b")
		assertEqual(t, "ui.dimmed", cfg.Palette.UI.Dimmed, "#708090")
		assertEqual(t, "ui.accent", cfg.Palette.UI.Accent, "#00d4ff")
		assertEqual(t, "ui.success", cfg.Palette.UI.Success, "#99FFE4")
		assertEqual(t, "ui.warning", cfg.Palette.UI.Warning, "#FDBD85")
		assertEqual(t, "ui.error", cfg.Palette.UI.Error, "#A167A5")
		assertEqual(t, "ui.info", cfg.Palette.UI.Info, "#87ceeb")
	})
}

func TestDefaults(t *testing.T) {
	tomlStr := `
[theme]
name = "minimal"

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

	cfg, err := palette.Parse([]byte(tomlStr))
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	cfg.ApplyDefaults()

	// Palette-level defaults
	assertEqual(t, "cursor", cfg.Palette.Cursor, "#0000aa")            // color4
	assertEqual(t, "selection_bg", cfg.Palette.SelectionBG, "#555555") // color8
	assertEqual(t, "selection_fg", cfg.Palette.SelectionFG, "#eeeeee") // fg

	// UI defaults from ANSI colors
	assertEqual(t, "ui.border", cfg.Palette.UI.Border, "#555555")   // color8
	assertEqual(t, "ui.dimmed", cfg.Palette.UI.Dimmed, "#555555")   // color8
	assertEqual(t, "ui.accent", cfg.Palette.UI.Accent, "#00aaaa")   // color6
	assertEqual(t, "ui.success", cfg.Palette.UI.Success, "#00aa00") // color2
	assertEqual(t, "ui.warning", cfg.Palette.UI.Warning, "#aaaa00") // color3
	assertEqual(t, "ui.error", cfg.Palette.UI.Error, "#aa0000")     // color1
	assertEqual(t, "ui.info", cfg.Palette.UI.Info, "#0000aa")       // color4
}

func TestDefaults_DoNotOverrideExplicit(t *testing.T) {
	tomlStr := `
[theme]
name = "explicit"

[palette]
bg = "#111111"
fg = "#eeeeee"
cursor = "#ff0000"
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

[palette.ui]
accent = "#abcdef"
error = "#fedcba"
`

	cfg, err := palette.Parse([]byte(tomlStr))
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	cfg.ApplyDefaults()

	// Explicit values must survive
	assertEqual(t, "cursor", cfg.Palette.Cursor, "#ff0000")
	assertEqual(t, "ui.accent", cfg.Palette.UI.Accent, "#abcdef")
	assertEqual(t, "ui.error", cfg.Palette.UI.Error, "#fedcba")

	// Other UI fields still get defaults
	assertEqual(t, "ui.success", cfg.Palette.UI.Success, "#00aa00") // color2
}

func TestValidation_MissingRequired(t *testing.T) {
	tests := []struct {
		name    string
		remove  string
		wantMsg string
	}{
		{"missing theme.name", "theme.name", "theme.name"},
		{"missing bg", "palette.bg", "palette.bg"},
		{"missing fg", "palette.fg", "palette.fg"},
		{"missing color0", "palette.color0", "palette.color0"},
		{"missing color2", "palette.color2", "palette.color2"},
		{"missing color8", "palette.color8", "palette.color8"},
		{"missing color15", "palette.color15", "palette.color15"},
	}

	baseTOML := `
[theme]
name = "test"

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

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Remove the target field by replacing its line
			modified := removeTOMLField(baseTOML, tc.remove)

			cfg, err := palette.Parse([]byte(modified))
			if err != nil {
				t.Fatalf("Parse failed: %v", err)
			}

			cfg.ApplyDefaults()
			err = cfg.Validate()

			if err == nil {
				t.Fatalf("expected validation error for %s, got nil", tc.remove)
			}
			if !strings.Contains(err.Error(), tc.wantMsg) {
				t.Errorf("error %q does not mention %q", err.Error(), tc.wantMsg)
			}
		})
	}
}

func TestValidation_InvalidHex(t *testing.T) {
	tests := []struct {
		name    string
		value   string
		wantMsg string
	}{
		{"no hash prefix", "050a14", "invalid hex format"},
		{"short hex", "#fff", "invalid hex format"},
		{"too long", "#050a14ff", "invalid hex format"},
		{"non-hex chars", "#gggggg", "invalid hex format"},
		{"empty hash", "#", "invalid hex format"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			tomlStr := `
[theme]
name = "test"

[palette]
bg = "` + tc.value + `"
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
			cfg, err := palette.Parse([]byte(tomlStr))
			if err != nil {
				t.Fatalf("Parse failed: %v", err)
			}

			cfg.ApplyDefaults()
			err = cfg.Validate()

			if err == nil {
				t.Fatalf("expected validation error for %q, got nil", tc.value)
			}
			if !strings.Contains(err.Error(), tc.wantMsg) {
				t.Errorf("error %q does not contain %q", err.Error(), tc.wantMsg)
			}
			if !strings.Contains(err.Error(), "palette.bg") {
				t.Errorf("error %q does not mention field name palette.bg", err.Error())
			}
		})
	}
}

func TestValidation_CollectsAllErrors(t *testing.T) {
	// Missing color2, color5, and invalid hex for bg -- three errors total.
	tomlStr := `
[theme]
name = "broken"

[palette]
bg = "not-a-color"
fg = "#eeeeee"
color0 = "#000000"
color1 = "#aa0000"
color3 = "#aaaa00"
color4 = "#0000aa"
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

	cfg, err := palette.Parse([]byte(tomlStr))
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	cfg.ApplyDefaults()
	err = cfg.Validate()

	if err == nil {
		t.Fatal("expected validation errors, got nil")
	}

	var ve palette.ValidationErrors
	if !errors.As(err, &ve) {
		t.Fatalf("expected ValidationErrors, got %T: %v", err, err)
	}

	if len(ve) != 3 {
		t.Errorf("expected 3 errors, got %d: %v", len(ve), ve)
	}

	// Verify each missing field is mentioned
	errStr := err.Error()
	for _, want := range []string{"palette.bg", "palette.color2", "palette.color5"} {
		if !strings.Contains(errStr, want) {
			t.Errorf("errors do not mention %q: %s", want, errStr)
		}
	}
}

func TestValidation_ThemeNameRequired(t *testing.T) {
	tomlStr := `
[theme]
author = "nobody"

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

	cfg, err := palette.Parse([]byte(tomlStr))
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	cfg.ApplyDefaults()
	err = cfg.Validate()

	if err == nil {
		t.Fatal("expected error for missing theme.name, got nil")
	}
	if !strings.Contains(err.Error(), "theme.name") {
		t.Errorf("error %q does not mention theme.name", err.Error())
	}
}

func TestParseInvalidTOML(t *testing.T) {
	_, err := palette.Parse([]byte(`this is [not valid toml`))
	if err == nil {
		t.Fatal("expected parse error for invalid TOML, got nil")
	}
}

// assertEqual is a test helper that reports field mismatches.
func assertEqual(t *testing.T, field, got, want string) {
	t.Helper()
	if got != want {
		t.Errorf("%s: got %q, want %q", field, got, want)
	}
}

// removeTOMLField removes a field from TOML by matching its key.
// Handles both "key = value" lines and [section] headers.
func removeTOMLField(toml, field string) string {
	lines := strings.Split(toml, "\n")
	var result []string

	// For "theme.name", remove the name line under [theme]
	// For "palette.bg", remove the bg line under [palette]
	parts := strings.SplitN(field, ".", 2)
	section := parts[0]
	key := ""
	if len(parts) > 1 {
		key = parts[1]
	}

	if key == "" {
		// Remove entire section -- not needed for current tests
		return toml
	}

	inSection := false
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Track which section we're in
		if strings.HasPrefix(trimmed, "[") {
			sectionName := strings.Trim(trimmed, "[] ")
			inSection = sectionName == section
		}

		// Skip the target field line
		if inSection && strings.HasPrefix(trimmed, key+" ") || inSection && strings.HasPrefix(trimmed, key+"=") {
			continue
		}

		result = append(result, line)
	}

	return strings.Join(result, "\n")
}
