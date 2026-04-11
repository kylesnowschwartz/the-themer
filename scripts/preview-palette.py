#!/usr/bin/env python3
# /// script
# requires-python = ">=3.11"
# dependencies = []
# ///
"""Render palette TOML swatches in the terminal using true-color escapes.

Usage:
    uv run scripts/preview-palette.py themes/cobalt-next-neon/palette.toml
    uv run scripts/preview-palette.py themes/cobalt-next-neon/palette.toml themes/cobalt-next-neon-v2/palette.toml
"""

import argparse
import tomllib

ANSI_NAMES = [
    "black", "red", "green", "yellow", "blue", "magenta", "cyan", "white",
    "br.black", "br.red", "br.green", "br.yello",
    "br.blue", "br.mgnta", "br.cyan", "br.white",
]

RST = "\033[0m"

# Each color column is exactly this many visible characters wide.
COL = 9


def hex_to_rgb(h: str) -> tuple[int, int, int]:
    h = h.lstrip("#")
    return int(h[0:2], 16), int(h[2:4], 16), int(h[4:6], 16)


def fg_esc(hex_color: str) -> str:
    r, g, b = hex_to_rgb(hex_color)
    return f"\033[38;2;{r};{g};{b}m"


def bg_esc(hex_color: str) -> str:
    r, g, b = hex_to_rgb(hex_color)
    return f"\033[48;2;{r};{g};{b}m"


def load_toml(path: str) -> dict:
    with open(path, "rb") as f:
        return tomllib.load(f)


def get_colors(data: dict) -> dict:
    """Extract palette info from TOML data."""
    theme = data.get("theme", {})
    pal = data.get("palette", {})
    result = {
        "name": theme.get("name", "unknown"),
        "variant": theme.get("variant", "dark"),
        "bg": pal.get("bg", "#000000"),
        "fg": pal.get("fg", "#ffffff"),
        "cursor": pal.get("cursor", ""),
        "selection_bg": pal.get("selection_bg", ""),
        "selection_fg": pal.get("selection_fg", ""),
        "colors": {},
    }
    for i in range(16):
        val = pal.get(f"color{i}", "")
        if val:
            result["colors"][i] = val
    return result


def pad(visible_text: str) -> str:
    """Pad visible text to COL width."""
    return visible_text + " " * (COL - len(visible_text))


def render_color_row(indices: list[int], pal: dict, theme_bg: str) -> None:
    """Render swatch blocks, names, and hex values for a range of ANSI slots."""
    colors = pal["colors"]
    bg_hex = pal["bg"]
    dim = fg_esc("#555555")

    # Row 1: swatch blocks
    line = f"  {theme_bg} "
    for i in indices:
        hex_val = colors.get(i, "")
        if hex_val:
            line += f" {bg_esc(hex_val)}{' ' * COL}{RST}{theme_bg}"
        else:
            line += f" {' ' * COL}"
    line += f"  {RST}"
    print(line)

    # Row 2: color name labels (in that color)
    line = f"  {theme_bg} "
    for i in indices:
        hex_val = colors.get(i, "")
        if hex_val:
            name = pad(ANSI_NAMES[i])
            line += f" {fg_esc(hex_val)}{name}{RST}{theme_bg}"
        else:
            line += f" {' ' * COL}"
    line += f"  {RST}"
    print(line)

    # Row 3: hex values (dimmed)
    line = f"  {theme_bg} "
    for i in indices:
        hex_val = colors.get(i, "")
        if hex_val:
            h = pad(hex_val.lower())
            line += f" {dim}{h}{RST}{theme_bg}"
        else:
            line += f" {' ' * COL}"
    line += f"  {RST}"
    print(line)


def render_special_row(specials: list[tuple[str, str]], pal: dict,
                       theme_bg: str) -> None:
    """Render swatch blocks and labels for special colors."""
    dim = fg_esc("#555555")

    entries = [(label, h) for label, h in specials if h]

    # Row 1: swatch blocks
    line = f"  {theme_bg} "
    for _, hex_val in entries:
        line += f" {bg_esc(hex_val)}{' ' * COL}{RST}{theme_bg}"
    line += f"  {RST}"
    print(line)

    # Row 2: labels (in that color)
    line = f"  {theme_bg} "
    for label, hex_val in entries:
        name = pad(label)
        line += f" {fg_esc(hex_val)}{name}{RST}{theme_bg}"
    line += f"  {RST}"
    print(line)

    # Row 3: hex values
    line = f"  {theme_bg} "
    for _, hex_val in entries:
        h = pad(hex_val.lower())
        line += f" {dim}{h}{RST}{theme_bg}"
    line += f"  {RST}"
    print(line)


