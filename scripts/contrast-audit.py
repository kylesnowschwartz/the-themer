#!/usr/bin/env python3
"""WCAG 2.1 contrast ratio audit for the-themer palette TOML files.

Usage:
    python3 scripts/contrast-audit.py themes/cobalt-next-neon/palette.toml
    python3 scripts/contrast-audit.py themes/*/palette.toml

Checks all 16 ANSI colors against the palette's bg color and reports:
- PASS (AA)   : >= 4.5:1 (body text safe)
- PASS (3:1+) : >= 3.0:1 (large text / UI elements)
- FAIL        : < 3.0:1  (unreadable on background)

Colors matching the bg exactly (e.g., color0 in dark, color15 in light)
are marked as expected and excluded from failure counts.
"""

import sys

try:
    import tomllib
except ImportError:
    try:
        import tomli as tomllib
    except ImportError:
        print("Requires Python 3.11+ (tomllib) or `pip install tomli`")
        sys.exit(1)


def hex_to_rgb(h: str) -> tuple[float, float, float]:
    h = h.lstrip("#")
    return tuple(int(h[i : i + 2], 16) / 255.0 for i in (0, 2, 4))


def linearize(c: float) -> float:
    return c / 12.92 if c <= 0.03928 else ((c + 0.055) / 1.055) ** 2.4


def relative_luminance(hex_color: str) -> float:
    r, g, b = hex_to_rgb(hex_color)
    return 0.2126 * linearize(r) + 0.7152 * linearize(g) + 0.0722 * linearize(b)


def contrast_ratio(hex1: str, hex2: str) -> float:
    l1 = relative_luminance(hex1)
    l2 = relative_luminance(hex2)
    return (max(l1, l2) + 0.05) / (min(l1, l2) + 0.05)


# ANSI color descriptions
ANSI_NAMES = {
    "color0": "black",
    "color1": "red",
    "color2": "green",
    "color3": "yellow",
    "color4": "blue",
    "color5": "magenta",
    "color6": "cyan",
    "color7": "white",
    "color8": "bright black",
    "color9": "bright red",
    "color10": "bright green",
    "color11": "bright yellow",
    "color12": "bright blue",
    "color13": "bright magenta",
    "color14": "bright cyan",
    "color15": "bright white",
}

# Primary text colors — held to a higher standard
PRIMARY_TEXT = {"color0", "color7", "color8", "color15"}


def audit_palette(path: str) -> int:
    """Audit a single palette.toml. Returns number of failures."""
    with open(path, "rb") as f:
        data = tomllib.load(f)

    palette = data.get("palette", {})
    theme = data.get("theme", {})
    name = theme.get("name", path)
    variant = theme.get("variant", "unknown")
    bg = palette.get("bg")

    if not bg:
        print(f"  ERROR: no bg color in {path}")
        return 1

    print(f"\n{'=' * 72}")
    print(f"  {name}  ({variant})  bg: {bg}")
    print(f"{'=' * 72}")
    print(f"  {'Color':<8} {'Hex':<10} {'Name':<16} {'Ratio':>7}  {'Status'}")
    print(f"  {'-' * 66}")

    failures = 0
    warnings = 0

    for i in range(16):
        key = f"color{i}"
        hex_val = palette.get(key)
        if not hex_val:
            continue

        ansi_name = ANSI_NAMES.get(key, "")
        ratio = contrast_ratio(hex_val, bg)
        is_primary = key in PRIMARY_TEXT

        if hex_val.lower() == bg.lower():
            status = "= bg"
        elif ratio >= 4.5:
            status = "PASS (AA)"
        elif ratio >= 3.0:
            status = "PASS (3:1+)"
            if is_primary:
                status += "  !! primary text below 4.5:1"
                warnings += 1
        else:
            status = "FAIL"
            failures += 1

        marker = " *" if is_primary else "  "
        print(f"  {key:<8} {hex_val:<10} {ansi_name:<16} {ratio:>5.2f}:1  {status}{marker}")

    # Also check fg contrast
    fg = palette.get("fg")
    if fg:
        fg_ratio = contrast_ratio(fg, bg)
        status = "PASS (AA)" if fg_ratio >= 4.5 else "PASS (3:1+)" if fg_ratio >= 3.0 else "FAIL"
        if fg_ratio < 3.0:
            failures += 1
        print(f"  {'fg':<8} {fg:<10} {'foreground':<16} {fg_ratio:>5.2f}:1  {status}")

    print()
    if failures:
        print(f"  RESULT: {failures} failure(s) — colors below 3:1 on {bg}")
    elif warnings:
        print(f"  RESULT: No failures. {warnings} primary text color(s) below 4.5:1.")
    else:
        print(f"  RESULT: All text colors >= 3:1. Primary text colors >= 4.5:1.")

    print(f"  * = primary text color (color0/7/8/15)")
    return failures


def main():
    if len(sys.argv) < 2:
        print(f"Usage: {sys.argv[0]} <palette.toml> [palette.toml ...]")
        sys.exit(1)

    total_failures = 0
    for path in sys.argv[1:]:
        total_failures += audit_palette(path)

    if total_failures:
        print(f"\n  Total failures across all palettes: {total_failures}")
        sys.exit(1)
    else:
        print(f"\n  All palettes pass contrast audit.")


if __name__ == "__main__":
    main()
