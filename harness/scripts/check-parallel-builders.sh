#!/usr/bin/env bash
# Meta-Harness — Check Parallel Builders (Sensor 10)
# Detecta overlap de path-scope entre sub-issues em paralelo
# (prontas para entrar em in-progress) e bloqueia se overlap
# existir sem depends-on explícito.
#
# Uso:
#   ./harness/scripts/check-parallel-builders.sh
#   ./harness/scripts/check-parallel-builders.sh --ready    # só ready (default: ready)
#   ./harness/scripts/check-parallel-builders.sh --in-progress
#   ./harness/scripts/check-parallel-builders.sh --json     # output em JSON
#
# Exit codes:
#   0 = sem overlap, OK paralelizar
#   1 = overlap detectado, bloquear
#   2 = sub-issue sem path-scope, bloquear (DoD incompleto)
#   3 = erro de tooling (gh não disponível, etc)
#
# Ver: harness/sensors/10-decomposition-safety.md

set -e

SCOPE="ready"
OUTPUT_JSON=0
while [[ $# -gt 0 ]]; do
  case "$1" in
    --ready)         SCOPE="ready"; shift ;;
    --in-progress)   SCOPE="in-progress"; shift ;;
    --all)           SCOPE="all"; shift ;;
    --json)          OUTPUT_JSON=1; shift ;;
    -h|--help)
      sed -n '2,20p' "$0"
      exit 0 ;;
    *)
      echo "Unknown flag: $1" >&2
      exit 3 ;;
  esac
done

# --- Pre-checks ---
command -v gh >/dev/null 2>&1 || {
  echo "❌ gh CLI not found. Install: https://cli.github.com/" >&2
  exit 3
}
command -v jq >/dev/null 2>&1 || {
  echo "❌ jq not found. Install: brew install jq" >&2
  exit 3
}
command -v python3 >/dev/null 2>&1 || {
  echo "❌ python3 not found." >&2
  exit 3
}

REPO=$(gh repo view --json nameWithOwner -q .nameWithOwner 2>/dev/null || echo "")
if [[ -z "$REPO" ]]; then
  echo "❌ Not in a GitHub repo (or gh not authenticated)." >&2
  exit 3
fi

# --- Fetch issues by scope ---
case "$SCOPE" in
  ready)
    LABEL_FILTER='label:ready -label:in-progress'
    ;;
  in-progress)
    LABEL_FILTER='label:in-progress'
    ;;
  all)
    LABEL_FILTER='-label:done -label:canceled -label:closed'
    ;;
esac

# Pull issues with their labels
ISSUES_JSON=$(gh issue list \
  --repo "$REPO" \
  --state open \
  --limit 200 \
  --json number,title,labels,state \
  --search "$LABEL_FILTER in:title" 2>/dev/null || echo "[]")

