#!/usr/bin/env bash
set -euo pipefail
ROOT=$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)
case "$(basename "$0")" in
  install-packages.sh) exec "$ROOT/install.sh" --configs-only --dry-run ;;
  install-configs.sh) exec "$ROOT/install.sh" --no-packages ;;
  backup-configs.sh) mkdir -p "$HOME/.config/hyprglass-backups/manual-$(date +%Y%m%d-%H%M%S)" ;;
  validate-configs.sh) exec "$ROOT/scripts/check.sh" ;;
esac
