#!/usr/bin/env bash
# ============================================================================
# git-meta-harness — Safe Commit Harness Sync
# ============================================================================
# Automatiza o `git add` + `git commit` + `git push` para syncs do
# framework em main, SEM usar `git add -A` (que capturou arquivos
# de feature branch 3 vezes: v1.8.0, v1.9.0, v1.10.0).
#
# Comportamento:
#   1. SEMPRE adiciona `harness/` e `VERSION` (são o sync do framework)
#   2. Auto-detecta customizações locais modificadas que matcham
#      whitelist explícita (ex.: .golangci.yml, .github/workflows/*.yml)
#   3. DETECTA arquivos modificados/untracked em paths que NÃO estão
#      na whitelist e PERGUNTA antes de prosseguir (proteção)
#   4. Pede confirmação antes de commitar
#   5. Mostra `git diff --cached --stat` antes do commit
#
# Uso:
#   ./bin/safe-commit-harness-sync.sh                    # add + commit + push
#   ./bin/safe-commit-harness-sync.sh --no-push          # só add + commit
#   ./bin/safe-commit-harness-sync.sh --message "msg"    # custom commit msg
#   ./bin/safe-commit-harness-sync.sh --dry-run          # mostra o que faria
#   ./bin/safe-commit-harness-sync.sh --auto             # skip confirmações
#                                                        # (use com cuidado)
#
# Saída:
#   - exit 0 = sucesso
#   - exit 1 = bloqueado (arquivos suspeitos, abortar)
#   - exit 2 = erro de git
# ============================================================================

set -e

# --- Args parsing ---
PUSH=1
DRY_RUN=0
AUTO=0
MESSAGE=""
while [[ $# -gt 0 ]]; do
  case "$1" in
    --no-push)   PUSH=0; shift ;;
    --dry-run)   DRY_RUN=1; shift ;;
    --auto)      AUTO=1; shift ;;
    -m|--message)
      [[ -z "${2:-}" ]] && { echo "❌ --message requires an argument" >&2; exit 2; }
      MESSAGE="$2"; shift 2 ;;
    -h|--help)
      sed -n '2,30p' "$0"
      exit 0 ;;
    *)
      echo "❌ Unknown flag: $1" >&2
      exit 2 ;;
  esac
done

# --- Pre-checks ---
command -v git >/dev/null 2>&1 || { echo "❌ git not found" >&2; exit 2; }
[[ -d .git ]] || { echo "❌ Not in a git repo" >&2; exit 2; }
[[ -d harness/ ]] || { echo "❌ harness/ not found — run from project root" >&2; exit 2; }
[[ -f VERSION ]] || { echo "❌ VERSION not found — run from project root" >&2; exit 2; }

# --- Whitelist: paths que podem ser commitados junto com sync do harness ---
# Esses são customizações LOCAIS esperadas (CI configs, etc) que
# geralmente são modificadas em paralelo com o sync do framework.
# Edite esta lista se seu projeto tem outros paths válidos.
WHITELIST_REGEX='^(\.github/workflows/.*\.yml|\.github/workflows/.*\.yaml|\.golangci\.yml|\.markdownlint\.json?$|\.eslintrc.*$|\.prettierrc.*$|Makefile|docker-compose.*\.yml$|docker-compose.*\.yaml$|deploy/.*\.sh$|docs/HOWTO.*\.md$)'

# --- Always-stage paths (framework sync core) ---
# These paths are ALWAYS staged (not just whitelisted) when modified.
# Add your framework-specific sync paths here.
ALWAYS_STAGE_PATHS=("harness/" "VERSION" "bin/safe-commit-harness-sync.sh" "bin/.gitignore")

# --- Detect suspicious changes ---
echo "==> Checking working tree for suspicious changes..."
echo

