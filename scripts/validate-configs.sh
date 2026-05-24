#!/usr/bin/env bash
set -euo pipefail
ROOT=$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)
case "$(basename "$0")" in
  install-packages.sh)
    mapfile -t pkgs < <(grep -vE '^\s*(#|$)' "$ROOT/packages/arch-core.txt" | sort -u)
    exec sudo pacman -S --needed "${pkgs[@]}"
    ;;
  install-configs.sh)
    exec "$ROOT/install.sh" --no-packages
    ;;
  backup-configs.sh)
    dst="$HOME/.config/hyprglass-backups/manual-$(date +%Y%m%d-%H%M%S)"
    mkdir -p "$dst"
    for d in hypr kitty waybar hyprlock hypridle mako fuzzel; do
      [[ -e "$HOME/.config/$d" ]] && cp -a "$HOME/.config/$d" "$dst/"
    done
    echo "Backed up installed configs to $dst"
    ;;
  validate-configs.sh)
    exec "$ROOT/scripts/check.sh"
    ;;
esac
