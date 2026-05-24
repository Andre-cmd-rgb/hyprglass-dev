#!/usr/bin/env bash
set -euo pipefail
ROOT=$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)
exec -a "$(basename "$0")" "$ROOT/scripts/hyprglass-helper.sh" "$@"