def blank_line(theme_bg: str, width: int = 80) -> None:
    print(f"  {theme_bg}{' ' * width}{RST}")


def render_palette(pal: dict) -> None:
    """Render a single palette preview."""
    theme_bg = bg_esc(pal["bg"])
    theme_fg = fg_esc(pal["fg"])
    width = 80

    # Header
    title = f" {pal['name']}  ({pal['variant']})  bg: {pal['bg']}  fg: {pal['fg']} "
    print()
    print(f"  {theme_bg}{theme_fg}{title}{' ' * (width - len(title))}{RST}")
    blank_line(theme_bg, width)

    # Section: normals
    label_line = f"  {theme_bg}{fg_esc('#888888')} normals{' ' * (width - 8)}{RST}"
    print(label_line)
    render_color_row(list(range(8)), pal, theme_bg)
    blank_line(theme_bg, width)

    # Section: brights
    label_line = f"  {theme_bg}{fg_esc('#888888')} brights{' ' * (width - 8)}{RST}"
    print(label_line)
    render_color_row(list(range(8, 16)), pal, theme_bg)
    blank_line(theme_bg, width)

    # Section: special
    specials = [
        ("fg", pal["fg"]),
        ("cursor", pal.get("cursor", "")),
        ("sel_bg", pal.get("selection_bg", "")),
        ("sel_fg", pal.get("selection_fg", "")),
    ]
    label_line = f"  {theme_bg}{fg_esc('#888888')} special{' ' * (width - 8)}{RST}"
    print(label_line)
    render_special_row(specials, pal, theme_bg)
    blank_line(theme_bg, width)

    # Section: code sample
    c = pal["colors"]
    kw = fg_esc(c.get(4, pal["fg"]))
    fn = fg_esc(pal["fg"])
    st = fg_esc(c.get(3, "#e9e75c"))
    cm = fg_esc(c.get(8, "#555555"))
    mg = fg_esc(c.get(5, "#888888"))
    cy = fg_esc(c.get(6, "#5fced8"))
    wh = fg_esc(c.get(7, "#bbbbbb"))
    rd = fg_esc(c.get(1, "#ff0000"))
    br_mg = fg_esc(c.get(13, c.get(5, "#888888")))

    def code(text: str, visible_len: int) -> None:
        """Print a code line padded to fill the bg block."""
        padding = " " * (width - 4 - visible_len)
        print(f"  {theme_bg}    {text}{RST}{theme_bg}{padding}{RST}")

    label_line = f"  {theme_bg}{fg_esc('#888888')} code{' ' * (width - 5)}{RST}"
    print(label_line)
    blank_line(theme_bg, width)
    code(f"{cm}# Process incoming theme palettes", 34)
    code(f"{br_mg}@validator.check", 16)
    code(f"{kw}def {fn}load_palette{wh}({mg}self{wh}, {fn}path{wh}: {cy}str{wh}) -> {cy}Config{wh}:", 42)
    code(f"    {st}\"\"\"Load and validate a TOML palette.\"\"\"", 43)
    code(f"    {kw}with {fn}open{wh}({fn}path{wh}, {st}\"rb\"{wh}) {kw}as {fn}f{wh}:", 27)
    code(f"        {mg}self{wh}.{fn}data {wh}= {fn}tomllib{wh}.{fn}load{wh}({fn}f{wh})", 29)
    code(f"    {kw}if {kw}not {mg}self{wh}.{fn}validate{wh}():", 25)
    code(f"        {kw}raise {rd}PaletteError{wh}({st}f\"Invalid: {wh}{{{fn}path{wh}}}{st}\"{wh})", 38)
    code(f"    {kw}return {mg}self{wh}.{fn}config", 18)
    blank_line(theme_bg, width)


def main() -> None:
    parser = argparse.ArgumentParser(
        description="Preview palette TOML colors in the terminal")
    parser.add_argument("palettes", nargs="+", metavar="TOML",
                        help="Palette TOML file(s) to preview")
    args = parser.parse_args()

    for path in args.palettes:
        data = load_toml(path)
        pal = get_colors(data)
        render_palette(pal)

    print()


if __name__ == "__main__":
    main()
