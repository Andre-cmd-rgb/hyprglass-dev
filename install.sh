#!/usr/bin/env bash
set -euo pipefail
YES=0 DRY=0 NO_PACKAGES=0 CONFIGS_ONLY=0 PREFIX="${HOME}/.local/bin"
for a in "$@"; do case "$a" in --yes) YES=1;; --dry-run) DRY=1;; --no-packages) NO_PACKAGES=1;; --configs-only) CONFIGS_ONLY=1; NO_PACKAGES=1;; *) echo "unknown option: $a"; exit 2;; esac; done
if [[ ${EUID} -eq 0 && $DRY -ne 1 ]]; then echo "Do not run install.sh fully as root. It uses sudo only where needed."; exit 1; fi
ROOT=$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)
run(){ echo "+ $*"; [[ $DRY -eq 1 ]] || "$@"; }
ask(){ [[ $YES -eq 1 || $DRY -eq 1 ]] && return 0; read -r -p "$1 [y/N] " ans; [[ $ans == y || $ans == Y ]]; }
if [[ $NO_PACKAGES -eq 0 ]]; then
  if command -v pacman >/dev/null 2>&1; then
    mapfile -t pkgs < <(grep -vE '^\s*(#|$)' "$ROOT/packages/arch-core.txt" | sort -u)
    run sudo pacman -S --needed "${pkgs[@]}"
  else
    echo "pacman not found; package install skipped. This is only valid off Arch or in dry validation."
  fi
fi
run mkdir -p "$PREFIX"
if [[ $CONFIGS_ONLY -eq 0 ]]; then run go build -o "$PREFIX/hyprglass" "$ROOT/cmd/hyprglass"; fi
backup="$HOME/.config/hyprglass-backups/$(date +%Y%m%d-%H%M%S)"
run mkdir -p "$backup"
copy_cfg(){ src="$1"; dst="$2"; if [[ -e "$dst" ]]; then run cp -a "$dst" "$backup/$(basename "$dst")"; fi; if [[ -e "$dst" && $YES -ne 1 && $DRY -ne 1 ]]; then ask "Overwrite $dst?" || return 0; fi; run mkdir -p "$(dirname "$dst")"; run cp -a "$src" "$dst"; }
copy_cfg "$ROOT/config/hypr" "$HOME/.config/hypr"
copy_cfg "$ROOT/config/kitty" "$HOME/.config/kitty"
copy_cfg "$ROOT/config/waybar" "$HOME/.config/waybar"
copy_cfg "$ROOT/config/hyprlock" "$HOME/.config/hyprlock"
copy_cfg "$ROOT/config/hypridle" "$HOME/.config/hypridle"
copy_cfg "$ROOT/config/mako" "$HOME/.config/mako"
copy_cfg "$ROOT/config/fuzzel" "$HOME/.config/fuzzel"
copy_cfg "$ROOT/config/gtk" "$HOME/.config/gtk-3.0"
copy_cfg "$ROOT/config/qt" "$HOME/.config/qt6ct"
run mkdir -p "$HOME/.config/hypr/assets" "$HOME/.config/hyprglass/docs"
run cp -a "$ROOT/assets/wallpapers" "$HOME/.config/hypr/assets/"
run cp -a "$ROOT/docs/shortcuts.md" "$HOME/.config/hyprglass/docs/shortcuts.md"
if [[ $DRY -eq 0 ]]; then chmod +x "$ROOT"/scripts/*.sh || true; fi
if command -v systemctl >/dev/null 2>&1 && [[ $DRY -eq 0 ]]; then
  if ask "Enable NetworkManager, bluetooth, and ModemManager services now?"; then
    sudo systemctl enable --now NetworkManager.service bluetooth.service ModemManager.service
  fi
fi
cat <<MSG
Hyprglass install complete.
Start Hyprland from your login/session manager or TTY.
Run: hyprglass doctor
Open TUIs with Super+W/B/M/A/D inside Hyprland.
Backups: $backup
MSG
