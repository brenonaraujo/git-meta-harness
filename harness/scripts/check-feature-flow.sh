#!/usr/bin/env bash
# ============================================================================
# git-meta-harness — Check Feature Flow (Sensor 13, v1.13.0)
# ============================================================================
# Valida que toda issue `type/feature` passou pelo flow canônico:
#   domain-expert (label `refined` + comentário com ACs)
#   → solutions-architect (label `ready` + comentário com DoD)
#   → in-progress (builder)
#
# BLOQUEANTE (exit 1) — feature sem refinamento/DoD = retrocesso
# silencioso pro time (builder recebe só a descrição, sem contexto).
#
# Uso:
#   ./harness/scripts/check-feature-flow.sh                # all type/feature
#   ./harness/scripts/check-feature-flow.sh 48             # one issue (#48)
#   ./harness/scripts/check-feature-flow.sh --repo org/repo # explicit repo
#
# Saída:
#   - exit 0 = flow OK
#   - exit 1 = violação (bloquear, listar recovery)
#   - exit 2 = erro de uso
#
# Lição do Mandaí v2 (jul/2026): épico #48 (F7+F8+F10) tem
# type/feature e ZERO comentários de domain-expert ou architect.
# As 4 sub-issues (#49-#52) também sem type/refined/ready.
# ============================================================================

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PY_COMPANION="${SCRIPT_DIR}/visual/check_feature_flow.py"

if [[ ! -f "$PY_COMPANION" ]]; then
  echo "❌ Python companion not found: $PY_COMPANION" >&2
  exit 2
fi

REPO=""
ISSUE_NUMBER=""

for arg in "$@"; do
  case "$arg" in
    --repo) shift; REPO="${1:-}" ;;
    -h|--help)
      echo "Usage: $0 [--repo owner/name] [issue_number]"
      echo ""
      echo "  --repo owner/name   GitHub repo (default: auto-detect from git)"
      echo "  issue_number        Check only this issue (default: all type/feature)"
      exit 0 ;;
    *) ISSUE_NUMBER="$arg" ;;
  esac
done

# Auto-detect repo from git remote (origin)
if [[ -z "$REPO" ]]; then
  REPO=$(git remote get-url origin 2>/dev/null \
    | sed -E 's#.*[:/]([^/]+)/([^/]+)(\.git)?/?$#\1/\2#' \
    | head -1 || true)
  # Strip trailing .git if any
  REPO="${REPO%.git}"
  if [[ -z "$REPO" || "$REPO" == "/" ]]; then
    echo "❌ Could not auto-detect repo. Use --repo owner/name" >&2
    exit 2
  fi
fi

if ! command -v gh >/dev/null 2>&1; then
  echo "❌ gh CLI not found. Install GitHub CLI first." >&2
  exit 2
fi

if ! gh auth status >/dev/null 2>&1; then
  echo "❌ gh CLI not authenticated. Run 'gh auth login' first." >&2
  exit 2
fi

echo "==> Feature flow check (sensor 13, v1.13.0)"
echo "==> Repo: $REPO"
if [[ -n "$ISSUE_NUMBER" ]]; then
  echo "==> Issue: #$ISSUE_NUMBER"
else
  echo "==> Issues: all open type/feature"
fi
echo

# Get issues JSON (one per line) and pass to Python companion
if [[ -n "$ISSUE_NUMBER" ]]; then
  JSON=$(gh issue view "$ISSUE_NUMBER" --repo "$REPO" \
    --json number,title,body,labels,comments,state 2>/dev/null || true)
  if [[ -z "$JSON" ]]; then
    echo "❌ Could not fetch issue #$ISSUE_NUMBER from $REPO" >&2
    exit 2
  fi
  echo "$JSON"
else
  # All open type/feature issues (gh issue list --label + jq)
  # Simpler: use gh issue list and filter in Python
  gh issue list --repo "$REPO" --state all --limit 200 \
    --json number,title,body,labels,comments,state 2>/dev/null
fi | REPO="$REPO" python3 "$PY_COMPANION"
