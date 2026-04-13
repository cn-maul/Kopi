#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"

if [[ $# -lt 2 ]]; then
  echo "用法: ./scripts/run_cli.sh <文件路径> <分类> [模板前缀] [配置文件路径]"
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

echo "开始执行归档命令..."
exec "${CMD[@]}"
