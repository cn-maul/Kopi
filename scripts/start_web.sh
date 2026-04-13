#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"

ADDR="${1:-:8080}"
CONFIG_PATH="${2:-}"

"$SCRIPT_DIR/build.sh"

cd "$ROOT_DIR"
echo "正在启动 Web 页面: http://localhost${ADDR}"
if [[ -n "$CONFIG_PATH" ]]; then
  exec ./archiver -web -addr "$ADDR" -config "$CONFIG_PATH"
else
  exec ./archiver -web -addr "$ADDR"
fi
