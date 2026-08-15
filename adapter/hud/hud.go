// Package hud generates theme files for tail-claude-hud, the Claude Code
// statusline. The output is a TOML file with one top-level table per widget,
// each carrying an fg color — the same shape as tail-claude-hud's
// [theme.overrides] map. the-themer's switch command deploys it to
// ~/.config/tail-claude-hud/theme-active.toml, which the statusline layers
// over its resolved built-in theme on every tick.
package hud

import (
	"bytes"
	"text/template"

	"github.com/kylesnowschwartz/the-themer/adapter"
	"github.com/kylesnowschwartz/the-themer/palette"
)

func init() {
	adapter.Register(&hudAdapter{})
}

type hudAdapter struct{}

func (h *hudAdapter) Name() string                     { return "hud" }
func (h *hudAdapter) DirName() string                  { return "hud" }
func (h *hudAdapter) FileName(themeName string) string { return themeName + ".toml" }

func (h *hudAdapter) Generate(cfg palette.Config) ([]byte, error) {
	var buf bytes.Buffer
	if err := hudTmpl.Execute(&buf, cfg); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// hudTmpl renders the widget color map. Backgrounds are omitted (transparent,
// terminal default) so the statusline sits on the terminal bg like the hud's
// own default theme. Hue intent follows the hud default theme: model/session
// cyan-ish, context/agents green, directory/project/tools blue, env magenta,
// todos/thinking yellow, duration dimmed.
var hudTmpl = template.Must(template.New("hud").Parse(`# {{.Theme.Name}} theme for tail-claude-hud
# Deployed by the-themer to ~/.config/tail-claude-hud/theme-active.toml

[model]
fg = "{{.Palette.UI.Accent}}"

[context]
fg = "{{.Palette.UI.Success}}"

[directory]
fg = "{{.Palette.Color4}}"

[git]
fg = "{{.Palette.Color6}}"

[project]
fg = "{{.Palette.Color4}}"

[env]
fg = "{{.Palette.Color5}}"

[duration]
fg = "{{.Palette.UI.Dimmed}}"

[tools]
fg = "{{.Palette.Color4}}"

[agents]
fg = "{{.Palette.Color2}}"

[todos]
fg = "{{.Palette.Color3}}"

[session]
fg = "{{.Palette.Color6}}"

[thinking]
fg = "{{.Palette.Color3}}"

[cost]
fg = "{{.Palette.Color6}}"
`))
