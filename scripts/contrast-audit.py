#!/usr/bin/env python3
# /// script
# requires-python = ">=3.11"
# dependencies = ["coloraide"]
# ///
"""APCA + OKLCH palette audit for the-themer palette TOML files.

Audits terminal color palettes for perceptual contrast (APCA-W3),
hue identity (OKLCH), normal/bright pair coherence, and cross-family
distinguishability.

Usage:
    uv run scripts/contrast-audit.py themes/cobalt-next-neon/palette.toml
    uv run scripts/contrast-audit.py themes/*/palette.toml
    uv run scripts/contrast-audit.py a.toml --compare b.toml
    uv run scripts/contrast-audit.py themes/*/palette.toml --strict
"""

import argparse
import math
import sys
import tomllib
from dataclasses import dataclass, field

from coloraide import Color

# ---------------------------------------------------------------------------
# ANSI slot names and classifications
# ---------------------------------------------------------------------------

ANSI_NAMES = {
    0: "black", 1: "red", 2: "green", 3: "yellow",
    4: "blue", 5: "magenta", 6: "cyan", 7: "white",
    8: "br black", 9: "br red", 10: "br green", 11: "br yellow",
    12: "br blue", 13: "br magenta", 14: "br cyan", 15: "br white",
}

# Normal/bright pairs sharing a hue family
ANSI_PAIRS = [(1, 9), (2, 10), (3, 11), (4, 12), (5, 13), (6, 14)]

# Chromatic slots (exclude achromatic 0/7/8/15)
CHROMATIC_NORMALS = [1, 2, 3, 4, 5, 6]
CHROMATIC_BRIGHTS = [9, 10, 11, 12, 13, 14]

# ---------------------------------------------------------------------------
# APCA-W3 0.0.98G-4g — vendored from Myndex/apca-w3
#
# Constants from: https://github.com/Myndex/apca-w3/blob/master/src/apca-w3.js
# Algorithm: SA98G (S-Luv Advanced Perceptual Contrast Algorithm)
# License: W3C Software and Document License (permits derivative works)
# ---------------------------------------------------------------------------

# sRGB coefficients for luminance
_APCA_SR = 0.2126729
_APCA_SG = 0.7151522
_APCA_SB = 0.0721750

# Exponents: normal polarity (dark text on light bg)
_APCA_NORM_BG = 0.56
_APCA_NORM_TXT = 0.57

# Exponents: reverse polarity (light text on dark bg)
_APCA_REV_BG = 0.65
_APCA_REV_TXT = 0.62

# Soft-clamp for near-black
_APCA_BLK_THRS = 0.022
_APCA_BLK_CLMP = 1.414

# Output scaling
_APCA_SCALE_BOW = 1.14
_APCA_SCALE_WOB = 1.14
_APCA_LO_BOW_OFFSET = 0.027
_APCA_LO_WOB_OFFSET = 0.027
_APCA_LO_CLIP = 0.1

# Minimum luminance delta
_APCA_DELTA_Y_MIN = 0.0005

# TRC (gamma)
_APCA_TRC = 2.4


def _srgb_to_y(r: float, g: float, b: float) -> float:
    """Convert sRGB 0-255 channel values to APCA luminance (Y)."""
    # Linearize via simple gamma (APCA uses 2.4, not the piecewise sRGB TRC)
    r_lin = (r / 255.0) ** _APCA_TRC
    g_lin = (g / 255.0) ** _APCA_TRC
    b_lin = (b / 255.0) ** _APCA_TRC
    return _APCA_SR * r_lin + _APCA_SG * g_lin + _APCA_SB * b_lin


def _hex_to_rgb(h: str) -> tuple[int, int, int]:
    """Parse '#RRGGBB' to (R, G, B) as 0-255 ints."""
    h = h.lstrip("#")
    return int(h[0:2], 16), int(h[2:4], 16), int(h[4:6], 16)


