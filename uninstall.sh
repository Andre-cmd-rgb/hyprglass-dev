#!/usr/bin/env bash
set -euo pipefail

YES=0 REMOVE_PACKAGES=0
for a in "$@"; do
  case "$a" in
    --yes) YES=1 ;;
    --remove-packages) REMOVE_PACKAGES=1 ;;
    *) echo "unknown option: $a"; exit 2 ;;
  esac
done

ask() {
  [[ $YES -eq 1 ]] && return 0
  read -r -p "$1 [y/N] " ans
  [[ $ans == y || $ans == Y ]]
}

backup="$HOME/.config/hyprglass-backups/uninstall-$(date +%Y%m%d-%H%M%S)"
backup_configs() {
  mkdir -p "$backup"
  for d in hypr kitty waybar mako fuzzel gtk-3.0 gtk-4.0 qt6ct hyprglass hyprlock hypridle; do
    [[ -e "$HOME/.config/$d" ]] && cp -a "$HOME/.config/$d" "$backup/"
  done
}

echo "This removes Hyprglass-installed configs after taking a backup."
ask "Continue?" || exit 0
backup_configs
for d in hypr kitty waybar mako fuzzel gtk-3.0 gtk-4.0 qt6ct hyprglass hyprlock hypridle; do
  [[ -e "$HOME/.config/$d" ]] && rm -rf "$HOME/.config/$d"
done
rm -f "$HOME/.local/bin/hyprglass"
if [[ $REMOVE_PACKAGES -eq 1 ]]; then echo "Package removal intentionally not automated in V0. Remove packages manually after reviewing dependencies."; fi
echo "Hyprglass configs removed. Backup saved to: $backup"
