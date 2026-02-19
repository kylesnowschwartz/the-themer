// Package bat generates bat/cat syntax highlighter theme files.
// bat uses TextMate .tmTheme format (XML plist) for syntax coloring.
package bat

import (
	"bytes"
	"text/template"

	"github.com/kylesnowschwartz/the-themer/adapter"
	"github.com/kylesnowschwartz/the-themer/palette"
)

func init() {
	adapter.Register(&batAdapter{})
}

type batAdapter struct{}

func (b *batAdapter) Name() string                     { return "bat" }
func (b *batAdapter) DirName() string                  { return "bat" }
func (b *batAdapter) FileName(themeName string) string { return themeName + ".tmTheme" }

func (b *batAdapter) Generate(cfg palette.Config) ([]byte, error) {
	var buf bytes.Buffer
	if err := batTmpl.Execute(&buf, cfg); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// titleCase returns the theme name with first letter capitalized.
// Used for the .tmTheme display name (e.g., "bleu" -> "Bleu").
func titleCase(s string) string {
	if s == "" {
		return s
	}
	// Capitalize first byte; works for ASCII theme names.
	return string(s[0]-32) + s[1:]
}

// batTmpl renders a TextMate .tmTheme XML file from the palette.
//
// Color mapping from palette to TextMate scopes:
//
//	Global settings: BG, FG, Cursor, SelectionBG/FG, Syntax.LineHighlight
//	Structural (type, operator, punctuation, tag, property): Color4
//	Keyword/accent (keyword, attribute, boolean, link, import, decorator): UI.Accent
//	String/literal (string, regexp, code): Color5
//	Comment/dimmed (comment, quote, preprocessor): UI.Dimmed
//	Emphasis (function, heading, bold): Color15
//	Variable, parameter: FG
//	Number/constant: Syntax.Number
//	Invalid/diff-deleted: Syntax.Error
//	Diff added: Color2
//	Diff changed: UI.Accent
var batTmpl = template.Must(template.New("bat").Funcs(template.FuncMap{
	"title": titleCase,
}).Parse(`<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
	<key>name</key>
	<string>{{title .Theme.Name}}</string>
	<key>settings</key>
	<array>
		<dict>
			<key>settings</key>
			<dict>
				<key>background</key>
				<string>{{.Palette.BG}}</string>
				<key>foreground</key>
				<string>{{.Palette.FG}}</string>
				<key>caret</key>
				<string>{{.Palette.Cursor}}</string>
				<key>selection</key>
				<string>{{.Palette.SelectionBG}}</string>
				<key>selectionForeground</key>
				<string>{{.Palette.SelectionFG}}</string>
				<key>lineHighlight</key>
				<string>{{.Palette.Syntax.LineHighlight}}</string>
				<key>findHighlight</key>
				<string>{{.Palette.SelectionBG}}</string>
				<key>findHighlightForeground</key>
				<string>{{.Palette.SelectionFG}}</string>
				<key>activeGuide</key>
				<string>{{.Palette.Cursor}}</string>
				<key>bracketsForeground</key>
				<string>{{.Palette.Cursor}}</string>
				<key>bracketsOptions</key>
				<string>underline</string>
				<key>bracketContentsForeground</key>
				<string>{{.Palette.Cursor}}</string>
				<key>bracketContentsOptions</key>
				<string>underline</string>
				<key>tagsOptions</key>
				<string>stippled_underline</string>
			</dict>
		</dict>
		<dict>
			<key>name</key>
			<string>Comment</string>
			<key>scope</key>
			<string>comment, punctuation.definition.comment</string>
			<key>settings</key>
			<dict>
				<key>foreground</key>
				<string>{{.Palette.UI.Dimmed}}</string>
				<key>fontStyle</key>
				<string>italic</string>
			</dict>
		</dict>
		<dict>
			<key>name</key>
			<string>Keyword</string>
			<key>scope</key>
			<string>keyword, storage.type, storage.modifier</string>
			<key>settings</key>
			<dict>
				<key>foreground</key>
				<string>{{.Palette.UI.Accent}}</string>
				<key>fontStyle</key>
				<string>italic</string>
			</dict>
		</dict>
		<dict>
			<key>name</key>
			<string>String</string>
			<key>scope</key>
			<string>string</string>
			<key>settings</key>
			<dict>
				<key>foreground</key>
				<string>{{.Palette.Color5}}</string>
			</dict>
		</dict>
		<dict>
			<key>name</key>
			<string>Number</string>
			<key>scope</key>
			<string>constant.numeric</string>
			<key>settings</key>
			<dict>
				<key>foreground</key>
				<string>{{.Palette.Syntax.Number}}</string>
			</dict>
		</dict>
		<dict>
			<key>name</key>
			<string>Boolean</string>
			<key>scope</key>
			<string>constant.language.boolean</string>
			<key>settings</key>
			<dict>
				<key>foreground</key>
				<string>{{.Palette.UI.Accent}}</string>
			</dict>
		</dict>
		<dict>
			<key>name</key>
			<string>Constant</string>
			<key>scope</key>
			<string>constant.language, constant.character, constant.other</string>
			<key>settings</key>
			<dict>
				<key>foreground</key>
				<string>{{.Palette.Syntax.Number}}</string>
			</dict>
		</dict>
		<dict>
			<key>name</key>
			<string>Function</string>
			<key>scope</key>
			<string>entity.name.function, support.function</string>
			<key>settings</key>
			<dict>
				<key>foreground</key>
				<string>{{.Palette.Color15}}</string>
				<key>fontStyle</key>
				<string>bold</string>
			</dict>
		</dict>
		<dict>
			<key>name</key>
			<string>Type, Class</string>
			<key>scope</key>
			<string>entity.name.type, entity.name.class, support.type, support.class</string>
			<key>settings</key>
			<dict>
				<key>foreground</key>
				<string>{{.Palette.Color4}}</string>
			</dict>
		</dict>
		<dict>
			<key>name</key>
			<string>Variable</string>
			<key>scope</key>
			<string>variable, support.variable</string>
			<key>settings</key>
			<dict>
				<key>foreground</key>
				<string>{{.Palette.FG}}</string>
			</dict>
		</dict>
		<dict>
			<key>name</key>
			<string>Operator</string>
			<key>scope</key>
			<string>keyword.operator</string>
			<key>settings</key>
			<dict>
				<key>foreground</key>
				<string>{{.Palette.Color4}}</string>
			</dict>
		</dict>
		<dict>
			<key>name</key>
			<string>Punctuation</string>
			<key>scope</key>
			<string>punctuation</string>
			<key>settings</key>
			<dict>
				<key>foreground</key>
				<string>{{.Palette.Color4}}</string>
			</dict>
		</dict>
		<dict>
			<key>name</key>
			<string>Tag</string>
			<key>scope</key>
			<string>entity.name.tag, meta.tag</string>
			<key>settings</key>
			<dict>
				<key>foreground</key>
				<string>{{.Palette.Color4}}</string>
			</dict>
		</dict>
		<dict>
			<key>name</key>
			<string>Attribute</string>
			<key>scope</key>
			<string>entity.other.attribute-name</string>
			<key>settings</key>
			<dict>
				<key>foreground</key>
				<string>{{.Palette.UI.Accent}}</string>
			</dict>
		</dict>
		<dict>
			<key>name</key>
			<string>Embedded</string>
			<key>scope</key>
			<string>punctuation.section.embedded</string>
			<key>settings</key>
			<dict>
				<key>foreground</key>
				<string>{{.Palette.Color4}}</string>
			</dict>
		</dict>
		<dict>
			<key>name</key>
			<string>Invalid</string>
			<key>scope</key>
			<string>invalid</string>
			<key>settings</key>
			<dict>
				<key>foreground</key>
				<string>{{.Palette.Syntax.Error}}</string>
			</dict>
		</dict>
		<dict>
			<key>name</key>
			<string>Diff Added</string>
			<key>scope</key>
			<string>markup.inserted, meta.diff.header.to-file</string>
			<key>settings</key>
			<dict>
				<key>foreground</key>
				<string>{{.Palette.Color2}}</string>
			</dict>
		</dict>
		<dict>
			<key>name</key>
			<string>Diff Deleted</string>
			<key>scope</key>
			<string>markup.deleted, meta.diff.header.from-file</string>
			<key>settings</key>
			<dict>
				<key>foreground</key>
				<string>{{.Palette.Syntax.Error}}</string>
			</dict>
		</dict>
		<dict>
			<key>name</key>
			<string>Diff Changed</string>
			<key>scope</key>
			<string>markup.changed</string>
			<key>settings</key>
			<dict>
				<key>foreground</key>
				<string>{{.Palette.UI.Accent}}</string>
			</dict>
		</dict>
		<dict>
			<key>name</key>
			<string>Markup Heading</string>
			<key>scope</key>
			<string>markup.heading, entity.name.section</string>
			<key>settings</key>
			<dict>
				<key>foreground</key>
				<string>{{.Palette.Color15}}</string>
				<key>fontStyle</key>
				<string>bold</string>
			</dict>
		</dict>
		<dict>
			<key>name</key>
			<string>Markup Bold</string>
			<key>scope</key>
			<string>markup.bold</string>
			<key>settings</key>
			<dict>
				<key>foreground</key>
				<string>{{.Palette.Color15}}</string>
				<key>fontStyle</key>
				<string>bold</string>
			</dict>
		</dict>
		<dict>
			<key>name</key>
			<string>Markup Italic</string>
			<key>scope</key>
			<string>markup.italic</string>
			<key>settings</key>
			<dict>
				<key>foreground</key>
				<string>{{.Palette.FG}}</string>
				<key>fontStyle</key>
				<string>italic</string>
			</dict>
		</dict>
		<dict>
			<key>name</key>
			<string>Markup Link</string>
			<key>scope</key>
			<string>markup.underline.link</string>
			<key>settings</key>
			<dict>
				<key>foreground</key>
				<string>{{.Palette.UI.Accent}}</string>
			</dict>
		</dict>
		<dict>
			<key>name</key>
			<string>Markup Code</string>
			<key>scope</key>
			<string>markup.raw, markup.inline.raw</string>
			<key>settings</key>
			<dict>
				<key>foreground</key>
				<string>{{.Palette.Color5}}</string>
			</dict>
		</dict>
		<dict>
			<key>name</key>
			<string>Markup Quote</string>
			<key>scope</key>
			<string>markup.quote</string>
			<key>settings</key>
			<dict>
				<key>foreground</key>
				<string>{{.Palette.UI.Dimmed}}</string>
				<key>fontStyle</key>
				<string>italic</string>
			</dict>
		</dict>
		<dict>
			<key>name</key>
			<string>Import</string>
			<key>scope</key>
			<string>keyword.control.import, keyword.control.include</string>
			<key>settings</key>
			<dict>
				<key>foreground</key>
				<string>{{.Palette.UI.Accent}}</string>
				<key>fontStyle</key>
				<string>italic</string>
			</dict>
		</dict>
		<dict>
			<key>name</key>
			<string>Decorator</string>
			<key>scope</key>
			<string>meta.decorator, storage.type.annotation</string>
			<key>settings</key>
			<dict>
				<key>foreground</key>
				<string>{{.Palette.UI.Accent}}</string>
				<key>fontStyle</key>
				<string>italic</string>
			</dict>
		</dict>
		<dict>
			<key>name</key>
			<string>Escape Sequence</string>
			<key>scope</key>
			<string>constant.character.escape</string>
			<key>settings</key>
			<dict>
				<key>foreground</key>
				<string>{{.Palette.Color4}}</string>
				<key>fontStyle</key>
				<string>bold</string>
			</dict>
		</dict>
		<dict>
			<key>name</key>
			<string>Regular Expression</string>
			<key>scope</key>
			<string>string.regexp</string>
			<key>settings</key>
			<dict>
				<key>foreground</key>
				<string>{{.Palette.Color5}}</string>
			</dict>
		</dict>
		<dict>
			<key>name</key>
			<string>Preprocessor</string>
			<key>scope</key>
			<string>meta.preprocessor, punctuation.definition.preprocessor</string>
			<key>settings</key>
			<dict>
				<key>foreground</key>
				<string>{{.Palette.UI.Dimmed}}</string>
				<key>fontStyle</key>
				<string>italic</string>
			</dict>
		</dict>
		<dict>
			<key>name</key>
			<string>Property</string>
			<key>scope</key>
			<string>meta.property-name, entity.name.tag.localname, meta.object-literal.key</string>
			<key>settings</key>
			<dict>
				<key>foreground</key>
				<string>{{.Palette.Color4}}</string>
			</dict>
		</dict>
		<dict>
			<key>name</key>
			<string>Parameter</string>
			<key>scope</key>
			<string>variable.parameter, variable.other.parameter</string>
			<key>settings</key>
			<dict>
				<key>foreground</key>
				<string>{{.Palette.FG}}</string>
				<key>fontStyle</key>
				<string>italic</string>
			</dict>
		</dict>
		<dict>
			<key>name</key>
			<string>Built-in Function</string>
			<key>scope</key>
			<string>support.function.builtin</string>
			<key>settings</key>
			<dict>
				<key>foreground</key>
				<string>{{.Palette.UI.Accent}}</string>
				<key>fontStyle</key>
				<string>bold</string>
			</dict>
		</dict>
		<dict>
			<key>name</key>
			<string>Built-in Constant</string>
			<key>scope</key>
			<string>support.constant.builtin</string>
			<key>settings</key>
			<dict>
				<key>foreground</key>
				<string>{{.Palette.Syntax.Number}}</string>
				<key>fontStyle</key>
				<string>bold</string>
			</dict>
		</dict>
		<dict>
			<key>name</key>
			<string>Namespace</string>
			<key>scope</key>
			<string>entity.name.namespace, meta.namespace</string>
			<key>settings</key>
			<dict>
				<key>foreground</key>
				<string>{{.Palette.Color4}}</string>
				<key>fontStyle</key>
				<string>italic</string>
			</dict>
		</dict>
	</array>
	<key>uuid</key>
	<string>{{.Theme.Name}}-theme</string>
</dict>
</plist>
`))