def apca_contrast(txt_hex: str, bg_hex: str) -> float:
    """Compute APCA Lc (lightness contrast) for text on background.

    Returns signed Lc * 100:
      positive = dark text on light bg (BoW)
      negative = light text on dark bg (WoB)
    """
    txt_y = _srgb_to_y(*_hex_to_rgb(txt_hex))
    bg_y = _srgb_to_y(*_hex_to_rgb(bg_hex))

    # Soft-clamp near-black
    if txt_y <= _APCA_BLK_THRS:
        txt_y += (_APCA_BLK_THRS - txt_y) ** _APCA_BLK_CLMP
    if bg_y <= _APCA_BLK_THRS:
        bg_y += (_APCA_BLK_THRS - bg_y) ** _APCA_BLK_CLMP

    # Bail on negligible difference
    if abs(bg_y - txt_y) < _APCA_DELTA_Y_MIN:
        return 0.0

    # Normal polarity: dark text on light bg
    if bg_y > txt_y:
        sapc = (bg_y**_APCA_NORM_BG - txt_y**_APCA_NORM_TXT) * _APCA_SCALE_BOW
        if sapc < _APCA_LO_CLIP:
            return 0.0
        return (sapc - _APCA_LO_BOW_OFFSET) * 100.0

    # Reverse polarity: light text on dark bg
    sapc = (bg_y**_APCA_REV_BG - txt_y**_APCA_REV_TXT) * _APCA_SCALE_WOB
    if sapc > -_APCA_LO_CLIP:
        return 0.0
    return (sapc + _APCA_LO_WOB_OFFSET) * 100.0


# ---------------------------------------------------------------------------
# OKLCH helpers (via coloraide)
# ---------------------------------------------------------------------------

@dataclass
class OklchColor:
    """OKLCH decomposition of a hex color."""
    hex: str
    L: float  # 0-1
    C: float  # 0-~0.4
    H: float  # 0-360 (NaN for achromatic)

    @property
    def is_achromatic(self) -> bool:
        return self.C < 0.02

    @property
    def hue_unreliable(self) -> bool:
        return self.C < 0.04


def hex_to_oklch(hex_color: str) -> OklchColor:
    """Convert a hex color to its OKLCH decomposition."""
    c = Color(hex_color).convert("oklch")
    L, C, H = c.coords()
    # coloraide returns NaN hue for achromatic colors
    if math.isnan(H):
        H = 0.0
    return OklchColor(hex=hex_color, L=L, C=C, H=H)


def delta_e_ok(hex1: str, hex2: str) -> float:
    """Compute deltaE_OK between two hex colors."""
    return Color(hex1).delta_e(Color(hex2), method="ok")


def circular_hue_distance(h1: float, h2: float) -> float:
    """Circular distance between two hue angles in degrees."""
    d = abs(h1 - h2) % 360
    return min(d, 360.0 - d)


# ---------------------------------------------------------------------------
# Hue family definitions — centroid-based with max angular distance
# ---------------------------------------------------------------------------

@dataclass
class HueFamily:
    name: str
    centroid: float  # degrees
    max_distance: float  # degrees
    normal_slot: int
    bright_slot: int


HUE_FAMILIES = [
    HueFamily("red", 30, 35, 1, 9),
    HueFamily("green", 145, 35, 2, 10),
    HueFamily("yellow", 82, 35, 3, 11),
    HueFamily("blue", 260, 30, 4, 12),
    HueFamily("magenta", 330, 45, 5, 13),
    HueFamily("cyan", 205, 35, 6, 14),
]

SLOT_TO_FAMILY = {}
for fam in HUE_FAMILIES:
    SLOT_TO_FAMILY[fam.normal_slot] = fam
    SLOT_TO_FAMILY[fam.bright_slot] = fam


def nearest_family(hue: float) -> tuple[HueFamily, float]:
    """Find the nearest hue family and its angular distance."""
    best_fam = HUE_FAMILIES[0]
    best_dist = circular_hue_distance(hue, best_fam.centroid)
    for fam in HUE_FAMILIES[1:]:
        d = circular_hue_distance(hue, fam.centroid)
        if d < best_dist:
            best_dist = d
            best_fam = fam
    return best_fam, best_dist


# ---------------------------------------------------------------------------
# APCA threshold classification
# ---------------------------------------------------------------------------

class Severity:
    FAIL = "FAIL"
    WARN = "WARN"
    PASS = "PASS"
    INFO = "INFO"
    EXEMPT = "EXEMPT"


@dataclass
class SlotThresholds:
    fail_below: float
    warn_below: float | None  # None means no warn tier


def classify_slot(slot_index: int, variant: str) -> str:
    """Return the threshold category for a given ANSI slot and theme variant."""
    is_dark = variant == "dark"

    # bg-matching slots are exempt
    if is_dark and slot_index == 0:
        return "bg-match"
    if not is_dark and slot_index == 15:
        return "bg-match"

    # Strong anchors: the text-end of the achromatic pair
    if is_dark and slot_index == 15:
        return "body"
    if not is_dark and slot_index == 0:
        return "body"

    # Comment/dim text
    if is_dark and slot_index == 8:
        return "comment"
    if not is_dark and slot_index == 7:
        return "comment"

    # The remaining achromatic that's not comment or body
    if is_dark and slot_index == 7:
        return "comment"  # color7 in dark = secondary text, same tier
    if not is_dark and slot_index == 8:
        return "comment"  # color8 in light = secondary text

    # Chromatic normals
    if slot_index in CHROMATIC_NORMALS:
        return "chromatic-normal"

    # Chromatic brights
    if slot_index in CHROMATIC_BRIGHTS:
        return "chromatic-bright"

    return "unknown"


