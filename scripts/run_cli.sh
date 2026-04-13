#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"

if [[ $# -lt 2 ]]; then
  echo "Usage: ./scripts/run_cli.sh <file_path> <category> [template_prefix] [config_path]"
  exit 1
fi

FILE_PATH="$1"
CATEGORY="$2"
TEMPLATE="${3:-}"
CONFIG_PATH="${4:-}"

"$SCRIPT_DIR/build.sh"

cd "$ROOT_DIR"
CMD=(./archiver -f "$FILE_PATH" -c "$CATEGORY")
if [[ -n "$TEMPLATE" ]]; then
  CMD+=( -t "$TEMPLATE" )
fi
if [[ -n "$CONFIG_PATH" ]]; then
  CMD+=( -config "$CONFIG_PATH" )
fi

exec "${CMD[@]}"
