#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"

cd "$ROOT_DIR"
if [[ "${ENABLE_TRAY:-0}" == "1" ]]; then
  echo "正在编译带托盘支持的 archiver 可执行文件..."
  CGO_ENABLED=1 \
  GOOS="${GOOS:-linux}" \
  GOARCH="${GOARCH:-amd64}" \
  go build \
    -trimpath \
    -buildmode=exe \
    -tags "tray" \
    -o archiver
else
  echo "正在编译静态 archiver 可执行文件..."
  CGO_ENABLED=0 \
  GOOS="${GOOS:-linux}" \
  GOARCH="${GOARCH:-amd64}" \
  go build \
    -trimpath \
    -buildmode=exe \
    -tags "netgo osusergo" \
    -ldflags="-s -w -buildid=" \
    -o archiver
fi

echo "编译完成: $ROOT_DIR/archiver"
