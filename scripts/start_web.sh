#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"

ADDR="${1:-:8082}"
CONFIG_PATH="${2:-}"

"$SCRIPT_DIR/build.sh"

cd "$ROOT_DIR"
echo "Starting Web UI at http://localhost${ADDR}"
if [[ -n "$CONFIG_PATH" ]]; then
  exec ./archiver -web -tray -addr "$ADDR" -config "$CONFIG_PATH"
else
  exec ./archiver -web -tray -addr "$ADDR"
fi
