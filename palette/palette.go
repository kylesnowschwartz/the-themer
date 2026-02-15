// Package palette defines the canonical color palette types used by all adapters.
// It handles TOML deserialization, default derivation, and validation.
package palette

import (
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/BurntSushi/toml"
)

// hexPattern validates a 6-digit hex color string like "#050a14".
var hexPattern = regexp.MustCompile(`^#[0-9a-fA-F]{6}$`)

// Theme holds theme metadata from the [theme] TOML section.
type Theme struct {
	Name    string `toml:"name"`
	Author  string `toml:"author"`
	Variant string `toml:"variant"` // "dark" or "light"
}

// UI holds semantic color overrides from the [palette.ui] TOML section.
// When omitted, these are derived from ANSI palette colors via ApplyDefaults.
type UI struct {
	Border  string `toml:"border"`
	Dimmed  string `toml:"dimmed"`
	Accent  string `toml:"accent"`
	Success string `toml:"success"`
	Warning string `toml:"warning"`
	Error   string `toml:"error"`
	Info    string `toml:"info"`
}

// PaletteColors holds the full color palette from the [palette] TOML section.
type PaletteColors struct {
	BG          string `toml:"bg"`
	FG          string `toml:"fg"`
	Cursor      string `toml:"cursor"`
	SelectionBG string `toml:"selection_bg"`
	SelectionFG string `toml:"selection_fg"`

	Color0  string `toml:"color0"`
	Color1  string `toml:"color1"`
	Color2  string `toml:"color2"`
	Color3  string `toml:"color3"`
	Color4  string `toml:"color4"`
	Color5  string `toml:"color5"`
	Color6  string `toml:"color6"`
	Color7  string `toml:"color7"`
	Color8  string `toml:"color8"`
	Color9  string `toml:"color9"`
	Color10 string `toml:"color10"`
	Color11 string `toml:"color11"`
	Color12 string `toml:"color12"`
	Color13 string `toml:"color13"`
	Color14 string `toml:"color14"`
	Color15 string `toml:"color15"`

	UI UI `toml:"ui"`
}

// Config is the top-level TOML deserialization target.
// Every adapter receives this as its input.
type Config struct {
	Theme   Theme         `toml:"theme"`
	Palette PaletteColors `toml:"palette"`
}

// ValidationErrors collects multiple validation failures.
type ValidationErrors []string

// Error joins all validation errors into a newline-separated string.
func (ve ValidationErrors) Error() string {
	return strings.Join([]string(ve), "\n")
}

// Load reads a TOML palette file, parses it, applies defaults, and validates.
// This is the primary entry point for palette loading.
func Load(path string) (Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Config{}, fmt.Errorf("reading palette file: %w", err)
	}

	cfg, err := Parse(data)
	if err != nil {
		return Config{}, err
	}

	cfg.ApplyDefaults()

	if err := cfg.Validate(); err != nil {
		return Config{}, err
	}

	return cfg, nil
}

// Parse decodes TOML bytes into a Config struct.
func Parse(data []byte) (Config, error) {
	var cfg Config
	if err := toml.Unmarshal(data, &cfg); err != nil {
		return Config{}, fmt.Errorf("parsing TOML: %w", err)
	}
	return cfg, nil
}

// ApplyDefaults fills zero-value optional fields with values derived from
// the ANSI palette. It does not overwrite explicitly set values.
func (c *Config) ApplyDefaults() {
	p := &c.Palette

	// Palette-level defaults
	if p.Cursor == "" {
		p.Cursor = p.Color4
	}
	if p.SelectionBG == "" {
		p.SelectionBG = p.Color8
	}
	if p.SelectionFG == "" {
		p.SelectionFG = p.FG
	}

	// UI semantic defaults derived from ANSI colors
	if p.UI.Border == "" {
		p.UI.Border = p.Color8
	}
	if p.UI.Dimmed == "" {
		p.UI.Dimmed = p.Color8
	}
	if p.UI.Accent == "" {
		p.UI.Accent = p.Color6
	}
	if p.UI.Success == "" {
		p.UI.Success = p.Color2
	}
	if p.UI.Warning == "" {
		p.UI.Warning = p.Color3
	}
	if p.UI.Error == "" {
		p.UI.Error = p.Color1
	}
	if p.UI.Info == "" {
		p.UI.Info = p.Color4
	}
}

// Validate checks that all required fields are present and all hex colors
// are well-formed. Returns ValidationErrors containing all problems found,
// or nil if the config is valid.
func (c *Config) Validate() error {
	var errs ValidationErrors

	if c.Theme.Name == "" {
		errs = append(errs, "theme.name is required but not set")
	}

	// Required hex fields in a stable order for deterministic error output.
	requiredHex := []struct {
		name  string
		value string
	}{
		{"palette.bg", c.Palette.BG},
		{"palette.fg", c.Palette.FG},
		{"palette.color0", c.Palette.Color0},
		{"palette.color1", c.Palette.Color1},
		{"palette.color2", c.Palette.Color2},
		{"palette.color3", c.Palette.Color3},
		{"palette.color4", c.Palette.Color4},
		{"palette.color5", c.Palette.Color5},
		{"palette.color6", c.Palette.Color6},
		{"palette.color7", c.Palette.Color7},
		{"palette.color8", c.Palette.Color8},
		{"palette.color9", c.Palette.Color9},
		{"palette.color10", c.Palette.Color10},
		{"palette.color11", c.Palette.Color11},
		{"palette.color12", c.Palette.Color12},
		{"palette.color13", c.Palette.Color13},
		{"palette.color14", c.Palette.Color14},
		{"palette.color15", c.Palette.Color15},
	}

	for _, f := range requiredHex {
		if f.value == "" {
			errs = append(errs, fmt.Sprintf("%s is required but not set", f.name))
		} else if !hexPattern.MatchString(f.value) {
			errs = append(errs, fmt.Sprintf("%s has invalid hex format: %q (expected #RRGGBB)", f.name, f.value))
		}
	}

	// Optional hex fields: validate format only if present.
	// These have already been filled by ApplyDefaults if the caller used Load,
	// but Validate can be called independently.
	optionalHex := []struct {
		name  string
		value string
	}{
		{"palette.cursor", c.Palette.Cursor},
		{"palette.selection_bg", c.Palette.SelectionBG},
		{"palette.selection_fg", c.Palette.SelectionFG},
		{"palette.ui.border", c.Palette.UI.Border},
		{"palette.ui.dimmed", c.Palette.UI.Dimmed},
		{"palette.ui.accent", c.Palette.UI.Accent},
		{"palette.ui.success", c.Palette.UI.Success},
		{"palette.ui.warning", c.Palette.UI.Warning},
		{"palette.ui.error", c.Palette.UI.Error},
		{"palette.ui.info", c.Palette.UI.Info},
	}

	for _, f := range optionalHex {
		if f.value != "" && !hexPattern.MatchString(f.value) {
			errs = append(errs, fmt.Sprintf("%s has invalid hex format: %q (expected #RRGGBB)", f.name, f.value))
		}
	}

	if len(errs) > 0 {
		return errs
	}
	return nil
}
