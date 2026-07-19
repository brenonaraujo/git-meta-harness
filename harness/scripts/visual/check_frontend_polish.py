#!/usr/bin/env python3
# ============================================================================
# git-meta-harness — Check Frontend Polish (Sensor 12, v1.12.0)
# ============================================================================
# Python companion to check-frontend-polish.sh. Reads files list from
# stdin (one per line) + env vars (SUGGEST, WHITELIST_HEX) and reports
# anti-patterns.
# ============================================================================

import sys, re, os

files_text = sys.stdin.read().strip()
files = [f for f in files_text.split("\n") if f]

whitelist_hex = set()
whitelist_raw = os.environ.get("WHITELIST_HEX", "").strip()
if whitelist_raw:
    for h in whitelist_raw.split():
        h = h.strip()
        if h.startswith("#"):
            whitelist_hex.add(h.lower())

suggest_mode = os.environ.get("SUGGEST", "0") == "1"

# Emoji regex (Unicode ranges covering most common emoji)
EMOJI_RE = re.compile(
    r"[\U0001F300-\U0001F9FF"
    r"\U0001FA00-\U0001FA6F"
    r"\U0001FA70-\U0001FAFF"
    r"\U00002600-\U000027BF"
    r"\U0001F600-\U0001F64F"
    r"\U0001F680-\U0001F6FF"
    r"\U0001F1E0-\U0001F1FF"
    r"]"
)

# Color regexes
HEX_RE = re.compile(r"#([0-9a-fA-F]{3,8})\b")
RGB_RE = re.compile(r"\brgba?\s*\([^)]+\)")
HSL_RE = re.compile(r"\bhsla?\s*\([^)]+\)")

# BEM naming
BEM_RE = re.compile(r"\.[a-z][a-z0-9-]*__[a-z]")
BEM_MOD_RE = re.compile(r"\.[a-z][a-z0-9-]*--[a-z]")

# Spacing scale: allow 1,2,4,6,8,12,16,20,24,32,40,48,64,80,96
# Note: 0 is special (= "no margin/padding"), always allowed.
SPACING_ALLOWED = {1, 2, 4, 6, 8, 12, 16, 20, 24, 32, 40, 48, 64, 80, 96}
SPACING_RE = re.compile(r"\b([pm][xytblr]?-|gap-|space-)([1-9][0-9]*)\b")

# Off-stack imports
OFFSTACK_RE = re.compile(
    r"""(?:from|import\s+.*?from)\s+['"](bootstrap|@mui/material|vuetify|element-plus|quasar|@chakra-ui)['"]"""
)

# Inline style with color
INLINE_COLOR_RE = re.compile(r'style\s*=\s*"[^"]*color\s*:\s*(?:#|rgb|hsl|oklch)', re.IGNORECASE)

# img without alt
IMG_NO_ALT_RE = re.compile(r"<img\b(?![^>]*\balt\s*=)[^>]*>", re.IGNORECASE)

# button without accessible text
BUTTON_NO_TEXT_RE = re.compile(
    r"<button\b(?![^>]*\b(aria-label|aria-labelledby)\s*=)[^>]*>\s*</button>",
    re.IGNORECASE,
)


def add_issue(issues, category, file, line, snippet, fix_hint=""):
    issues.append({
        "category": category,
        "file": file,
        "line": line,
        "snippet": snippet,
        "fix": fix_hint,
    })


def is_emoji_whitelisted(file):
    base = os.path.basename(file).lower()
    return any(w.lower() in base for w in ["404", "notfound", "error", "empty"])


issues = []

