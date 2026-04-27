// Package opensessions generates theme files for the opensessions panel/tmux header.
//
// Opensessions reads ~/.config/opensessions/active-theme.json at startup and
// watches it via fs.watch. Missing palette tokens fall through to the runtime's
// builtin defaults via merge, so emitting only the tokens we map below is safe.
//
// The contract — locked on the opensessions side — is:
//
//	{
//	  "name":    "<theme>",
//	  "variant": "light" | "dark",
//	  "palette": {
//	    "text": "#...", "subtext0": "#...", ...,
//	    "base": "#...", "mantle": "#...", "crust": "#..."
//	  }
//	}
//
// The 21 palette tokens are derived from palette.toml per the table in
// CLAUDE.md (Switch Mechanisms / opensessions). Background tiers (base/
// mantle/crust) emit the palette's bg hex so the panel matches Ghostty's bg
// at switch time. The original contract specified the literal "transparent"
// for these tokens but OpenTUI's transparent rendering (PR #932) doesn't yet
// reach the panel's outer container; emitting hex avoids a black panel under
// light themes.
package opensessions

import (
	"bytes"
	"encoding/json"

	"github.com/kylesnowschwartz/the-themer/adapter"
	"github.com/kylesnowschwartz/the-themer/palette"
)

func init() {
	adapter.Register(&opensessionsAdapter{})
}

type opensessionsAdapter struct{}

func (o *opensessionsAdapter) Name() string                     { return "opensessions" }
func (o *opensessionsAdapter) DirName() string                  { return "opensessions" }
func (o *opensessionsAdapter) FileName(themeName string) string { return themeName + ".json" }

// osPalette mirrors the 21-token shape opensessions's loadExternalTheme expects.
// Field order is significant: encoding/json emits in declaration order, and
// the oracle test asserts byte equality.
type osPalette struct {
	Text     string `json:"text"`
	Subtext0 string `json:"subtext0"`
	Subtext1 string `json:"subtext1"`
	Overlay0 string `json:"overlay0"`
	Overlay1 string `json:"overlay1"`
	Blue     string `json:"blue"`
	Lavender string `json:"lavender"`
	Pink     string `json:"pink"`
	Mauve    string `json:"mauve"`
	Yellow   string `json:"yellow"`
	Green    string `json:"green"`
	Red      string `json:"red"`
	Peach    string `json:"peach"`
	Teal     string `json:"teal"`
	Sky      string `json:"sky"`
	Surface0 string `json:"surface0"`
	Surface1 string `json:"surface1"`
	Surface2 string `json:"surface2"`
	Base     string `json:"base"`
	Mantle   string `json:"mantle"`
	Crust    string `json:"crust"`
}

type osTheme struct {
	Name    string    `json:"name"`
	Variant string    `json:"variant"`
	Palette osPalette `json:"palette"`
}

// Background tiers (base/mantle/crust) emit the palette's bg hex rather than
// the literal "transparent" the JSON contract documents. OpenTUI's transparent
// support (upstream PR #932) only handles BoxRenderable bg/border, and the
// opensessions panel's outer container falls through to a builtin dark default
// when given "transparent" — manifesting as a black panel under light themes.
// Emitting bg hex gives an opaque panel that matches Ghostty's bg colour at
// switch time. Once OpenTUI's transparent handling is plumbed through the panel
// renderable, this can revert to the literal "transparent" per the original
// contract.

func (o *opensessionsAdapter) Generate(cfg palette.Config) ([]byte, error) {
	p := cfg.Palette // post-ApplyDefaults: every UI/Selection field is populated.

	// Spec fallback for "blue" diverges from ApplyDefaults:
	//   ui.accent (if user-set) else color4   — vs. ApplyDefaults' color6.
	// Opensessions's "blue" token drives the working severity + accent pill
	// where a real blue reads better than the UI.Accent's color6 default.
	blue := cfg.RawPalette.UI.Accent
	if blue == "" {
		blue = p.Color4
	}

	t := osTheme{
		Name:    cfg.Theme.Name,
		Variant: cfg.Theme.Variant,
		Palette: osPalette{
			Text:     p.FG,
			Subtext0: p.UI.Dimmed,
			Subtext1: p.UI.Dimmed,
			Overlay0: p.Color8,
			Overlay1: p.Color8,
			Blue:     blue,
			Lavender: p.Color5,
			Pink:     p.Color5,
			Mauve:    p.Color5,
			Yellow:   p.UI.Warning,
			Green:    p.UI.Success,
			Red:      p.UI.Error,
			Peach:    p.Color3,
			Teal:     p.Color6,
			Sky:      p.Color6,
			Surface0: p.SelectionBG,
			Surface1: p.SelectionBG,
			Surface2: p.Color8,
			Base:     p.BG,
			Mantle:   p.BG,
			Crust:    p.BG,
		},
	}

	// json.Encoder.Encode appends a trailing newline; SetIndent gives a
	// 2-space pretty-printed layout that matches the rest of the warehouse.
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetIndent("", "  ")
	if err := enc.Encode(t); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