THRESHOLDS = {
    "body":             SlotThresholds(fail_below=75, warn_below=90),
    "comment":          SlotThresholds(fail_below=30, warn_below=45),
    "chromatic-normal": SlotThresholds(fail_below=30, warn_below=45),
    "chromatic-bright": SlotThresholds(fail_below=45, warn_below=60),
    "ui-element":       SlotThresholds(fail_below=45, warn_below=60),
    "ui-structural":    SlotThresholds(fail_below=15, warn_below=None),
    "bg-match":         SlotThresholds(fail_below=0, warn_below=None),  # exempt
}


def classify_ui_slot(name: str) -> str:
    """Return the threshold category for a UI/semantic slot."""
    if name in ("border", "dimmed"):
        return "ui-structural"
    return "ui-element"


def evaluate_contrast(lc: float, category: str) -> str:
    """Return severity for an Lc value in a given category."""
    if category == "bg-match":
        return Severity.EXEMPT

    t = THRESHOLDS[category]
    abs_lc = abs(lc)

    if abs_lc < t.fail_below:
        return Severity.FAIL
    if t.warn_below is not None and abs_lc < t.warn_below:
        return Severity.WARN
    return Severity.PASS


# ---------------------------------------------------------------------------
# Palette loading — mirrors Go's palette.ApplyDefaults
# ---------------------------------------------------------------------------

@dataclass
class Palette:
    """Loaded and defaults-applied palette."""
    name: str
    variant: str
    path: str

    bg: str = ""
    fg: str = ""
    cursor: str = ""
    selection_bg: str = ""
    selection_fg: str = ""

    colors: dict[int, str] = field(default_factory=dict)

    # UI semantic
    ui_border: str = ""
    ui_dimmed: str = ""
    ui_accent: str = ""
    ui_success: str = ""
    ui_warning: str = ""
    ui_error: str = ""
    ui_info: str = ""

    # Syntax
    syntax_number: str = ""
    syntax_error: str = ""
    syntax_line_highlight: str = ""

    # Adapter overrides (name -> sub-Palette)
    adapter_overrides: dict[str, "Palette"] = field(default_factory=dict)


def load_palette(path: str) -> Palette:
    """Load a palette TOML, apply defaults, return a Palette."""
    with open(path, "rb") as f:
        data = tomllib.load(f)

    theme = data.get("theme", {})
    pal = data.get("palette", {})
    ui = pal.get("ui", {})
    syntax = pal.get("syntax", {})

    p = Palette(
        name=theme.get("name", path),
        variant=theme.get("variant", "dark"),
        path=path,
        bg=pal.get("bg", ""),
        fg=pal.get("fg", ""),
        cursor=pal.get("cursor", ""),
        selection_bg=pal.get("selection_bg", ""),
        selection_fg=pal.get("selection_fg", ""),
        ui_border=ui.get("border", ""),
        ui_dimmed=ui.get("dimmed", ""),
        ui_accent=ui.get("accent", ""),
        ui_success=ui.get("success", ""),
        ui_warning=ui.get("warning", ""),
        ui_error=ui.get("error", ""),
        ui_info=ui.get("info", ""),
        syntax_number=syntax.get("number", ""),
        syntax_error=syntax.get("error", ""),
        syntax_line_highlight=syntax.get("line_highlight", ""),
    )

    for i in range(16):
        val = pal.get(f"color{i}", "")
        if val:
            p.colors[i] = val

    # Apply defaults (mirrors Go's ApplyDefaults)
    if not p.cursor:
        p.cursor = p.colors.get(4, "")
    if not p.selection_bg:
        p.selection_bg = p.colors.get(8, "")
    if not p.selection_fg:
        p.selection_fg = p.fg
    if not p.ui_border:
        p.ui_border = p.colors.get(8, "")
    if not p.ui_dimmed:
        p.ui_dimmed = p.colors.get(8, "")
    if not p.ui_accent:
        p.ui_accent = p.colors.get(6, "")
    if not p.ui_success:
        p.ui_success = p.colors.get(2, "")
    if not p.ui_warning:
        p.ui_warning = p.colors.get(3, "")
    if not p.ui_error:
        p.ui_error = p.colors.get(1, "")
    if not p.ui_info:
        p.ui_info = p.colors.get(4, "")
    if not p.syntax_number:
        p.syntax_number = p.colors.get(4, "")
    if not p.syntax_error:
        p.syntax_error = p.colors.get(1, "")
    if not p.syntax_line_highlight:
        p.syntax_line_highlight = p.selection_bg

    # Load adapter overrides
    adapters = data.get("adapters", {})
    for adapter_name, adapter_data in adapters.items():
        override_pal = adapter_data.get("palette", {})
        if not override_pal:
            continue
        override_ui = override_pal.get("ui", {})
        op = Palette(
            name=f"{p.name} ({adapter_name} override)",
            variant=p.variant,
            path=path,
            bg=override_pal.get("bg", ""),
            fg=override_pal.get("fg", ""),
            cursor=override_pal.get("cursor", ""),
            selection_bg=override_pal.get("selection_bg", ""),
            selection_fg=override_pal.get("selection_fg", ""),
            ui_border=override_ui.get("border", ""),
            ui_dimmed=override_ui.get("dimmed", ""),
            ui_accent=override_ui.get("accent", ""),
            ui_success=override_ui.get("success", ""),
            ui_warning=override_ui.get("warning", ""),
            ui_error=override_ui.get("error", ""),
            ui_info=override_ui.get("info", ""),
        )
        for i in range(16):
            val = override_pal.get(f"color{i}", "")
            if val:
                op.colors[i] = val

        # Apply defaults for override
        if not op.cursor:
            op.cursor = op.colors.get(4, "")
        if not op.selection_bg:
            op.selection_bg = op.colors.get(8, "")
        if not op.selection_fg:
            op.selection_fg = op.fg
        if not op.ui_border:
            op.ui_border = op.colors.get(8, "")
        if not op.ui_dimmed:
            op.ui_dimmed = op.colors.get(8, "")
        if not op.ui_accent:
            op.ui_accent = op.colors.get(6, "")
        if not op.ui_success:
            op.ui_success = op.colors.get(2, "")
        if not op.ui_warning:
            op.ui_warning = op.colors.get(3, "")
        if not op.ui_error:
            op.ui_error = op.colors.get(1, "")
        if not op.ui_info:
            op.ui_info = op.colors.get(4, "")

        p.adapter_overrides[adapter_name] = op

    return p


