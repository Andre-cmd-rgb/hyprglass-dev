#!/usr/bin/env bash
set -euo pipefail

# Argument parsing
YES=0 DRY=0 NO_PACKAGES=0 CONFIGS_ONLY=0 UPDATE=0
PREFIX="${HYPRGLASS_PREFIX:-${HOME}/.local/bin}"

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

ensure_executable_bits() {
  if [[ $DRY -eq 1 ]]; then
    echo "+ chmod +x $ROOT/install.sh $ROOT/uninstall.sh $ROOT/scripts/*"
    return 0
  fi
  chmod +x "$ROOT/install.sh" "$ROOT/uninstall.sh" "$ROOT"/scripts/* 2>/dev/null || true
}

build_version() {
  local v=""
  if [[ -n "${HYPRGLASS_VERSION:-}" ]]; then
    v="$HYPRGLASS_VERSION"
  elif [[ -d "$ROOT/.git" ]] && command -v git >/dev/null 2>&1; then
    v=$(git -C "$ROOT" describe --tags --always --dirty 2>/dev/null || true)
  fi
  printf '%s\n' "${v:-0.1.0}"
}

ensure_go_cache() {
  local cache
  cache="${GOCACHE:-${XDG_CACHE_HOME:-$HOME/.cache}/go-build}"
  if mkdir -p "$cache" 2>/dev/null && [[ -w "$cache" ]]; then
    export GOCACHE="$cache"
    return 0
  fi
  export GOCACHE="${TMPDIR:-/tmp}/hyprglass-go-cache"
  mkdir -p "$GOCACHE"
}

path_expr() {
  if [[ "$PREFIX" == "$HOME/.local/bin" ]]; then
    printf '%s\n' '$HOME/.local/bin'
  else
    printf '%s\n' "$PREFIX"
  fi
}

path_block_posix() {
  local bin_path
  bin_path=$(path_expr)
  cat <<EOF

# >>> hyprglass PATH >>>
case ":\$PATH:" in
  *":$bin_path:"*) ;;
  *) export PATH="$bin_path:\$PATH" ;;
esac
# <<< hyprglass PATH <<<
EOF
}

path_block_fish() {
  local bin_path
  bin_path=$(path_expr)
  cat <<EOF

# >>> hyprglass PATH >>>
if functions -q fish_add_path
    fish_add_path -g "$bin_path"
else if not contains "$bin_path" \$PATH
    set -gx PATH "$bin_path" \$PATH
end
# <<< hyprglass PATH <<<
EOF
}

ensure_path_block() {
  local file="$1" shell_kind="$2"
  [[ -n "$file" ]] || return 0
  if [[ $DRY -eq 1 ]]; then
    echo "+ ensure $PREFIX is on PATH in $file"
    return 0
  fi
  mkdir -p "$(dirname "$file")"
  touch "$file"
  if grep -Fq ">>> hyprglass PATH >>>" "$file"; then
    return 0
  fi
  if [[ "$shell_kind" == fish ]]; then
    path_block_fish >>"$file"
  else
    path_block_posix >>"$file"
  fi
}

configure_path() {
  local shell_name rc_file profile bin_path
  shell_name=$(basename "${SHELL:-}")
  profile="$HOME/.profile"
  bin_path=$(path_expr)

  case "$shell_name" in
    bash) rc_file="$HOME/.bashrc" ;;
    zsh) rc_file="$HOME/.zshrc" ;;
    fish) rc_file="$HOME/.config/fish/config.fish" ;;
    *) rc_file="" ;;
  esac

  if [[ "$shell_name" == fish ]]; then
    ensure_path_block "$rc_file" fish
  elif [[ -n "$rc_file" ]]; then
    ensure_path_block "$rc_file" posix
  fi
  ensure_path_block "$profile" posix

  if [[ $DRY -eq 1 ]]; then
    cat <<MSG

PATH would be configured for future shells.
After a real install, open a new terminal or run:
  exec ${SHELL:-/bin/sh} -l

For the current terminal only after install:
  export PATH="$bin_path:\$PATH"
MSG
    return 0
  fi

  cat <<MSG

PATH configured for future shells.
Open a new terminal, or run:
  exec ${SHELL:-/bin/sh} -l

For this current terminal only, you can also run:
  export PATH="$bin_path:\$PATH"
MSG
}

write_source_root() {
  local dst="$HOME/.config/hyprglass/source-root"
  if [[ $DRY -eq 1 ]]; then
    echo "+ write source root to $dst"
    return 0
  fi
  mkdir -p "$(dirname "$dst")"
  printf '%s\n' "$ROOT" >"$dst"
}

configure_desktop_theme() {
  if [[ $DRY -eq 1 ]]; then
    echo "+ set GTK/libadwaita dark preference when gsettings is available"
    return 0
  fi
  if command -v gsettings >/dev/null 2>&1; then
    gsettings set org.gnome.desktop.interface color-scheme prefer-dark 2>/dev/null || true
    gsettings set org.gnome.desktop.interface gtk-theme Adwaita-dark 2>/dev/null || true
  fi
}

write_installed_hyprpaper_config() {
  local wallpaper="$HOME/.config/hypr/assets/wallpapers/hyprglass-dusk.png"
  local dst="$HOME/.config/hypr/hyprpaper.conf"
  if [[ $DRY -eq 1 ]]; then
    echo "+ write absolute wallpaper path to $dst"
    return 0
  fi
  mkdir -p "$(dirname "$dst")"
  cat >"$dst" <<EOF
# Hyprglass hyprpaper configuration
preload = $wallpaper
wallpaper = , $wallpaper
splash = false
EOF
}

ensure_current_shell_command() {
  local target="$PREFIX/hyprglass"
  local link="/usr/local/bin/hyprglass"

  [[ $CONFIGS_ONLY -eq 0 ]] || return 0
  [[ -x "$target" || $DRY -eq 1 ]] || return 0

  if [[ ":$PATH:" == *":$PREFIX:"* ]]; then
    return 0
  fi

  if [[ ":$PATH:" != *":/usr/local/bin:"* ]]; then
    cat <<MSG

$PREFIX is not on this terminal's current PATH, and /usr/local/bin is not either.
Future shells are configured. For this terminal run:
  export PATH="$PREFIX:\$PATH"
MSG
    return 0
  fi

  if [[ $DRY -eq 1 ]]; then
    echo "+ link $link -> $target so hyprglass works in the current terminal"
    return 0
  fi

  if [[ -w /usr/local/bin ]]; then
    ln -sf "$target" "$link"
    echo "Linked $link -> $target"
  elif command -v sudo >/dev/null 2>&1; then
    sudo ln -sf "$target" "$link"
    echo "Linked $link -> $target"
  else
    cat <<MSG

Could not create $link because sudo is unavailable.
Future shells are configured. For this terminal run:
  export PATH="$PREFIX:\$PATH"
MSG
  fi
}

restart_session_components() {
  [[ -n "${HYPRLAND_INSTANCE_SIGNATURE:-}" ]] || return 0
  [[ $DRY -eq 0 ]] || return 0

  start_bg() {
    local name="$1"
    shift
    command -v "$1" >/dev/null 2>&1 || return 0
    if command -v setsid >/dev/null 2>&1; then
      setsid -f "$@" >/dev/null 2>&1 || true
    else
      nohup "$@" >/dev/null 2>&1 &
    fi
    echo "Started $name."
  }

  echo "Restarting Hyprglass session components..."
  pkill -x waybar 2>/dev/null || true
  pkill -x hyprpaper 2>/dev/null || true
  pkill -x mako 2>/dev/null || true
  pkill -x hypridle 2>/dev/null || true
  start_bg hyprpaper hyprpaper
  start_bg waybar waybar
  start_bg mako mako
  start_bg hypridle hypridle
}

verify_installed_configs() {
  [[ $DRY -eq 0 ]] || return 0
  [[ -f "$HOME/.config/hypr/hyprland.conf" ]] || { echo "Install failed: missing ~/.config/hypr/hyprland.conf"; exit 1; }
  [[ -f "$HOME/.config/hypr/hyprpaper.conf" ]] || { echo "Install failed: missing ~/.config/hypr/hyprpaper.conf"; exit 1; }
  [[ -f "$HOME/.config/waybar/config.jsonc" ]] || { echo "Install failed: missing ~/.config/waybar/config.jsonc"; exit 1; }
  [[ -f "$HOME/.config/waybar/style.css" ]] || { echo "Install failed: missing ~/.config/waybar/style.css"; exit 1; }
  [[ -f "$HOME/.config/hypr/assets/wallpapers/hyprglass-dusk.png" ]] || { echo "Install failed: missing Hyprglass wallpaper asset"; exit 1; }
  grep -Fq "Hyprglass main Hyprland config" "$HOME/.config/hypr/hyprland.conf" || {
    echo "Install failed: ~/.config/hypr/hyprland.conf was not replaced with Hyprglass config"
    exit 1
  }
}

ensure_executable_bits

# --update: pull latest from git
if [[ $UPDATE -eq 1 ]]; then
  echo "==> Hyprglass update"
  if [[ -d "$ROOT/.git" ]]; then
    if [[ $DRY -eq 0 ]]; then
      if ! git -C "$ROOT" diff --quiet || ! git -C "$ROOT" diff --cached --quiet || [[ -n $(git -C "$ROOT" ls-files --others --exclude-standard) ]]; then
        stash_name="hyprglass-auto-update-$(date +%Y%m%d-%H%M%S)"
        echo "Local repo changes detected - saving them to git stash: $stash_name"
        git -C "$ROOT" stash push -u -m "$stash_name"
      fi
    else
      echo "+ check for local changes and stash before pull if needed"
    fi
    echo "Git repo detected - pulling latest changes..."
    run git -C "$ROOT" pull --ff-only
    if [[ $DRY -eq 0 && "${HYPRGLASS_UPDATE_REEXECED:-0}" != 1 ]]; then
      echo "Restarting installer after pull so any updated install logic is used..."
      exec env HYPRGLASS_UPDATE_REEXECED=1 bash "$ROOT/install.sh" "$@"
    fi
  else
    echo "Not a git repo (installed from zip). Skipping git pull."
    echo "To get updates: download the latest zip from the repo and re-run ./install.sh --update"
  fi
fi

# Packages
if [[ $NO_PACKAGES -eq 0 ]]; then
  if command -v pacman >/dev/null 2>&1; then
    mapfile -t pkgs < <(grep -vE '^\s*(#|$)' "$ROOT/packages/arch-core.txt" | sort -u)
    if [[ ${#pkgs[@]} -eq 0 ]]; then
      echo "No packages listed in packages/arch-core.txt; skipping package install."
    else
      run sudo pacman -S --needed "${pkgs[@]}"
    fi
  else
    echo "pacman not found; skipping package install."
  fi
fi

# Build binary
run mkdir -p "$PREFIX"
if [[ $CONFIGS_ONLY -eq 0 ]]; then
  VERSION=$(build_version)
  [[ $DRY -eq 1 ]] || ensure_go_cache
  run go build -buildvcs=false -ldflags "-s -w -X main.version=$VERSION -X main.sourceRoot=$ROOT" -o "$PREFIX/hyprglass" "$ROOT/cmd/hyprglass"
fi

# Back up + copy configs
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
  if [[ -d "$src" ]]; then
    if [[ -e "$dst" && ! -d "$dst" ]]; then
      run rm -f "$dst"
    fi
    run mkdir -p "$dst"
    run cp -a "$src"/. "$dst"/
  else
    if [[ -d "$dst" ]]; then
      run rm -rf "$dst"
    fi
    run mkdir -p "$(dirname "$dst")"
    run cp -af "$src" "$dst"
  fi
}

copy_cfg "$ROOT/config/hypr"     "$HOME/.config/hypr"
copy_cfg "$ROOT/config/kitty"    "$HOME/.config/kitty"
copy_cfg "$ROOT/config/waybar"   "$HOME/.config/waybar"
copy_cfg "$ROOT/config/hyprlock/hyprlock.conf" "$HOME/.config/hypr/hyprlock.conf"
copy_cfg "$ROOT/config/hypridle/hypridle.conf" "$HOME/.config/hypr/hypridle.conf"
copy_cfg "$ROOT/config/mako"     "$HOME/.config/mako"
copy_cfg "$ROOT/config/fuzzel"   "$HOME/.config/fuzzel"
copy_cfg "$ROOT/config/gtk"      "$HOME/.config/gtk-3.0"
copy_cfg "$ROOT/config/gtk-4.0"  "$HOME/.config/gtk-4.0"
copy_cfg "$ROOT/config/qt"       "$HOME/.config/qt6ct"

run mkdir -p "$HOME/.config/hypr/assets" "$HOME/.config/hyprglass/docs"
run cp -a "$ROOT/assets/wallpapers" "$HOME/.config/hypr/assets/"
run cp -a "$ROOT/docs/shortcuts.md" "$HOME/.config/hyprglass/docs/shortcuts.md"
write_installed_hyprpaper_config
write_source_root
configure_path
configure_desktop_theme
ensure_current_shell_command
verify_installed_configs

ensure_executable_bits

# Systemd services
enable_service_if_present() {
  local svc="$1"
  if systemctl list-unit-files "$svc" >/dev/null 2>&1; then
    sudo systemctl enable --now "$svc"
  else
    echo "Skipping missing service: $svc"
  fi
}

if command -v systemctl >/dev/null 2>&1 && [[ $DRY -eq 0 && $UPDATE -eq 0 && "${HYPRGLASS_SKIP_SERVICES:-0}" != 1 ]]; then
  if ask "Enable laptop/network services now?"; then
    for svc in NetworkManager.service bluetooth.service ModemManager.service power-profiles-daemon.service; do
      enable_service_if_present "$svc"
    done
  fi
fi

# Reload Hyprland if running
if [[ $DRY -eq 0 && -n "${HYPRLAND_INSTANCE_SIGNATURE:-}" ]]; then
  echo "Hyprland session detected — reloading config..."
  run hyprctl reload
  echo "Config reloaded."
  restart_session_components
fi

# Summary
if [[ $DRY -eq 1 ]]; then
  cat <<MSG

Hyprglass dry run complete.
No files, packages, services, or configs were changed.
Backups would be written under: $backup
MSG
elif [[ $UPDATE -eq 1 ]]; then
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
