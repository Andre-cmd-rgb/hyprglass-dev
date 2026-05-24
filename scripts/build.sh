#!/usr/bin/env bash
set -euo pipefail
mkdir -p build
go build -o build/hyprglass ./cmd/hyprglass
