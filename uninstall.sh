#!/usr/bin/env bash
set -euo pipefail
YES=0 REMOVE_PACKAGES=0
for a in "$@"; do case "$a" in --yes) YES=1;; --remove-packages) REMOVE_PACKAGES=1;; *) echo "unknown option: $a"; exit 2;; esac; done
ask(){ [[ $YES -eq 1 ]] && return 0; read -r -p "$1 [y/N] " ans; [[ $ans == y || $ans == Y ]]; }
echo "This removes Hyprglass-installed configs, not backups."
ask "Continue?" || exit 0
for d in hypr kitty waybar hyprlock hypridle mako fuzzel hyprglass; do [[ -e "$HOME/.config/$d" ]] && rm -rf "$HOME/.config/$d"; done
rm -f "$HOME/.local/bin/hyprglass"
if [[ $REMOVE_PACKAGES -eq 1 ]]; then echo "Package removal intentionally not automated in V0. Remove packages manually after reviewing dependencies."; fi
echo "Hyprglass configs removed. Backups were left untouched."
