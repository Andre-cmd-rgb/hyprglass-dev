#!/usr/bin/env bash
set -euo pipefail
ROOT=$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)
VERSION=${HYPRGLASS_VERSION:-$(tr -d "[:space:]" < "$ROOT/VERSION" 2>/dev/null || git -C "$ROOT" describe --tags --always --dirty 2>/dev/null || printf dev)}
if [[ -z "${GOCACHE:-}" ]]; then
  export GOCACHE="${XDG_CACHE_HOME:-$HOME/.cache}/go-build"
fi
if ! mkdir -p "$GOCACHE" 2>/dev/null || [[ ! -w "$GOCACHE" ]]; then
  export GOCACHE="${TMPDIR:-/tmp}/hyprglass-go-cache"
  mkdir -p "$GOCACHE"
fi
mkdir -p "$ROOT/build"
go build -buildvcs=false -ldflags "-s -w -X main.version=$VERSION -X main.sourceRoot=$ROOT" -o "$ROOT/build/hyprglass" "$ROOT/cmd/hyprglass"
