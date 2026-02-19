# The Themer

## What This Is

A Go CLI tool and central warehouse for terminal environment theming. One repo holds all theme definitions — palettes, generated configs, hand-crafted configs, and external references. Three commands: `generate` (palette → per-app configs), `install` (deploy to filesystem), `switch` (change active theme across all apps, replacing switch-theme.sh).

## Development Process

You start every session with amnesia. This file and `/load-dev` are your lifeline.

1. **The user runs `/load-dev`** at the start of each session. Follow its instructions to rebuild context.
2. **This file (CLAUDE.md)** holds static, proven facts about the codebase -- architecture decisions, file layout, build commands, conventions. Trust it.
3. **`.agent-history/`** holds research, context packets, and planning docs. `/load-dev` tells you which files to read.
4. **Update this file** as decisions solidify. When something moves from "planned" to "implemented," rewrite CLAUDE.md to reflect reality, not aspirations.

### Key Files

| File | Purpose |
|------|---------|
| `.agent-history/context-packet-20260215.md` | Original project specification (palette-generator era) |
| `.agent-history/theme-switching-architecture.md` | How Kyle's multi-app theme switching works today |
| `.agent-history/nelson/pivot-to-theme-warehouse/` | Research from the warehouse pivot (scout reports, synthesis) |
| `.claude/commands/load-dev.md` | Session bootstrap instructions (invoked by user via `/load-dev`) |
| `CLAUDE.md` | This file. Static project knowledge. |

## Implementation Status

| Phase | Status | Description |
|-------|--------|-------------|
| 1. Foundation | **Done** | Go module, palette parsing/defaults/validation, adapter interface, CLI skeleton, Justfile |
| 2. Ghostty adapter | **Done** | Ghostty adapter, oracle test, per-adapter palette overrides, Colors() helper |
| 3. fzf + delta adapters | **Done** | fzf (zsh --color flags, per-adapter override), delta (gitconfig named feature) |
| 4. Theme warehouse structure | **Done** | themes/ directory, palette TOMLs, bat adapter with [palette.syntax], generated + hand-crafted configs |
| 5. Install + Switch commands | **Done** | `install` deploys to filesystem, `switch` activates theme across apps, theme/ package |
| 6. Polish & ship | Not started | Error UX, additional adapters, README |

## Architecture

- **Language**: Go 1.25
- **CLI framework**: spf13/cobra
- **TOML parsing**: BurntSushi/toml
- **Template engine**: Go `text/template` (used by generated adapters)

### The Three-Category Model

Apps fall into three categories based on how the-themer handles their configs:

1. **Generated** — Config derived from palette via template. Ghostty, fzf, delta, bat.
2. **Hand-crafted** — Config lives in the themes/ directory, checked into the repo. Starship, eza, gh-dash. Too structural to auto-derive from colors alone.
3. **Referenced** — External thing the-themer just knows the name of. Neovim colorschemes (standalone plugin repos like cobalt-neon.nvim), claude-cli ("dark"/"light").

### Theme Directory Structure

Convention over configuration. A theme is a directory. Its structure IS the manifest.

```
themes/
  cobalt-next-neon/
    palette.toml                       # Colors + metadata + references
    ghostty/cobalt-next-neon.ghostty   # Generated from palette
    fzf/cobalt-next-neon.zsh           # Generated from palette
    delta/cobalt-next-neon.gitconfig    # Generated from palette
    starship/chef-starship.toml        # Hand-crafted, checked in
    gh-dash/cobalt2.yml                # Hand-crafted, checked in
  dayfox/
    palette.toml
    ghostty/dayfox.ghostty
    fzf/dayfox.zsh
    delta/dayfox.gitconfig
    starship/chef-light-starship.toml
    gh-dash/dayfox.yml
```

Convention: `themes/<name>/<app>/` exists → this theme has configs for that app. No separate manifest file.

### palette.toml Format

```toml
[theme]
name = "cobalt-next-neon"
variant = "dark"

[palette]
bg = "#122738"
fg = "#e0ecf4"
# ... ANSI 16 + cursor/selection ...

[palette.ui]
# Semantic overrides beyond ANSI slots
accent = "#00d4ff"
# ...

[references]
neovim = "cobalt-neon"      # colorscheme name for Themery
claude = "dark"              # claude.json theme value

[adapters.fzf.palette]
# Per-adapter palette override (completely replaces [palette] for fzf)
```

### Working Themes

| Theme | Variant | Neovim | Starship | Usage |
|-------|---------|--------|----------|-------|
| cobalt-next-neon | dark | cobalt-neon (plugin repo) | chef | Primary daily driver |
| dayfox | light | dayfox (nightfox.nvim) | chef-light | Light mode |
| bleu | dark | bleu | bleu | Reference only (from meta-terminal/bleu-theme) |

### Project Layout

