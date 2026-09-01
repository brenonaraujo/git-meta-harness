#!/usr/bin/env bash
# Detached persona spawn (not tracked by the parent Hermes chat).
# Usage: spawn-persona.sh <profile> <brief-file> [repo]
set -euo pipefail
profile="${1:?profile}"
brief="${2:?brief file}"
SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
ROOT="$(cd "$SCRIPT_DIR/../../.." && pwd)"
if [[ -f "$ROOT/harness/loop.env" ]]; then
  set -a
  # shellcheck disable=SC1091
  source "$ROOT/harness/loop.env"
  set +a
fi
repo="${3:-$ROOT}"
id="$(basename "$brief" .md)"
log="/tmp/gmh-${id}.log"
if pgrep -fl "hermes -p ${profile} " 2>/dev/null | grep -v pgrep >/dev/null; then
  echo "busy: $profile"
  exit 0
fi
nohup hermes -p "$profile" chat --oneshot --in "$repo" --query-file "$brief" >>"$log" 2>&1 &
echo "spawned $profile pid $! log=$log"
