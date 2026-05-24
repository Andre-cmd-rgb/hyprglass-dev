#!/usr/bin/env bash
set -euo pipefail
mkdir -p build
go build -buildvcs=false -o build/hyprglass ./cmd/hyprglass