# ---------------------------------------------------------------------------
# Audit result types
# ---------------------------------------------------------------------------

@dataclass
class ContrastResult:
    slot: str
    hex_color: str
    lc: float
    category: str
    severity: str


@dataclass
class OklchResult:
    slot: str
    hex_color: str
    oklch: OklchColor


@dataclass
class HueResult:
    slot: int
    hex_color: str
    expected_family: str
    actual_family: str
    distance_from_centroid: float
    max_distance: float
    severity: str
    note: str = ""


@dataclass
class PairResult:
    family: str
    normal_slot: int
    bright_slot: int
    delta_L: float
    delta_C: float
    hue_drift: float
    severity: str
    note: str = ""


@dataclass
class DistinguishResult:
    slot_a: int
    slot_b: int
    delta_e: float
    hue_distance: float
    severity: str
    note: str = ""


# ---------------------------------------------------------------------------
# US-1: APCA contrast audit
# ---------------------------------------------------------------------------

def audit_apca_contrast(pal: Palette) -> list[ContrastResult]:
    """Audit every color slot against bg using APCA."""
    results = []

    # fg and selection_fg — body text
    for slot_name, hex_val in [("fg", pal.fg), ("selection_fg", pal.selection_fg)]:
        if not hex_val:
            continue
        lc = apca_contrast(hex_val, pal.bg)
        sev = evaluate_contrast(lc, "body")
        results.append(ContrastResult(slot_name, hex_val, lc, "body", sev))

    # ANSI 16
    for i in range(16):
        hex_val = pal.colors.get(i)
        if not hex_val:
            continue
        cat = classify_slot(i, pal.variant)
        lc = apca_contrast(hex_val, pal.bg)
        sev = evaluate_contrast(lc, cat)
        results.append(ContrastResult(f"color{i}", hex_val, lc, cat, sev))

    # UI slots
    ui_slots = [
        ("ui.accent", pal.ui_accent),
        ("ui.success", pal.ui_success),
        ("ui.warning", pal.ui_warning),
        ("ui.error", pal.ui_error),
        ("ui.info", pal.ui_info),
        ("ui.border", pal.ui_border),
        ("ui.dimmed", pal.ui_dimmed),
    ]
    for slot_name, hex_val in ui_slots:
        if not hex_val:
            continue
        cat = classify_ui_slot(slot_name.split(".")[-1])
        lc = apca_contrast(hex_val, pal.bg)
        sev = evaluate_contrast(lc, cat)
        results.append(ContrastResult(slot_name, hex_val, lc, cat, sev))

    return results


