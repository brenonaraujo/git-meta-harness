#!/usr/bin/env bash
# ============================================================================
# git-meta-harness — Setup Playwright for visual screenshots (v1.12.0)
# ============================================================================
# Instala Playwright + Chromium no projeto pra rodar
# harness/scripts/visual/playwright-screenshot.mjs.
#
# Uso:
#   ./harness/scripts/visual/setup-playwright-screenshot.sh
#   ./harness/scripts/visual/setup-playwright-screenshot.sh --browser firefox
#
# Idempotente: detecta instalação existente e pula.
# ============================================================================

set -e

cd "$(git rev-parse --show-toplevel 2>/dev/null || pwd)"

BROWSER="chromium"
for arg in "$@"; do
  case "$arg" in
    --browser) shift; BROWSER="${1:-chromium}" ;;
  esac
done

echo "==> Setting up Playwright (v1.12.0)"
echo "==> Browser: $BROWSER"
echo

# 1. Check if Playwright is already in devDependencies
HAS_PLAYWRIGHT=$(node -e "
const pkg = require('./package.json');
const deps = { ...pkg.dependencies, ...pkg.devDependencies };
process.stdout.write(deps.playwright ? '1' : '0');
" 2>/dev/null || echo "0")

if [[ "$HAS_PLAYWRIGHT" == "1" ]]; then
  echo "  ✓ playwright already in package.json"
else
  echo "  → Installing playwright as devDependency..."
  if command -v pnpm >/dev/null 2>&1; then
    pnpm add -D playwright
  elif command -v npm >/dev/null 2>&1; then
    npm install --save-dev playwright
  else
    echo "❌ Neither pnpm nor npm found. Install Node.js first." >&2
    exit 1
  fi
fi

# 2. Install browser binary
echo "  → Installing $BROWSER browser binary..."
if command -v pnpm >/dev/null 2>&1; then
  pnpm exec playwright install "$BROWSER"
else
  npx playwright install "$BROWSER"
fi

# 3. Add npm script (idempotent)
echo "  → Adding 'screenshot' script to package.json..."
node -e "
const fs = require('fs');
const pkg = JSON.parse(fs.readFileSync('./package.json', 'utf8'));
pkg.scripts = pkg.scripts || {};
if (!pkg.scripts.screenshot) {
  pkg.scripts.screenshot = 'node harness/scripts/visual/playwright-screenshot.mjs';
}
if (!pkg.scripts['screenshot:setup']) {
  pkg.scripts['screenshot:setup'] = 'bash harness/scripts/visual/setup-playwright-screenshot.sh';
}
fs.writeFileSync('./package.json', JSON.stringify(pkg, null, 2) + '\n');
"

echo
echo "✅ Playwright setup complete."
echo
echo "Next steps:"
echo "  1. Start your dev server:  pnpm dev  (or npm run dev)"
echo "  2. Run screenshot:         pnpm screenshot"
echo "  3. Run sensor 12:          ./harness/scripts/check-frontend-polish.sh"
echo
echo "Output: qa/screenshots/<route>-<viewport>.png"
echo
