// Package delta generates delta git diff viewer theme files.
// Delta uses gitconfig-format named feature sections.
package delta

import (
	"bytes"
	"text/template"

	"github.com/kylesnowschwartz/the-themer/adapter"
	"github.com/kylesnowschwartz/the-themer/palette"
)

func init() {
	adapter.Register(&deltaAdapter{})
}

type deltaAdapter struct{}

func (d *deltaAdapter) Name() string                     { return "delta" }
func (d *deltaAdapter) DirName() string                  { return "delta" }
func (d *deltaAdapter) FileName(themeName string) string { return themeName + ".gitconfig" }

func (d *deltaAdapter) Generate(cfg palette.Config) ([]byte, error) {
	var buf bytes.Buffer
	if err := deltaTmpl.Execute(&buf, cfg); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// deltaTmpl renders the delta theme as a gitconfig named feature section.
//
// Output format: [delta "<name>"] section with tab-indented key-value pairs.
// Semicolon comments before the section header.
// Trailing newline after the last key-value pair.
var deltaTmpl = template.Must(template.New("delta").Parse("; {{.Theme.Name}} theme for delta\n; Add to .gitconfig or include via [include] path = <this-file>\n[delta \"{{.Theme.Name}}\"]\n\tlight = {{if eq .Theme.Variant \"light\"}}true{{else}}false{{end}}\n\tsyntax-theme = Nord\n\tnavigate = true\n\tkeep-plus-minus-markers = true\n\tfile-decoration-style = \"none\"\n\tfile-style = \"{{.Palette.Color4}} bold\"\n\tminus-style = \"{{.Palette.Color1}}\"\n\tminus-emph-style = \"{{.Palette.Color1}} bold\"\n\tplus-style = \"{{.Palette.Color2}}\"\n\tplus-emph-style = \"{{.Palette.Color2}} bold\"\n\thunk-header-style = \"{{.Palette.UI.Accent}} bold\"\n\tline-numbers = true\n\tline-numbers-minus-style = \"{{.Palette.Color1}}\"\n\tline-numbers-plus-style = \"{{.Palette.Color2}}\"\n\tline-numbers-left-style = \"{{.Palette.Color8}}\"\n\tline-numbers-right-style = \"{{.Palette.Color8}}\"\n\tline-numbers-zero-style = \"{{.Palette.UI.Dimmed}}\"\n\tzero-style = \"syntax\"\n\twhitespace-error-style = \"reverse {{.Palette.Color5}}\"\n"))
