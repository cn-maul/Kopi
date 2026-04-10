#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"

cd "$ROOT_DIR"
echo "Building static archiver binary..."
CGO_ENABLED=0 \
GOOS="${GOOS:-linux}" \
GOARCH="${GOARCH:-amd64}" \
go build \
  -trimpath \
  -buildmode=exe \
  -tags "netgo osusergo" \
  -ldflags="-s -w -buildid=" \
  -o archiver
