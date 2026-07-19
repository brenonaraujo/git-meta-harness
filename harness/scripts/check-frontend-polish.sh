#!/usr/bin/env bash
# ============================================================================
# git-meta-harness — Check Frontend Polish (Sensor 12, v1.12.0)
# ============================================================================
# Detecta anti-patterns visuais em arquivos .vue/.css/.scss de projetos
# frontend (Nuxt UI, Tailwind-only, Vue, React). BLOQUEANTE (exit 1).
#
# Uso:
#   ./harness/scripts/check-frontend-polish.sh                       # scan all
#   ./harness/scripts/check-frontend-polish.sh path/to/file.vue      # 1 file
#   ./harness/scripts/check-frontend-polish.sh web/app/              # 1 dir
#
# Saída:
#   - exit 0 = sem violação (pode abrir PR)
#   - exit 1 = violação detectada (bloquear, mostrar recovery)
#   - exit 2 = erro de uso
#
# Diferente do check-scope-discipline.sh (sensor 11, warning-only),
# este SENSOR é BLOQUEANTE porque anti-patterns visuais são
# consertáveis em < 5min mas custam caro se passam (Mandaí v2 PR #23).
#
# Companion Python: harness/scripts/visual/check_frontend_polish.py
# ============================================================================

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PY_COMPANION="${SCRIPT_DIR}/visual/check_frontend_polish.py"

if [[ ! -f "$PY_COMPANION" ]]; then
  echo "❌ Python companion not found: $PY_COMPANION" >&2
  exit 2
fi

SUGGEST_FIX=0
TARGET=""

for arg in "$@"; do
  case "$arg" in
    --suggest-fix) SUGGEST_FIX=1 ;;
    -h|--help)
      echo "Usage: $0 [--suggest-fix] [path]"
      echo ""
      echo "  --suggest-fix  show suggested replacements (don't apply)"
      echo "  path           file.vue, dir, or empty (= scan web/app + components)"
      exit 0 ;;
    *) TARGET="$arg" ;;
  esac
done

# Default target
if [[ -z "$TARGET" ]]; then
  if [[ -d "web/app" ]]; then
    TARGET="web/app"
  elif [[ -d "app" ]]; then
    TARGET="app"
  elif [[ -d "src" ]]; then
    TARGET="src"
  else
    echo "❌ No frontend dir found. Pass a path: $0 <path>" >&2
    exit 2
  fi
fi

if [[ ! -e "$TARGET" ]]; then
  echo "❌ Path not found: $TARGET" >&2
  exit 2
fi

# Collect files
if [[ -f "$TARGET" ]]; then
  FILES="$TARGET"
else
  FILES=$(find "$TARGET" -type f \( -name "*.vue" -o -name "*.css" -o -name "*.scss" \) \
    -not -path "*/node_modules/*" \
    -not -path "*/.nuxt/*" \
    -not -path "*/.output/*" \
    -not -path "*/dist/*" \
    -not -path "*/.git/*" 2>/dev/null || true)
fi

FILE_COUNT=$(echo "$FILES" | grep -c . || echo 0)
if [[ "$FILE_COUNT" -eq 0 ]]; then
  echo "❌ No .vue/.css/.scss files found in $TARGET" >&2
  exit 2
fi

echo "==> Frontend polish check (sensor 12, v1.12.0)"
echo "==> Scanning $FILE_COUNT files in $TARGET..."
echo

# Build whitelist from app.config.ts (ui.colors.* are valid token hex)
WHITELIST_HEX=$(grep -oE "['\"]#[0-9a-fA-F]{3,8}['\"]" app.config.ts 2>/dev/null \
  | tr -d "'\"" | sort -u | tr '\n' ' ' || true)

# Run Python companion (env vars + stdin)
SUGGEST="$SUGGEST_FIX" WHITELIST_HEX="$WHITELIST_HEX" \
  python3 "$PY_COMPANION" <<< "$FILES"