# --- Parse path-scope from labels + depends-on from labels ---
# Output structure (Python):
#   [{number, title, path_scopes: [glob...], depends_on: [int...]}]
PARSED=$(echo "$ISSUES_JSON" | python3 -c '
import json, sys, re
data = json.load(sys.stdin)
result = []
for issue in data:
    path_scopes = []
    depends_on = []
    for label in issue.get("labels", []):
        name = label.get("name", "")
        if name.startswith("path-scope:"):
            glob = name.split("path-scope:", 1)[1].strip()
            if glob:
                path_scopes.append(glob)
        elif name.startswith("depends-on:"):
            ref = name.split("depends-on:", 1)[1].strip()
            m = re.match(r"#?(\d+)", ref)
            if m:
                depends_on.append(int(m.group(1)))
    result.append({
        "number": issue["number"],
        "title": issue["title"],
        "path_scopes": path_scopes,
        "depends_on": depends_on,
    })
print(json.dumps(result))
')

# --- Issue count + missing path-scope warning ---
COUNT=$(echo "$PARSED" | python3 -c 'import json,sys; print(len(json.load(sys.stdin)))')
MISSING=$(echo "$PARSED" | python3 -c '
import json, sys
data = json.load(sys.stdin)
print(" ".join(f"#{i[\"number\"]}" for i in data if not i["path_scopes"]))
')

if [[ "$OUTPUT_JSON" -eq 1 ]]; then
  echo "$PARSED"
  exit 0
fi

echo
echo "==> Parallel-builders safety check (sensor 10)"
echo "==> Repo:    $REPO"
echo "==> Scope:   $SCOPE"
echo "==> Issues:  $COUNT"
echo

if [[ -z "$PARSED" || "$PARSED" == "[]" ]]; then
  echo "✅ No $SCOPE issues found. Nothing to check."
  exit 0
fi

# Print per-issue summary
echo "$PARSED" | python3 -c '
import json, sys
data = json.load(sys.stdin)
for i in data:
    n = i["number"]; t = i["title"][:60]
    ps = i["path_scopes"] or ["⚠️  (no path-scope)"]
    dep = i["depends_on"]
    dep_str = f"  depends-on: {dep}" if dep else ""
    print(f"  #{n:<3} {t}{dep_str}")
    for g in ps:
        print(f"        path-scope: {g}")
print()
'

# --- Missing path-scope is a hard fail ---
if [[ -n "$MISSING" ]]; then
  echo "❌ Issues WITHOUT path-scope: $MISSING"
  echo "   → All $SCOPE issues must declare path-scope in DoD."
  echo "   → See: harness/personas/solutions-architect.md §Path scoping"
  exit 2
fi

# --- Detect overlap among all pairs ---
RESULT=$(echo "$PARSED" | python3 -c '
import json, sys, re
from itertools import combinations
import fnmatch
import os

def glob_to_regex(glob):
    """Convert .gitignore-style glob to regex."""
    # fnmatch.translate handles * and ? but not ** (recursive).
    # We translate ** manually.
    pattern = re.escape(glob)
    # Unescape the ** back to .* for recursive
    pattern = pattern.replace(r"\*\*/", ".*").replace(r"\*", "[^/]*")
    pattern = pattern.replace(r"\?", ".")
    return "^" + pattern + "$"

def files_match(glob_a, glob_b):
    """Return True if glob_a and glob_b could overlap on any real file.
    Heuristic: convert to regex, then synthesize a few test paths
    that match glob_a and check if any match glob_b."""
    # Generate test paths matching glob_a
    test_paths = []
    if glob_a.endswith("/**"):
        base = glob_a[:-3]
        for ext in ["", "/file.go", "/sub/file.go", "/deep/nested/file_test.go"]:
            test_paths.append(base + ext)
    elif "**" in glob_a:
        # Generic
        base = glob_a.split("**")[0]
        for ext in ["file.go", "sub/file.go"]:
            test_paths.append(base + ext)
    else:
        test_paths.append(glob_a)

    regex_b = re.compile(glob_to_regex(glob_b))
    return any(regex_b.match(p) for p in test_paths)

data = json.load(sys.stdin)
overlaps = []
for a, b in combinations(data, 2):
    # If A depends on B (or vice versa), skip (serialized is OK)
    if b["number"] in a["depends_on"] or a["number"] in b["depends_on"]:
        continue
    for ga in a["path_scopes"]:
        for gb in b["path_scopes"]:
            if files_match(ga, gb) or files_match(gb, ga):
                overlaps.append({
                    "a": a["number"], "a_glob": ga,
                    "b": b["number"], "b_glob": gb,
                })

print(json.dumps({"overlaps": overlaps, "count": len(overlaps)}))
')

OVERLAP_COUNT=$(echo "$RESULT" | python3 -c 'import json,sys; print(json.load(sys.stdin)["count"])')

if [[ "$OVERLAP_COUNT" -gt 0 ]]; then
  echo "⚠️  OVERLAP DETECTED:"
  echo "$RESULT" | python3 -c '
import json, sys
data = json.load(sys.stdin)
for o in data["overlaps"]:
    print(f"   #{o[\"a\"]} (path-scope: {o[\"a_glob\"]})")
    print(f"   #{o[\"b\"]} (path-scope: {o[\"b_glob\"]})")
    print(f"   → overlapping file(s) possible")
    print()
'
  echo "❌ Action required: add depends-on to one of them, or refactor path-scope."
  echo "   See: harness/sensors/10-decomposition-safety.md"
  exit 1
fi

echo "✅ No path-scope overlap detected. Safe to parallelize."
exit 0
