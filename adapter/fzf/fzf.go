// Package fzf generates fzf color configuration files for zsh.
// fzf uses --color flags appended to FZF_DEFAULT_OPTS.
package fzf

import (
	"bytes"
	"text/template"

	"github.com/kylesnowschwartz/the-themer/adapter"
	"github.com/kylesnowschwartz/the-themer/palette"
)

func init() {
	adapter.Register(&fzfAdapter{})
}

type fzfAdapter struct{}

func (f *fzfAdapter) Name() string                     { return "fzf" }
func (f *fzfAdapter) DirName() string                  { return "fzf" }
func (f *fzfAdapter) FileName(themeName string) string { return themeName + ".zsh" }

func (f *fzfAdapter) Generate(cfg palette.Config) ([]byte, error) {
	var buf bytes.Buffer
	if err := fzfTmpl.Execute(&buf, cfg); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// fzfTmpl renders the fzf color configuration for zsh.
//
// Output format: export FZF_DEFAULT_OPTS appended with --color flags.
// 18 color parameters across 7 --color lines.
// Trailing newline after closing single quote.
var fzfTmpl = template.Must(template.New("fzf").Parse(
	`#!/bin/zsh
# {{.Theme.Name}} theme for fzf

export FZF_DEFAULT_OPTS=$FZF_DEFAULT_OPTS'
  --color=fg:{{.Palette.FG}},bg:{{.Palette.BG}},hl:{{.Palette.UI.Accent}}
  --color=fg+:{{.Palette.Color15}},bg+:{{.Palette.SelectionBG}},hl+:{{.Palette.UI.Accent}}
  --color=info:{{.Palette.UI.Info}},prompt:{{.Palette.Color4}},pointer:{{.Palette.UI.Accent}}
  --color=marker:{{.Palette.UI.Success}},spinner:{{.Palette.Color4}},header:{{.Palette.UI.Dimmed}}
  --color=border:{{.Palette.UI.Border}},gutter:{{.Palette.BG}}
  --color=query:{{.Palette.FG}},disabled:{{.Palette.UI.Dimmed}}
  --color=preview-fg:{{.Palette.FG}},preview-bg:{{.Palette.UI.Border}}
'
`))