# ---------------------------------------------------------------------------
# US-2: OKLCH decomposition
# ---------------------------------------------------------------------------

def audit_oklch_decomposition(pal: Palette) -> list[OklchResult]:
    """Decompose every palette color into OKLCH."""
    results = []

    for slot_name, hex_val in [("bg", pal.bg), ("fg", pal.fg)]:
        if hex_val:
            results.append(OklchResult(slot_name, hex_val, hex_to_oklch(hex_val)))

    for i in range(16):
        hex_val = pal.colors.get(i)
        if hex_val:
            results.append(OklchResult(
                f"color{i}", hex_val, hex_to_oklch(hex_val)))

    return results


# ---------------------------------------------------------------------------
# US-3: Hue identity audit
# ---------------------------------------------------------------------------

def audit_hue_identity(pal: Palette) -> list[HueResult]:
    """Check that chromatic slots land in their expected hue family."""
    results = []

    for slot_idx in CHROMATIC_NORMALS + CHROMATIC_BRIGHTS:
        hex_val = pal.colors.get(slot_idx)
        if not hex_val:
            continue

        oklch = hex_to_oklch(hex_val)
        expected_fam = SLOT_TO_FAMILY[slot_idx]

        # Achromatic in a chromatic slot
        if oklch.is_achromatic:
            results.append(HueResult(
                slot=slot_idx,
                hex_color=hex_val,
                expected_family=expected_fam.name,
                actual_family="achromatic",
                distance_from_centroid=0,
                max_distance=expected_fam.max_distance,
                severity=Severity.WARN,
                note=f"C={oklch.C:.3f} — achromatic in chromatic slot",
            ))
            continue

        dist = circular_hue_distance(oklch.H, expected_fam.centroid)
        actual_fam, _ = nearest_family(oklch.H)

        if dist <= expected_fam.max_distance:
            sev = Severity.PASS
            note = ""
        elif oklch.hue_unreliable:
            sev = Severity.INFO
            note = f"C={oklch.C:.3f} — low chroma, hue unreliable"
        else:
            sev = Severity.WARN
            note = f"H={oklch.H:.0f}° — reads as {actual_fam.name}"

        results.append(HueResult(
            slot=slot_idx,
            hex_color=hex_val,
            expected_family=expected_fam.name,
            actual_family=actual_fam.name,
            distance_from_centroid=dist,
            max_distance=expected_fam.max_distance,
            severity=sev,
            note=note,
        ))

    return results


# ---------------------------------------------------------------------------
# US-4: Normal/bright pair coherence
# ---------------------------------------------------------------------------

def audit_pair_coherence(pal: Palette) -> list[PairResult]:
    """Check that normal/bright pairs have coherent L, C, and hue."""
    results = []

    for normal_idx, bright_idx in ANSI_PAIRS:
        normal_hex = pal.colors.get(normal_idx)
        bright_hex = pal.colors.get(bright_idx)
        if not normal_hex or not bright_hex:
            continue

        n = hex_to_oklch(normal_hex)
        b = hex_to_oklch(bright_hex)
        fam = SLOT_TO_FAMILY[normal_idx]

        delta_L = b.L - n.L
        delta_C = b.C - n.C
        hue_drift = circular_hue_distance(n.H, b.H)

        notes = []
        severity = Severity.PASS

        # In dark themes, brights should be lighter. In light themes, brights
        # should also typically be lighter (more saturated/visible). But if the
        # normal is already very light in a light theme, brights can be similar.
        if pal.variant == "dark" and delta_L < -0.05:
            notes.append("bright is darker than normal")
            severity = Severity.WARN
        elif pal.variant == "light" and delta_L > 0.05:
            # In light themes, brights are often lighter (less contrast) which
            # can be intentional. Only flag if they're significantly lighter.
            if delta_L > 0.15:
                notes.append("bright significantly lighter than normal")
                severity = Severity.WARN

        # Hue drift
        if not (n.is_achromatic or b.is_achromatic):
            if hue_drift > 30:
                notes.append(f"hue drift {hue_drift:.0f}° between normal/bright")
                severity = Severity.WARN

        results.append(PairResult(
            family=fam.name,
            normal_slot=normal_idx,
            bright_slot=bright_idx,
            delta_L=delta_L,
            delta_C=delta_C,
            hue_drift=hue_drift,
            severity=severity,
            note="; ".join(notes),
        ))

    return results


# ---------------------------------------------------------------------------
# US-5: Cross-family distinguishability
# ---------------------------------------------------------------------------