for fpath in files:
    if not fpath or not os.path.isfile(fpath):
        continue
    try:
        with open(fpath, "r", encoding="utf-8") as fh:
            content = fh.read()
    except (UnicodeDecodeError, OSError):
        continue

    lines = content.split("\n")
    component_name = None
    if fpath.endswith(".vue"):
        component_name = os.path.basename(fpath).replace(".vue", "")

    # 1. Hex colors hardcoded
    for i, line in enumerate(lines, 1):
        stripped = line.strip()
        if stripped.startswith("//") or stripped.startswith("/*") or stripped.startswith("*"):
            continue
        for m in HEX_RE.finditer(line):
            hex_val = "#" + m.group(1)
            if hex_val.lower() in whitelist_hex:
                continue
            if "var(--ui-" in line:
                continue
            add_issue(issues, "hardcoded_colors", fpath, i, hex_val,
                      'use color="primary" or var(--ui-bg-elevated)')
        for m in RGB_RE.finditer(line):
            if "var(--ui-" in line:
                continue
            add_issue(issues, "hardcoded_rgb", fpath, i, m.group(0)[:30], "use var(--ui-...)")
        for m in HSL_RE.finditer(line):
            if "var(--ui-" in line:
                continue
            add_issue(issues, "hardcoded_hsl", fpath, i, m.group(0)[:30], "use var(--ui-...)")

    # 2. BEM naming
    for i, line in enumerate(lines, 1):
        for m in BEM_RE.finditer(line):
            add_issue(issues, "bem_naming", fpath, i, m.group(0),
                      "use Nuxt UI props or Tailwind utilities")
        for m in BEM_MOD_RE.finditer(line):
            add_issue(issues, "bem_naming_modifier", fpath, i, m.group(0),
                      "use Nuxt UI props or Tailwind variants")

    # 3. Spacing off-scale
    for i, line in enumerate(lines, 1):
        for m in SPACING_RE.finditer(line):
            prefix, num = m.group(1), int(m.group(2))
            if num not in SPACING_ALLOWED and num < 100:
                add_issue(issues, "spacing_off_scale", fpath, i,
                          f"{prefix}{num}",
                          "use scale value (1,2,4,6,8,12,16,24,32,48,64,96)")

    # 4. Emojis excessive
    if not is_emoji_whitelisted(fpath):
        emoji_count = len(EMOJI_RE.findall(content))
        is_serious = "form" in fpath.lower() or "dashboard" in fpath.lower() \
                     or "list" in fpath.lower() or "table" in fpath.lower()
        threshold = 1 if is_serious else 3
        if emoji_count > threshold:
            add_issue(issues, "emojis_excessive", fpath, 0,
                      f"{emoji_count} emojis (threshold: {threshold})",
                      'use icons (icon="i-lucide-...") instead of emojis')

    # 5. Redundant comments
    if component_name:
        for i, line in enumerate(lines, 1):
            stripped = line.strip()
            if not stripped.startswith("//"):
                continue
            if any(x in stripped for x in ["TODO", "FIXME", "HACK", "ADR-", "why:", "because:", "rationale:", "refs #", "Refs #"]):
                continue
            comment_text = stripped[2:].strip()
            if comment_text.startswith(component_name) and len(comment_text) > len(component_name) + 5:
                add_issue(issues, "redundant_comment", fpath, i,
                          comment_text[:60],
                          "explain WHY, not WHAT (see code-style.md)")

    # 6. Off-stack imports
    for i, line in enumerate(lines, 1):
        for m in OFFSTACK_RE.finditer(line):
            lib = m.group(1)
            add_issue(issues, "off_stack_imports", fpath, i, lib,
                      f"project uses @nuxt/ui; use Nuxt UI components instead of {lib}")

    # 7. Inline color styles
    for i, line in enumerate(lines, 1):
        if INLINE_COLOR_RE.search(line):
            add_issue(issues, "inline_color_style", fpath, i,
                      line.strip()[:60],
                      'use class="text-primary" instead of inline color')

    # 8. img without alt
    for i, line in enumerate(lines, 1):
        for m in IMG_NO_ALT_RE.finditer(line):
            add_issue(issues, "img_no_alt", fpath, i,
                      m.group(0)[:60],
                      'add alt="" (or descriptive text) to <img>')

    # 9. button without text/aria-label
    for i, line in enumerate(lines, 1):
        for m in BUTTON_NO_TEXT_RE.finditer(line):
            add_issue(issues, "button_no_text", fpath, i,
                      m.group(0)[:60],
                      "add text content or aria-label to <button>")


# Report
if not issues:
    print("OK: No frontend polish issues detected.")
    sys.exit(0)

# Group by category
by_cat = {}
for iss in issues:
    by_cat.setdefault(iss["category"], []).append(iss)

print(f"BLOCKING: POLISH ISSUES DETECTED (sensor 12, v1.12.0):")
print()
for cat, items in by_cat.items():
    print(f"  {cat} ({len(items)} occurrences):")
    for iss in items[:5]:
        loc = f"{iss['file']}:{iss['line']}" if iss["line"] else iss["file"]
        print(f"    {loc} -> {iss['snippet']}")
    if len(items) > 5:
        print(f"    ... and {len(items) - 5} more")
    print()

print(f"Total: {len(issues)} issues across {len(by_cat)} categories.")
print()
print("Recovery (apos corrigir, rode o sensor novamente):")
print('  - Cores hex: use color="primary" ou var(--ui-bg-elevated)')
print("  - BEM: use Nuxt UI props ou Tailwind utilities, NAO misture")
print('  - Emojis: remova ou troque por icones (icon="i-lucide-...")')
print("  - Comentarios: explique POR QUE, nao O QUE (ver code-style.md)")
print("  - Spacing: use scale 1,2,4,6,8,12,16,24,32,48,64,96 (nunca 3,5,7,9,10,11)")
print()
print("Detalhes: harness/sensors/12-frontend-polish.md")
print("Skills:  harness/skills/{frontend-public-skills,nuxt-ui-patterns,visual-polish}/")
print()

sys.exit(1)
