# the-themer pi extension

Auto-switches a running [pi](https://github.com/badlogic/pi-mono) session's
theme between `light` and `dark` whenever `the-themer switch <name>` runs.

## How it works

`the-themer switch` writes `~/.config/the-themer/pi-variant` (atomic rename)
containing `light` or `dark` — sourced from the active theme's `[theme].variant`
in its `palette.toml`. The extension watches that file via `fs.watch` and calls
`ctx.ui.setTheme(...)` on change.

Also exposes manual slash commands:

- `/light` — switch to built-in light theme
- `/dark` — switch to built-in dark theme
- `/theme` — toggle (uses last value set by this extension)

## Install

Pi auto-discovers extensions from `~/.pi/agent/extensions/<name>/index.ts`.
Source of truth stays in this repo; symlink for discovery:

```sh
ln -s "$(pwd)/integrations/pi-extension" ~/.pi/agent/extensions/the-themer
```

Then `/reload` in pi (or restart the session).

## Opt-in per theme

A theme participates in pi auto-switching if its directory has a `pi/` subdir
(typically just containing `.keep`). Themes shipped in this repo all opt in.

## Requirements

- macOS or Linux (`fs.watch` on the the-themer config dir)
- Pi shipped by `@mariozechner/pi-coding-agent`
