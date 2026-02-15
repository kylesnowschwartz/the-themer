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

## Architecture (Planned -- not yet implemented)

- **Language**: Go
- **CLI framework**: TBD (cobra or similar)
- **Template engine**: Go `text/template`
- **Input**: TOML file with flat 16-color ANSI palette + bg/fg/cursor/selection + optional semantic overrides
- **Output**: `{theme-name}-theme/` directory with subdirectories per app

### Adapter Pattern

Each app adapter is a self-contained unit that:
1. Receives a canonical `Palette` struct
2. Knows its output directory name and filename
3. Renders its config via Go template
4. Can be registered/deregistered from the generator

### MVP Target Apps

1. ghostty (simplest -- nearly 1:1 palette mapping)
2. starship (TOML with color refs in style strings)
3. neovim (Lua colorscheme with highlight groups)

### Reference Material

- `/Users/kyle/Code/meta-terminal/bleu-theme/` -- gold standard output format
- `/Users/kyle/Code/dotfiles/ghostty/switch-theme.sh` -- current theme switching logic
- `/Users/kyle/Code/my-projects/cobalt-neon.nvim/` -- existing neovim theme (additional directory access configured)

## Build & Test

```bash
# Build (once Go code exists)
go build ./...

# Test
go test ./...

# Oracle test: generate bleu theme and diff against reference
# the-themer generate --input testdata/bleu.toml --output /tmp/bleu-test/
# diff /tmp/bleu-test/ghostty/bleu /path/to/bleu-theme/ghostty/bleu
```

## Conventions

- Conventional commits: `feat:`, `fix:`, `refactor:`, `test:`, `docs:`
- One adapter per PR when expanding app support
- Copy bleu-theme file structure literally into templates, then parameterize
- Validate input completely before writing any output files