SUSPICIOUS=$(git status --short | awk '
{
  status = $1
  path = $2
  # Skip empty lines
  if (path == "") next
  # Ignore harness/, VERSION, .git, bin/ (this script)
  if (path ~ /^harness\//) next
  if (path == "VERSION") next
  if (path ~ /^\.git/) next
  # Skip other bin/ files but keep bin/safe-commit-harness-sync.sh
  # (handled separately by ALWAYS_STAGE_PATHS)
  if (path ~ /^bin\// && path != "bin/safe-commit-harness-sync.sh" && path != "bin/.gitignore") next
  if (path ~ /^\.DS_Store/) next
  # Modified or untracked
  if (status == "M" || status == "??" || status == "A" || status == "D" || status ~ /^R/) {
    print status "\t" path
  }
}')

if [[ -n "$SUSPICIOUS" ]]; then
  echo "⚠️  Found changes OUTSIDE harness/ and VERSION:"
  echo
  echo "$SUSPICIOUS" | while IFS=$'\t' read -r status path; do
    case "$status" in
      M) tag="modified" ;;
      \?\?) tag="untracked" ;;
      A) tag="added" ;;
      D) tag="deleted" ;;
      R*) tag="renamed" ;;
      *) tag="changed" ;;
    esac
    # Check if path matches whitelist
    if [[ "$path" =~ $WHITELIST_REGEX ]]; then
      echo "   [$tag] $path  (whitelisted — will be added)"
    else
      echo "   [$tag] $path  ⚠️  NOT WHITELISTED"
    fi
  done
  echo
  echo "These may be work from another branch (feature/fix/chore)."
  echo "If you commit them by accident, you'll pollute the harness sync."
  echo
fi

# --- Compute what to add ---
TO_ADD=()
for p in "${ALWAYS_STAGE_PATHS[@]}"; do
  TO_ADD+=("$p")
done

echo "==> Pre-selecting files to stage..."
echo
for path in $SUSPICIOUS; do
  if [[ "$path" =~ $WHITELIST_REGEX ]]; then
    TO_ADD+=("$path")
  fi
done

# Show diff summary
echo "Files that WILL be staged:"
for f in "${TO_ADD[@]}"; do
  echo "  + $f"
done
echo

# Check for non-whitelisted suspicious files
NON_WHITELIST=$(echo "$SUSPICIOUS" | awk -v re="$WHITELIST_REGEX" '
NF >= 2 {
  status = $1; path = $2
  if (path !~ re) print status "\t" path
}')

if [[ -n "$NON_WHITELIST" ]]; then
  echo "❌ BLOCKED: Found non-whitelisted files:"
  echo "$NON_WHITELIST" | while IFS=$'\t' read -r status path; do
    echo "   $status $path"
  done
  echo
  echo "These are probably from another branch (feature/fix/chore)."
  echo "Options:"
  echo "  1. Stash them first:    git stash push -u -- <paths>"
  echo "  2. Switch to feature:   git checkout <branch> && git stash pop"
  echo "  3. Re-evaluate: is the file really part of the sync?"
  echo
  echo "Refusing to commit to avoid mixing harness sync with other work."
  exit 1
fi

# --- Dry run ---
if [[ "$DRY_RUN" -eq 1 ]]; then
  echo "✅ Dry-run: would stage ${#TO_ADD[@]} paths and commit."
  echo "Run without --dry-run to proceed."
  exit 0
fi

# --- Confirm ---
if [[ "$AUTO" -eq 0 ]]; then
  read -p "Proceed with staging + commit? [y/N] " -n 1 -r
  echo
  if [[ ! $REPLY =~ ^[Yy]$ ]]; then
    echo "Aborted."
    exit 0
  fi
fi

# --- Stage ---
echo
echo "==> Staging ${#TO_ADD[@]} paths..."
for f in "${TO_ADD[@]}"; do
  git add -- "$f"
done

# Show what was actually staged
echo
echo "==> Staged files:"
git diff --cached --stat | head -50
echo

# --- Commit ---
if [[ -z "$MESSAGE" ]]; then
  MESSAGE="chore: harness sync to $(cat VERSION) (framework update)

Files auto-staged by bin/safe-commit-harness-sync.sh:
- harness/  (framework sync)
- VERSION   (version bump)
- whitelisted customizations (CI configs, etc)"
fi

echo "==> Committing..."
git commit -m "$MESSAGE"

# --- Push ---
if [[ "$PUSH" -eq 1 ]]; then
  BRANCH=$(git rev-parse --abbrev-ref HEAD)
  echo
  echo "==> Pushing to origin/$BRANCH..."
  git push origin "$BRANCH"
  echo
  echo "✅ Done. Harness sync to $(cat VERSION) pushed to origin/$BRANCH."
else
  echo
  echo "✅ Committed locally. Skipped push (--no-push)."
fi
