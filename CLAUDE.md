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
| `.agent-history/phase-{1..5}-*.md` | Per-phase implementation plans |
| `.claude/commands/load-dev.md` | Session bootstrap instructions (invoked by user via `/load-dev`) |
| `CLAUDE.md` | This file. Static project knowledge. |

## Implementation Status

| Phase | Status | Description |
|-------|--------|-------------|
| 1. Foundation | **Done** | Go module, palette parsing/defaults/validation, adapter interface, CLI skeleton, Justfile |
| 2. Ghostty adapter | **Done** | Ghostty adapter, oracle test, per-adapter palette overrides, Colors() helper |
| 3. Starship adapter | Not started | TOML output with style strings |
| 4. Neovim adapter | Not started | Lua colorscheme, complex color mapping |
| 5. Polish & ship | Not started | Output directory management, error UX, README |

## Architecture

- **Language**: Go 1.25
- **CLI framework**: spf13/cobra
- **TOML parsing**: BurntSushi/toml
- **Template engine**: Go `text/template` (used by adapters)
- **Input**: TOML file with flat 16-color ANSI palette + bg/fg/cursor/selection + optional `[palette.ui]` semantic overrides
- **Output**: `{theme-name}-theme/` directory with subdirectories per app

### Project Layout

```
the-themer/
  cmd/
    root.go           # Cobra root command
    generate.go       # generate subcommand: parse -> defaults -> validate -> generate
  palette/
    palette.go        # Config, Theme, PaletteColors, UI, AdapterConfig; Load/Parse/ApplyDefaults/Validate/Colors
    palette_test.go   # 10 tests: parsing, defaults, validation, Colors(), adapter overrides
  adapter/
    adapter.go        # Adapter interface + slice-based registry (Register, All, ByName)
    ghostty/
      ghostty.go      # Ghostty adapter: text/template rendering, init() self-registration
      ghostty_test.go # Oracle test: byte-for-byte comparison against expected fixture
  testdata/
    bleu.toml         # Oracle fixture: bleu-theme's exact palette
    expected/
      ghostty/bleu    # Expected ghostty output for bleu palette
  main.go             # Entry point; adapter blank imports go here
  Justfile            # just check, just build, just generate
```

### Adapter Pattern

Each app adapter implements `adapter.Adapter`:
- `Name() string` -- identifier (e.g., "ghostty")
- `DirName() string` -- output subdirectory
- `FileName(themeName string) string` -- output filename
- `Generate(cfg palette.Config) ([]byte, error)` -- render themed config

Adapters self-register via `init()`. Adding/removing an adapter = adding/removing a blank import in `main.go`.

### Per-Adapter Palette Overrides

One TOML file per theme. Adapter-specific palettes live in `[adapters.<name>.palette]` sections. If present, the section **completely replaces** `[palette]` for that adapter. No merging, no cascading. The override goes through the same ApplyDefaults + Validate pipeline.

```toml
[palette]
# ... base colors used by all adapters ...

# Completely replaces [palette] for starship only
[adapters.starship.palette]
# ... must be a fully valid palette ...
```

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

### MVP Target Apps

1. ghostty (simplest -- nearly 1:1 palette mapping)
2. starship (TOML with color refs in style strings)
3. neovim (Lua colorscheme with highlight groups)

### Known Color Divergences

The bleu-theme reference uses different colors per-app for semantic purposes:
- Neovim `fg=#e8f4f8` vs ghostty `fg=#e0ecf4` -- deferred to Phase 4
- `#00d4ff` (accent) is brighter than ANSI cyan `#6bb6d6` -- lives in `[palette.ui]`
- Neovim error `#ff6b8a` differs from ANSI color1 `#A167A5` -- adapter-specific mapping

### Reference Material

- `/Users/kyle/Code/meta-terminal/bleu-theme/` -- gold standard output format
- `/Users/kyle/Code/dotfiles/ghostty/switch-theme.sh` -- current theme switching logic
- `/Users/kyle/Code/my-projects/cobalt-neon.nvim/` -- existing neovim theme (additional directory access configured)

## Build & Test

```bash
just check              # go vet + go test -v
just build              # go build -o the-themer .
just generate <file>    # go run . generate --input <file>

# Oracle test (ghostty):
# the-themer generate --input testdata/bleu.toml --output /tmp/bleu-test/
# diff /tmp/bleu-test/ghostty/bleu /Users/kyle/Code/meta-terminal/bleu-theme/ghostty/bleu
# Expected: only comment header line differs
```

## Conventions

- Conventional commits: `feat:`, `fix:`, `refactor:`, `test:`, `docs:`
- One adapter per PR when expanding app support
- Copy bleu-theme file structure literally into templates, then parameterize
- Validate input completely before writing any output files
- Table-driven tests, ordered validation slices for deterministic output
