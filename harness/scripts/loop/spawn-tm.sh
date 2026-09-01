#!/usr/bin/env bash
# Cron (no_agent): wake team-manager if it is not already running.
# Always print one line so the cron UI is not "no execution".
# Do NOT attach a change-detector gate to this job.
set -euo pipefail
SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
ROOT="$(cd "$SCRIPT_DIR/../../.." && pwd)"
if [[ -f "$ROOT/harness/loop.env" ]]; then
  set -a
  # shellcheck disable=SC1091
  source "$ROOT/harness/loop.env"
  set +a
fi
PROFILE="${GMH_TM_PROFILE:-team-manager}"
TICK="${GMH_TICK_FILE:-$ROOT/harness/scripts/loop/team-manager-tick.md}"
SLUG="${GMH_PROJECT_SLUG:-$(basename "$ROOT")}"
LOG="${GMH_TM_LOG:-/tmp/gmh-${SLUG}-tm.log}"

if pgrep -fl "hermes -p ${PROFILE}" 2>/dev/null | grep -v pgrep >/dev/null; then
  echo "busy ${PROFILE}"
  exit 0
fi
nohup hermes -p "$PROFILE" chat --oneshot --in "$ROOT" --query-file "$TICK" >>"$LOG" 2>&1 &
echo "spawned ${PROFILE} pid $! log=$LOG"
