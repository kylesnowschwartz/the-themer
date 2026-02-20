// Package ghostty generates Ghostty terminal theme files.
// Ghostty uses a simple key-value format with palette index assignments.
package ghostty

import (
	"bytes"
	"text/template"

	"github.com/kylesnowschwartz/the-themer/adapter"
	"github.com/kylesnowschwartz/the-themer/palette"
)

func init() {
	adapter.Register(&ghosttyAdapter{})
}

type ghosttyAdapter struct{}

func (g *ghosttyAdapter) Name() string                     { return "ghostty" }
func (g *ghosttyAdapter) DirName() string                  { return "ghostty" }
func (g *ghosttyAdapter) FileName(themeName string) string { return themeName + ".ghostty" }

func (g *ghosttyAdapter) Generate(cfg palette.Config) ([]byte, error) {
	var buf bytes.Buffer
	if err := ghosttyTmpl.Execute(&buf, cfg); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// ghosttyTmpl renders the Ghostty theme config.
//
// Whitespace layout:
//   - Blank line after comment header (newline before range body provides it)
//   - 16 palette lines, no blank lines between them
//   - Blank line before terminal settings
//   - Blank line before transparency section
//   - No trailing newline (file ends after "0.7")
var ghosttyTmpl = template.Must(template.New("ghostty").Parse(
	`# {{.Theme.Name}} theme for Ghostty
{{range $i, $c := .Palette.Colors}}
palette = {{$i}}={{$c}}{{end}}

background = {{.Palette.BG}}
foreground = {{.Palette.FG}}
cursor-color = {{.Palette.Cursor}}
cursor-text = {{.Palette.FG}}
selection-background = {{.Palette.SelectionBG}}
selection-foreground = {{.Palette.SelectionFG}}

# Transparency and blur effects
unfocused-split-opacity = 0.7`))
