#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"

OUTPUT_NAME="${1:-archiver}"
GOOS_VALUE="${GOOS:-linux}"
GOARCH_VALUE="${GOARCH:-amd64}"
ADDR="${ADDR:-:8082}"
CONFIG_PATH="${CONFIG_PATH:-}"

cd "$ROOT_DIR"
echo "正在编译带托盘支持的可执行文件..."
echo "  输出文件: $ROOT_DIR/$OUTPUT_NAME"
echo "  GOOS/GOARCH: ${GOOS_VALUE}/${GOARCH_VALUE}"

CGO_ENABLED=1 \
GOOS="$GOOS_VALUE" \
GOARCH="$GOARCH_VALUE" \
go build \
  -trimpath \
  -buildmode=exe \
  -tags "tray" \
  -o "$OUTPUT_NAME"

echo "编译完成: $ROOT_DIR/$OUTPUT_NAME"

echo "正在启动 Web 页面: http://localhost${ADDR}"
if [[ -n "$CONFIG_PATH" ]]; then
  exec "$ROOT_DIR/$OUTPUT_NAME" -web -tray -addr "$ADDR" -config "$CONFIG_PATH"
else
  exec "$ROOT_DIR/$OUTPUT_NAME" -web -tray -addr "$ADDR"
fi
