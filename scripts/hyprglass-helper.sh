#!/usr/bin/env bash
set -euo pipefail

ROOT=$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)

install_packages() {
  mapfile -t pkgs < <(grep -vE '^\s*(#|$)' "$ROOT/packages/arch-core.txt" | sort -u)
  if [[ ${#pkgs[@]} -eq 0 ]]; then
    echo "No packages listed in packages/arch-core.txt; skipping package install."
    return 0
  fi
  exec sudo pacman -S --needed "${pkgs[@]}"
}

backup_configs() {
  local dst
  dst="$HOME/.config/hyprglass-backups/manual-$(date +%Y%m%d-%H%M%S)"
  mkdir -p "$dst"
  for d in hypr kitty waybar mako fuzzel gtk-3.0 qt6ct hyprglass hyprlock hypridle; do
    [[ -e "$HOME/.config/$d" ]] && cp -a "$HOME/.config/$d" "$dst/"
  done
  echo "Backed up installed configs to $dst"
}

case "$(basename "$0")" in
  install-packages.sh)
    install_packages
    ;;
  install-configs.sh)
    exec "$ROOT/install.sh" --no-packages "$@"
    ;;
  backup-configs.sh)
    backup_configs
    ;;
  validate-configs.sh)
    exec "$ROOT/scripts/check.sh" "$@"
    ;;
  hyprglass-helper.sh)
    cat <<HELP
Usage:
  scripts/install-packages.sh
  scripts/install-configs.sh [install.sh flags]
  scripts/backup-configs.sh
  scripts/validate-configs.sh
HELP
    ;;
  *)
    echo "unknown helper entrypoint: $(basename "$0")" >&2
    exit 2
    ;;
esac
