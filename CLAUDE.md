# The Themer

## What This Is

A Go CLI tool that takes a TOML color palette and generates a meta-theme directory with per-app config files. A white-label theme factory -- one palette in, themed configs for ghostty, neovim, starship, etc. out.

## Development Process

You start every session with amnesia. This file and `/load-dev` are your lifeline.

1. **The user runs `/load-dev`** at the start of each session. Follow its instructions to rebuild context.
2. **This file (CLAUDE.md)** holds static, proven facts about the codebase -- architecture decisions, file layout, build commands, conventions. Trust it.
3. **`.agent-history/`** holds research, context packets, and planning docs. `/load-dev` tells you which files to read.
4. **Update this file** as decisions solidify. When something moves from "planned" to "implemented," rewrite CLAUDE.md to reflect reality, not aspirations.

### Key Files

| File | Purpose |
|------|---------|
| `.agent-history/context-packet-20260215.md` | Full project specification, goals, oracle, risks |
| `.claude/commands/load-dev.md` | Session bootstrap instructions (invoked by user via `/load-dev`) |
| `CLAUDE.md` | This file. Static project knowledge. |

## Architecture

- **Language**: Go 1.25
- **CLI framework**: spf13/cobra
- **TOML parsing**: BurntSushi/toml
- **Template engine**: Go `text/template` (used by adapters, not yet implemented)
- **Input**: TOML file with flat 16-color ANSI palette + bg/fg/cursor/selection + optional `[palette.ui]` semantic overrides
- **Output**: `{theme-name}-theme/` directory with subdirectories per app

### Project Layout

```
the-themer/
  cmd/
    root.go           # Cobra root command
    generate.go       # generate subcommand: parse -> defaults -> validate -> generate
  palette/
    palette.go        # Config, Theme, PaletteColors, UI structs; Load/Parse/ApplyDefaults/Validate
    palette_test.go   # 8 tests: parsing, defaults, validation, error accumulation
  adapter/
    adapter.go        # Adapter interface + slice-based registry (Register, All, ByName)
  testdata/
    bleu.toml         # Oracle fixture: bleu-theme's exact palette
  main.go             # Entry point; adapter imports go here
  Justfile            # just check, just build, just generate
```

### Adapter Pattern

Each app adapter implements `adapter.Adapter`:
- `Name() string` -- identifier (e.g., "ghostty")
- `DirName() string` -- output subdirectory
- `FileName(themeName string) string` -- output filename
- `Generate(cfg palette.Config) ([]byte, error)` -- render themed config

Adapters self-register via `init()`. Adding/removing an adapter = adding/removing a blank import in `main.go`.

### Palette Defaults

When optional fields are omitted, they derive from ANSI colors:

| Field | Default | Field | Default |
|-------|---------|-------|---------|
| cursor | color4 | ui.border | color8 |
| selection_bg | color8 | ui.dimmed | color8 |
| selection_fg | fg | ui.accent | color6 |
| | | ui.success | color2 |
| | | ui.warning | color3 |
| | | ui.error | color1 |
| | | ui.info | color4 |

### MVP Target Apps (not yet implemented)

1. ghostty (simplest -- nearly 1:1 palette mapping)
2. starship (TOML with color refs in style strings)
3. neovim (Lua colorscheme with highlight groups)

### Reference Material

- `/Users/kyle/Code/meta-terminal/bleu-theme/` -- gold standard output format
- `/Users/kyle/Code/dotfiles/ghostty/switch-theme.sh` -- current theme switching logic
- `/Users/kyle/Code/my-projects/cobalt-neon.nvim/` -- existing neovim theme (additional directory access configured)

## Build & Test

```bash
just check              # go vet + go test -v
just build              # go build -o the-themer .
just generate <file>    # go run . generate --input <file>

# Oracle test (once adapters exist):
# the-themer generate --input testdata/bleu.toml --output /tmp/bleu-test/
# diff /tmp/bleu-test/ghostty/bleu .cloned-sources/bleu-theme/ghostty/bleu
```

## Conventions

- Conventional commits: `feat:`, `fix:`, `refactor:`, `test:`, `docs:`
- One adapter per PR when expanding app support
- Copy bleu-theme file structure literally into templates, then parameterize
- Validate input completely before writing any output files
- Table-driven tests, ordered validation slices for deterministic output