def audit_distinguishability(pal: Palette) -> list[DistinguishResult]:
    """Check pairwise deltaE_OK among chromatic normals and brights."""
    results = []

    for slots in [CHROMATIC_NORMALS, CHROMATIC_BRIGHTS]:
        hex_colors = {}
        for s in slots:
            if s in pal.colors:
                hex_colors[s] = pal.colors[s]

        slot_list = sorted(hex_colors.keys())
        for i, sa in enumerate(slot_list):
            for sb in slot_list[i + 1:]:
                de = delta_e_ok(hex_colors[sa], hex_colors[sb])
                hd = circular_hue_distance(
                    hex_to_oklch(hex_colors[sa]).H,
                    hex_to_oklch(hex_colors[sb]).H,
                )

                if de < 0.04:
                    sev = Severity.FAIL
                    note = "likely confusable"
                elif de < 0.07:
                    if hd > 30:
                        sev = Severity.INFO
                        note = "metrically close but hue-distinct"
                    else:
                        sev = Severity.WARN
                        note = "may be confused in some contexts"
                else:
                    sev = Severity.PASS
                    note = ""

                # Only report non-passing pairs
                if sev != Severity.PASS:
                    results.append(DistinguishResult(
                        slot_a=sa, slot_b=sb, delta_e=de,
                        hue_distance=hd, severity=sev, note=note))

    return results


# ---------------------------------------------------------------------------
# US-6: Cross-context contrast
# ---------------------------------------------------------------------------

def audit_cross_context(pal: Palette) -> list[ContrastResult]:
    """Audit contrast in non-standard contexts (selection, cursor, etc.)."""
    results = []

    pairs = [
        ("selection_fg on selection_bg", pal.selection_fg, pal.selection_bg, "body"),
        ("fg on cursor (block)", pal.fg, pal.cursor, "body"),
        ("cursor on bg", pal.cursor, pal.bg, "ui-element"),
    ]

    for label, fg_hex, bg_hex, category in pairs:
        if not fg_hex or not bg_hex:
            continue
        lc = apca_contrast(fg_hex, bg_hex)
        sev = evaluate_contrast(lc, category)
        results.append(ContrastResult(label, fg_hex, lc, category, sev))

    # UI and syntax colors on bg
    ui_slots = [
        ("ui.accent on bg", pal.ui_accent),
        ("ui.success on bg", pal.ui_success),
        ("ui.warning on bg", pal.ui_warning),
        ("ui.error on bg", pal.ui_error),
        ("ui.info on bg", pal.ui_info),
    ]
    for label, hex_val in ui_slots:
        if not hex_val:
            continue
        lc = apca_contrast(hex_val, pal.bg)
        sev = evaluate_contrast(lc, "ui-element")
        results.append(ContrastResult(label, hex_val, lc, "ui-element", sev))

    # Syntax colors on bg
    syntax_slots = [
        ("syntax.number on bg", pal.syntax_number),
        ("syntax.error on bg", pal.syntax_error),
    ]
    for label, hex_val in syntax_slots:
        if not hex_val:
            continue
        lc = apca_contrast(hex_val, pal.bg)
        sev = evaluate_contrast(lc, "chromatic-normal")
        results.append(ContrastResult(label, hex_val, lc, "chromatic-normal", sev))

    return results


# ---------------------------------------------------------------------------
# US-8: Adapter override audit
# ---------------------------------------------------------------------------

def audit_adapter_overrides(pal: Palette) -> dict[str, dict]:
    """Run the full audit pipeline on each adapter override palette."""
    results = {}
    for adapter_name, override_pal in pal.adapter_overrides.items():
        results[adapter_name] = {
            "contrast": audit_apca_contrast(override_pal),
            "oklch": audit_oklch_decomposition(override_pal),
            "hue": audit_hue_identity(override_pal),
            "pairs": audit_pair_coherence(override_pal),
            "distinguish": audit_distinguishability(override_pal),
        }
    return results


# ---------------------------------------------------------------------------
# US-7: Comparison
# ---------------------------------------------------------------------------

@dataclass
class ComparisonDelta:
    slot: str
    lc_a: float
    lc_b: float
    lc_change: float
    sev_a: str
    sev_b: str
    regression: bool


def compare_palettes(pal_a: Palette, pal_b: Palette) -> list[ComparisonDelta]:
    """Compare two palettes and report Lc changes and regressions."""
    results_a = {r.slot: r for r in audit_apca_contrast(pal_a)}
    results_b = {r.slot: r for r in audit_apca_contrast(pal_b)}

    deltas = []
    all_slots = sorted(set(results_a.keys()) | set(results_b.keys()))
    for slot in all_slots:
        ra = results_a.get(slot)
        rb = results_b.get(slot)
        if not ra or not rb:
            continue

        lc_change = rb.lc - ra.lc
        # Regression = was passing, now failing (or was warn, now fail)
        severity_rank = {Severity.PASS: 0, Severity.WARN: 1,
                         Severity.FAIL: 2, Severity.EXEMPT: -1,
                         Severity.INFO: -1}
        reg = severity_rank.get(rb.severity, 0) > severity_rank.get(ra.severity, 0)

        deltas.append(ComparisonDelta(
            slot=slot, lc_a=ra.lc, lc_b=rb.lc, lc_change=lc_change,
            sev_a=ra.severity, sev_b=rb.severity, regression=reg))

    return deltas


