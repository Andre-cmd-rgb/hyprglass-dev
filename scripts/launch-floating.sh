#!/usr/bin/env bash
set -euo pipefail
name=${1:?usage: launch-floating.sh <name> [args...]}
shift || true
exec kitty --class hyprglass-float --title "Hyprglass :: ${name}" -e hyprglass "$name" "$@"
