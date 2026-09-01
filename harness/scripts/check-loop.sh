#!/usr/bin/env bash
# Sensor 14 — loop liveness (filesystem). Exit 1 on any fail.
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/../.." && pwd)"
FAILS=0
pass() { echo "  ✅ $1"; }
fail() { echo "  ❌ $1"; FAILS=$((FAILS+1)); }

echo "🔎 Sensor 14 — loop liveness"
echo "Root: $ROOT"

need() {
  if [[ -f "$1" ]]; then pass "$2"; else fail "$2 ($1 missing)"; fi
}
need "$ROOT/harness/scripts/loop/spawn-tm.sh" "spawn-tm.sh"
need "$ROOT/harness/scripts/loop/spawn-persona.sh" "spawn-persona.sh"
need "$ROOT/harness/scripts/loop/team-manager-tick.md" "team-manager-tick.md"
need "$ROOT/harness/scripts/loop/orchestrator-verify.md" "orchestrator-verify.md"
need "$ROOT/harness/loop.env" "loop.env"

if grep -q 'nohup' "$ROOT/harness/scripts/loop/spawn-tm.sh" 2>/dev/null; then
  pass "spawn-tm uses nohup"
else
  fail "spawn-tm missing nohup"
fi
if grep -q -- '--monitor' "$ROOT/harness/scripts/loop/spawn-tm.sh" 2>/dev/null; then
  fail "spawn-tm mentions --monitor (forbidden)"
else
  pass "spawn-tm has no --monitor"
fi
if grep -Eqi 'NÃO escreva código|do not write' "$ROOT/harness/scripts/loop/team-manager-tick.md" 2>/dev/null; then
  pass "tick forbids implementing"
else
  fail "tick does not forbid implementing"
fi
if [[ -f "$ROOT/harness/personas/domain-expert.md" ]]; then
  fail "generic domain-expert.md present (invariant 12)"
else
  pass "no generic domain-expert.md"
fi
if [[ -f "$ROOT/.github/workflows/ci.yml" ]]; then
  pass "ci.yml present"
else
  fail "ci.yml missing"
fi

echo
if [[ "$FAILS" -gt 0 ]]; then
  echo "❌ $FAILS fail(s)"
  exit 1
fi
echo "✅ loop liveness OK"
exit 0