# ---------------------------------------------------------------------------
# Report formatting
# ---------------------------------------------------------------------------

SEV_MARKERS = {
    Severity.FAIL: "\033[1;31mFAIL\033[0m",
    Severity.WARN: "\033[1;33mWARN\033[0m",
    Severity.PASS: "\033[32mPASS\033[0m",
    Severity.INFO: "\033[36mINFO\033[0m",
    Severity.EXEMPT: "\033[90mEXEMPT\033[0m",
}


def sev_marker(sev: str) -> str:
    return SEV_MARKERS.get(sev, sev)


def print_header(title: str, pal: Palette) -> None:
    print(f"\n{'=' * 76}")
    print(f"  {title}")
    print(f"  {pal.name}  ({pal.variant})  bg: {pal.bg}")
    print(f"{'=' * 76}")


def print_section(title: str) -> None:
    print(f"\n  --- {title} ---\n")


def report_contrast(results: list[ContrastResult]) -> int:
    """Print contrast results. Returns failure count."""
    print_section("APCA Contrast (Lc)")
    print(f"  {'Slot':<26} {'Hex':<10} {'Lc':>7}  {'Category':<18} Status")
    print(f"  {'-' * 70}")

    failures = 0
    for r in results:
        if r.severity == Severity.FAIL:
            failures += 1
        lc_str = f"{r.lc:>+7.1f}"
        print(f"  {r.slot:<26} {r.hex_color:<10} {lc_str}  {r.category:<18} {sev_marker(r.severity)}")

    return failures


def report_oklch(results: list[OklchResult]) -> None:
    print_section("OKLCH Decomposition")
    print(f"  {'Slot':<10} {'Hex':<10} {'L':>6} {'C':>6} {'H':>6}  Note")
    print(f"  {'-' * 52}")

    for r in results:
        o = r.oklch
        note = ""
        if o.is_achromatic:
            note = "achromatic"
        elif o.hue_unreliable:
            note = "low chroma"
        print(f"  {r.slot:<10} {r.hex_color:<10} {o.L:>5.3f} {o.C:>5.3f} {o.H:>5.0f}°  {note}")


def report_hue_identity(results: list[HueResult]) -> None:
    print_section("Hue Identity")
    print(f"  {'Slot':<6} {'Expected':<10} {'Actual':<10} {'Dist':>5}°/{'Max':>3}°  Status  Note")
    print(f"  {'-' * 62}")

    for r in results:
        slot_name = f"{r.slot:<3} {ANSI_NAMES[r.slot]}"
        print(f"  {slot_name:<16} {r.expected_family:<10} {r.actual_family:<10} "
              f"{r.distance_from_centroid:>4.0f}°/{r.max_distance:>2.0f}°  "
              f"{sev_marker(r.severity)}  {r.note}")


def report_pair_coherence(results: list[PairResult]) -> None:
    print_section("Pair Coherence (normal/bright)")
    print(f"  {'Family':<10} {'Pair':<8} {'dL':>6} {'dC':>6} {'dH':>5}°  Status  Note")
    print(f"  {'-' * 60}")

    for r in results:
        pair = f"{r.normal_slot}/{r.bright_slot}"
        print(f"  {r.family:<10} {pair:<8} {r.delta_L:>+5.3f} {r.delta_C:>+5.3f} "
              f"{r.hue_drift:>4.0f}°  {sev_marker(r.severity)}  {r.note}")


def report_distinguishability(results: list[DistinguishResult]) -> None:
    print_section("Distinguishability (closest pairs)")
    if not results:
        print("  All chromatic pairs sufficiently distinct.")
        return
    print(f"  {'Pair':<18} {'dE_OK':>6} {'dH':>5}°  Status  Note")
    print(f"  {'-' * 55}")

    for r in results:
        a_name = ANSI_NAMES[r.slot_a]
        b_name = ANSI_NAMES[r.slot_b]
        pair = f"{a_name}/{b_name}"
        print(f"  {pair:<18} {r.delta_e:>5.3f} {r.hue_distance:>4.0f}°  "
              f"{sev_marker(r.severity)}  {r.note}")


