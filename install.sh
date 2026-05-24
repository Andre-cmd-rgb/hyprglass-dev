#!/usr/bin/env bash
set -euo pipefail

# ── Argument parsing ─────────────────────────────────────────────────────────
YES=0 DRY=0 NO_PACKAGES=0 CONFIGS_ONLY=0 UPDATE=0
PREFIX="${HOME}/.local/bin"

for a in "$@"; do
  case "$a" in
    --yes)          YES=1 ;;
    --dry-run)      DRY=1 ;;
    --no-packages)  NO_PACKAGES=1 ;;
    --configs-only) CONFIGS_ONLY=1; NO_PACKAGES=1 ;;
    --update)       UPDATE=1; YES=1; NO_PACKAGES=1 ;;
    --help|-h)
      cat <<HELP
Usage: install.sh [options]

  (no flags)       Full install: packages + binary + configs
  --update         Pull latest changes from git and overwrite all configs.
                   Skips packages. Implies --yes. Safe to re-run any time.
  --no-packages    Skip pacman; install binary + configs only
  --configs-only   Skip pacman and binary build; configs only
  --yes            Overwrite existing configs without prompting
  --dry-run        Print what would happen; make no changes
  --help           Show this help

Examples:
  ./install.sh                   # first-time install
  ./install.sh --update          # pull + refresh all configs (keeps backups)
  ./install.sh --dry-run         # preview only
HELP
      exit 0
      ;;
    *) echo "unknown option: $a"; exit 2 ;;
  esac
done

if [[ ${EUID} -eq 0 && $DRY -ne 1 ]]; then
  echo "Do not run install.sh as root. It calls sudo only where needed."
  exit 1
fi

ROOT=$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)
run(){ echo "+ $*"; [[ $DRY -eq 1 ]] || "$@"; }
ask(){
  [[ $YES -eq 1 || $DRY -eq 1 ]] && return 0
  read -r -p "$1 [y/N] " ans
  [[ $ans == y || $ans == Y ]]
}

# ── --update: pull latest from git ──────────────────────────────────────────
if [[ $UPDATE -eq 1 ]]; then
  echo "==> Hyprglass update"
  if [[ -d "$ROOT/.git" ]]; then
    echo "Git repo detected — pulling latest changes..."
    run git -C "$ROOT" pull --ff-only
  else
    echo "Not a git repo (installed from zip). Skipping git pull."
    echo "To get updates: download the latest zip from the repo and re-run ./install.sh --update"
  fi
fi

# ── Packages ─────────────────────────────────────────────────────────────────
if [[ $NO_PACKAGES -eq 0 ]]; then
  if command -v pacman >/dev/null 2>&1; then
    mapfile -t pkgs < <(grep -vE '^\s*(#|$)' "$ROOT/packages/arch-core.txt" | sort -u)
    run sudo pacman -S --needed "${pkgs[@]}"
  else
    echo "pacman not found; skipping package install."
  fi
fi

# ── Build binary ─────────────────────────────────────────────────────────────
run mkdir -p "$PREFIX"
if [[ $CONFIGS_ONLY -eq 0 ]]; then
  run go build -buildvcs=false -o "$PREFIX/hyprglass" "$ROOT/cmd/hyprglass"
fi

# ── Back up + copy configs ───────────────────────────────────────────────────
backup="$HOME/.config/hyprglass-backups/$(date +%Y%m%d-%H%M%S)"
run mkdir -p "$backup"

copy_cfg() {
  local src="$1" dst="$2"
  # Back up whatever is currently installed
  if [[ -e "$dst" ]]; then
    run cp -a "$dst" "$backup/$(basename "$dst")"
  fi
  # Ask before overwriting (skipped when --yes / --update / --dry-run)
  if [[ -e "$dst" && $YES -ne 1 && $DRY -ne 1 ]]; then
    ask "Overwrite $dst?" || return 0
  fi
  run mkdir -p "$(dirname "$dst")"
  run cp -a "$src" "$dst"
}

copy_cfg "$ROOT/config/hypr"     "$HOME/.config/hypr"
copy_cfg "$ROOT/config/kitty"    "$HOME/.config/kitty"
copy_cfg "$ROOT/config/waybar"   "$HOME/.config/waybar"
copy_cfg "$ROOT/config/hyprlock" "$HOME/.config/hyprlock"
copy_cfg "$ROOT/config/hypridle" "$HOME/.config/hypridle"
copy_cfg "$ROOT/config/mako"     "$HOME/.config/mako"
copy_cfg "$ROOT/config/fuzzel"   "$HOME/.config/fuzzel"
copy_cfg "$ROOT/config/gtk"      "$HOME/.config/gtk-3.0"
copy_cfg "$ROOT/config/qt"       "$HOME/.config/qt6ct"

run mkdir -p "$HOME/.config/hypr/assets" "$HOME/.config/hyprglass/docs"
run cp -a "$ROOT/assets/wallpapers" "$HOME/.config/hypr/assets/"
run cp -a "$ROOT/docs/shortcuts.md" "$HOME/.config/hyprglass/docs/shortcuts.md"

if [[ $DRY -eq 0 ]]; then chmod +x "$ROOT"/scripts/*.sh || true; fi

# ── Systemd services ─────────────────────────────────────────────────────────
if command -v systemctl >/dev/null 2>&1 && [[ $DRY -eq 0 && $UPDATE -eq 0 ]]; then
  if ask "Enable NetworkManager, bluetooth, and ModemManager services now?"; then
    sudo systemctl enable --now \
      NetworkManager.service bluetooth.service ModemManager.service
  fi
fi

# ── Reload Hyprland if running ───────────────────────────────────────────────
if [[ $DRY -eq 0 && -n "${HYPRLAND_INSTANCE_SIGNATURE:-}" ]]; then
  echo "Hyprland session detected — reloading config..."
  run hyprctl reload
  echo "Config reloaded."
fi

# ── Summary ──────────────────────────────────────────────────────────────────
if [[ $UPDATE -eq 1 ]]; then
  cat <<MSG

Hyprglass update complete.
All configs refreshed. Previous configs backed up to:
  $backup

If Hyprland was running, config was reloaded automatically.
If not, reload manually with: hyprctl reload
Run: hyprglass doctor
MSG
else
  cat <<MSG

Hyprglass install complete.
Start Hyprland from your login/session manager or TTY.
Run: hyprglass doctor
Open TUIs with Super+W/B/M/A/D inside Hyprland.
Backups: $backup
MSG
fi