```
the-themer/
  cmd/
    root.go           # Cobra root command
    generate.go       # generate subcommand: parse -> defaults -> validate -> generate
    install.go        # install subcommand: deploy theme configs to filesystem
    switch.go         # switch subcommand: activate theme across all apps
  palette/
    palette.go        # Config, Theme, PaletteColors, UI, AdapterConfig, References; Load/Parse/ApplyDefaults/Validate/Colors
    palette_test.go   # 10 tests: parsing, defaults, validation, Colors(), adapter overrides
  adapter/
    adapter.go        # Adapter interface + slice-based registry (Register, All, ByName)
    ghostty/
      ghostty.go      # Ghostty adapter: text/template rendering, init() self-registration
      ghostty_test.go # Oracle test: byte-for-byte comparison against expected fixture
    fzf/
      fzf.go          # fzf adapter: zsh --color flags, per-adapter palette override
      fzf_test.go     # Oracle test + registration test
    bat/
      bat.go          # bat adapter: .tmTheme XML template, titleCase helper
      bat_test.go     # Oracle test + registration test
    delta/
      delta.go        # delta adapter: gitconfig [delta "<name>"] section
      delta_test.go   # Oracle test + registration test
  theme/
    theme.go          # Theme type, LoadTheme, ListThemes, AppDirs
    install.go        # Install() with per-app handlers (ghostty, bat, delta, fzf, starship, eza, gh-dash)
    switch.go         # Switch() with per-app handlers (+ neovim via Themery, claude.json edit)
    state.go          # ReadState/WriteState for ~/.config/the-themer/current
    theme_test.go     # 15 integration tests: install, switch, state, references, titleCase
  themes/               # Theme warehouse
    cobalt-next-neon/   # Dark theme: palette + ghostty/fzf/delta/bat + starship/gh-dash
    dayfox/             # Light theme: palette + ghostty/fzf/delta/bat + starship/gh-dash
  testdata/
    bleu.toml         # Oracle fixture: bleu-theme's exact palette + fzf adapter override
    expected/
      bat/bleu.tmTheme      # Expected bat output for bleu palette
      ghostty/bleu          # Expected ghostty output for bleu palette
      fzf/bleu.zsh          # Expected fzf output for bleu palette (with override)
      delta/bleu.gitconfig  # Expected delta output for bleu palette
  main.go             # Entry point; adapter blank imports go here
  Justfile            # just check, just build, just generate
```

### Adapter Pattern

Each generated app adapter implements `adapter.Adapter`:
- `Name() string` -- identifier (e.g., "ghostty")
- `DirName() string` -- output subdirectory
- `FileName(themeName string) string` -- output filename
- `Generate(cfg palette.Config) ([]byte, error)` -- render themed config

Adapters self-register via `init()`. Adding/removing an adapter = adding/removing a blank import in `main.go`.

### Per-Adapter Palette Overrides

One TOML file per theme. Adapter-specific palettes live in `[adapters.<name>.palette]` sections. If present, the section **completely replaces** `[palette]` for that adapter. No merging, no cascading. The override goes through the same ApplyDefaults + Validate pipeline.

### Syntax Highlighting Colors

The `[palette.syntax]` section holds colors specific to syntax highlighting (bat .tmTheme, future neovim). These differ from ANSI/UI colors -- e.g., a muted blue for numeric literals vs. the bright accent for keywords.

| Field | Default | Purpose |
|-------|---------|---------|
| number | color4 | Numeric literals and constants |
| error | color1 | Invalid/error tokens (distinct from ui.error) |
| line_highlight | selection_bg | Current line background tint |

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

### CLI Commands

```
the-themer generate --input themes/cobalt-next-neon/palette.toml --output themes/cobalt-next-neon/
the-themer install cobalt-next-neon --themes-dir ./themes/   # deploy configs to filesystem
the-themer switch cobalt-next-neon --themes-dir ./themes/    # activate theme across all apps
```

### Install Destinations

| App | Install Location |
|-----|-----------------|
| Ghostty | `~/.config/ghostty/themes/` |
| bat | `$(bat --config-dir)/themes/` + cache rebuild |
| Delta | `~/.config/the-themer/delta/` + git include.path |
| fzf | `~/.config/the-themer/fzf/` |
| Starship | `~/.config/the-themer/starship/` |
| eza | `~/.config/eza/themes/` |
| gh-dash | `~/.config/the-themer/gh-dash/` |

### Switch Mechanisms

| App | Trigger | Switch Action |
|-----|---------|---------------|
| Ghostty | `ghostty/` dir exists | Write `~/.config/ghostty/theme.local` |
| bat | `bat/` dir or `references.bat` | Write theme name to `~/.config/bat-theme.txt` |
| Delta | `delta/` dir or `references.delta` | Write feature name to `~/.config/delta-theme.txt` |
| fzf | `fzf/` dir exists | Symlink `~/.config/the-themer/fzf/current.zsh` |
| Starship | `starship/` dir exists | Symlink `~/.config/starship.toml` |
| eza | `eza/` dir exists | Symlink `~/.config/eza/theme.yml` |
| gh-dash | `gh-dash/` dir exists | Copy to `~/.config/gh-dash/config.yml` |
| Neovim | `references.neovim` set | Headless nvim + Themery |
| Claude CLI | `references.claude` set | Pure Go edit of `~/.claude.json` |

Active theme state recorded at `~/.config/the-themer/current`.

### Reference Material

- `/Users/kyle/Code/meta-terminal/bleu-theme/` -- gold standard output (22 files across 17 apps)
- `/Users/kyle/Code/dotfiles/ghostty/switch-theme.sh` -- current switching logic (to be replaced)
- `/Users/kyle/Code/dotfiles/.bash_aliases` -- bat()/git() theme wrappers (lines 100-121)
- `/Users/kyle/Code/my-projects/cobalt-neon.nvim/` -- existing neovim theme (standalone plugin)

## Build & Test

```bash
just check              # go vet + go test -v
just build              # go build -o the-themer .
just generate <file>    # go run . generate --input <file>
```

## Conventions

- Conventional commits: `feat:`, `fix:`, `refactor:`, `test:`, `docs:`
- One adapter per PR when expanding app support
- Convention-based theme directories: presence of subdirectory = app support
- Generated configs go through oracle tests (byte-for-byte fixture match)
- Validate input completely before writing any output files
- Table-driven tests, ordered validation slices for deterministic output