def report_cross_context(results: list[ContrastResult]) -> int:
    """Print cross-context results. Returns failure count."""
    print_section("Cross-Context Contrast")
    print(f"  {'Context':<30} {'Hex':<10} {'Lc':>7}  {'Category':<18} Status")
    print(f"  {'-' * 74}")

    failures = 0
    for r in results:
        if r.severity == Severity.FAIL:
            failures += 1
        print(f"  {r.slot:<30} {r.hex_color:<10} {r.lc:>+7.1f}  "
              f"{r.category:<18} {sev_marker(r.severity)}")

    return failures


def report_comparison(deltas: list[ComparisonDelta],
                      name_a: str, name_b: str) -> None:
    print_section(f"Comparison: {name_a} -> {name_b}")
    print(f"  {'Slot':<26} {'Lc A':>7} {'Lc B':>7} {'Change':>7}  A     B     Reg")
    print(f"  {'-' * 72}")

    regressions = 0
    for d in deltas:
        if d.sev_a == Severity.EXEMPT and d.sev_b == Severity.EXEMPT:
            continue  # Skip bg-matching slots
        reg_marker = " <<<" if d.regression else ""
        if d.regression:
            regressions += 1
        print(f"  {d.slot:<26} {d.lc_a:>+7.1f} {d.lc_b:>+7.1f} {d.lc_change:>+7.1f}  "
              f"{d.sev_a:<5} {d.sev_b:<5} {reg_marker}")

    if regressions:
        print(f"\n  {regressions} regression(s) detected.")
    else:
        print(f"\n  No regressions.")


# ---------------------------------------------------------------------------
# Full audit pipeline
# ---------------------------------------------------------------------------

def audit_palette(pal: Palette, strict: bool = False) -> int:
    """Run all audit sections on a palette. Returns failure count."""
    print_header("Palette Audit", pal)

    # US-1: APCA Contrast
    contrast_results = audit_apca_contrast(pal)
    failures = report_contrast(contrast_results)

    # US-2: OKLCH Decomposition
    oklch_results = audit_oklch_decomposition(pal)
    report_oklch(oklch_results)

    # US-3: Hue Identity
    hue_results = audit_hue_identity(pal)
    report_hue_identity(hue_results)

    # US-4: Pair Coherence
    pair_results = audit_pair_coherence(pal)
    report_pair_coherence(pair_results)

    # US-5: Distinguishability
    distinguish_results = audit_distinguishability(pal)
    report_distinguishability(distinguish_results)

    # US-6: Cross-Context Contrast
    cross_results = audit_cross_context(pal)
    failures += report_cross_context(cross_results)

    # US-8: Adapter Overrides
    override_results = audit_adapter_overrides(pal)
    for adapter_name, sections in override_results.items():
        print(f"\n  {'=' * 40}")
        print(f"  Adapter Override: {adapter_name}")
        print(f"  {'=' * 40}")
        failures += report_contrast(sections["contrast"])
        report_oklch(sections["oklch"])
        report_hue_identity(sections["hue"])
        report_pair_coherence(sections["pairs"])
        report_distinguishability(sections["distinguish"])

    # Warn count
    warns = sum(1 for r in contrast_results if r.severity == Severity.WARN)
    warns += sum(1 for r in cross_results if r.severity == Severity.WARN)

    if strict:
        failures += warns

    # Summary
    print()
    if failures:
        print(f"  RESULT: {failures} failure(s)")
    elif warns:
        print(f"  RESULT: No failures. {warns} warning(s).")
    else:
        print(f"  RESULT: All checks pass.")

    return failures


# ---------------------------------------------------------------------------
# CLI
# ---------------------------------------------------------------------------

def main() -> None:
    parser = argparse.ArgumentParser(
        description="APCA + OKLCH palette audit for the-themer")
    parser.add_argument("palettes", nargs="+", metavar="TOML",
                        help="Palette TOML file(s) to audit")
    parser.add_argument("--compare", metavar="TOML",
                        help="Compare palette(s) against this baseline")
    parser.add_argument("--strict", action="store_true",
                        help="Treat warnings as failures")
    args = parser.parse_args()

    total_failures = 0

    for path in args.palettes:
        pal = load_palette(path)
        total_failures += audit_palette(pal, strict=args.strict)

        if args.compare:
            compare_pal = load_palette(args.compare)
            deltas = compare_palettes(compare_pal, pal)
            report_comparison(deltas, compare_pal.name, pal.name)

    print()
    if total_failures:
        print(f"  Total failures: {total_failures}")
        sys.exit(1)
    else:
        print(f"  All palettes pass audit.")


if __name__ == "__main__":
    main()
