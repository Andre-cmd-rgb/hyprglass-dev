#!/usr/bin/env bash
set -euo pipefail

# Argument parsing
YES=0 DRY=0 NO_PACKAGES=0 CONFIGS_ONLY=0 UPDATE=0 SKIP_SETUP=0 RATE_MIRRORS=0 AUTO_HARDWARE=0
DISTRO=auto
SETUP_THEME=dark
SETUP_ACCENT=graphite
SETUP_LAYOUT=us
SETUP_VARIANT=
SETUP_SCALE=auto
SETUP_MODEM_APN=
SETUP_MODEM_PIN=
SETUP_ENABLE_SERVICES=1
SETUP_LOGIN_MANAGER=1
SETUP_AUTOLOGIN=0
SETUP_CACHY_CHWD=0
PREFIX="${HYPRGLASS_PREFIX:-${HOME}/.local/bin}"

for a in "$@"; do
  case "$a" in
    --yes)          YES=1 ;;
    --dry-run)      DRY=1 ;;
    --no-packages)  NO_PACKAGES=1 ;;
    --configs-only) CONFIGS_ONLY=1; NO_PACKAGES=1 ;;
    --update)       UPDATE=1; YES=1; NO_PACKAGES=1 ;;
    --skip-setup)   SKIP_SETUP=1 ;;
    --rate-mirrors) RATE_MIRRORS=1 ;;
    --auto-hardware) AUTO_HARDWARE=1; SETUP_CACHY_CHWD=1 ;;
    --login-manager) SETUP_LOGIN_MANAGER=1 ;;
    --no-login-manager) SETUP_LOGIN_MANAGER=0 ;;
    --autologin) SETUP_AUTOLOGIN=1; SETUP_LOGIN_MANAGER=1 ;;
    --distro=*)     DISTRO="${a#*=}" ;;
    --help|-h)
      cat <<HELP
Usage: install.sh [options]

  (no flags)       Full install: packages + binary + configs
  --update         Pull latest changes from git and refresh configs while preserving
                   existing display/monitor rules. Skips packages. Implies --yes.
  --no-packages    Skip pacman; install binary + configs only
  --configs-only   Skip pacman and binary build; configs only
  --yes            Overwrite existing configs without prompting
  --dry-run        Print what would happen; make no changes
  --skip-setup     Use defaults instead of first-setup questions
  --login-manager Enable greetd + tuigreet so boot lands in a Hyprland login. Default.
  --no-login-manager
                  Do not configure greetd.
  --autologin     Add a greetd initial_session for passwordless Hyprland startup.
                  Off by default. Use only on a private machine.
  --distro=auto|arch|cachyos
                  Select package profile. Default auto-detects /etc/os-release.
  --rate-mirrors   On CachyOS, run cachyos-rate-mirrors before package install.
  --auto-hardware  On CachyOS, run chwd -a after package install.
                  This is opt-in because driver changes are OS-level.
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

if [[ ${EUID} -eq 0 && $DRY -ne 1 && "${HYPRGLASS_ALLOW_ROOT:-0}" != 1 ]]; then
  echo "Do not run install.sh as root. It calls sudo only where needed."
  exit 1
fi

ROOT=$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)
run(){ echo "+ $*"; [[ $DRY -eq 1 ]] || "$@"; }

os_release_value() {
  local key="$1"
  [[ -r /etc/os-release ]] || return 1
  awk -F= -v k="$key" '$1==k {gsub(/^"|"$/, "", $2); print tolower($2); exit}' /etc/os-release
}

is_cachyos() {
  local d="${DISTRO,,}" id
  [[ "$d" == cachyos ]] && return 0
  [[ "$d" == arch ]] && return 1
  id=$(os_release_value ID 2>/dev/null || true)
  [[ "$id" == cachyos ]]
}

package_file() {
  local d="${DISTRO,,}"
  case "$d" in
    cachyos) printf '%s\n' "$ROOT/packages/cachyos-core.txt"; return 0 ;;
    arch)    printf '%s\n' "$ROOT/packages/arch-core.txt"; return 0 ;;
    auto)    ;;
    *)       echo "unknown distro profile: $DISTRO" >&2; exit 2 ;;
  esac
  if is_cachyos && [[ -f "$ROOT/packages/cachyos-core.txt" ]]; then
    printf '%s\n' "$ROOT/packages/cachyos-core.txt"
  else
    printf '%s\n' "$ROOT/packages/arch-core.txt"
  fi
}

rank_cachyos_mirrors_if_requested() {
  [[ $RATE_MIRRORS -eq 1 ]] || return 0
  if ! is_cachyos; then
    echo "--rate-mirrors requested, but this is not detected as CachyOS. Skipping."
    return 0
  fi
  if ! command -v cachyos-rate-mirrors >/dev/null 2>&1; then
    echo "cachyos-rate-mirrors is missing. Skipping mirror ranking."
    return 0
  fi
  run sudo cachyos-rate-mirrors
}

ensure_icon_fonts() {
  [[ $CONFIGS_ONLY -eq 0 ]] || return 0
  local missing=()
  if command -v pacman >/dev/null 2>&1; then
    pacman -Q ttf-jetbrains-mono-nerd >/dev/null 2>&1 || missing+=(ttf-jetbrains-mono-nerd)
    pacman -Q ttf-nerd-fonts-symbols-mono >/dev/null 2>&1 || missing+=(ttf-nerd-fonts-symbols-mono)
    pacman -Q fontconfig >/dev/null 2>&1 || missing+=(fontconfig)
    if [[ ${#missing[@]} -gt 0 ]]; then
      echo "Installing missing Hyprglass icon/font packages: ${missing[*]}"
      run sudo pacman -S --needed "${missing[@]}"
    fi
  else
    echo "pacman not found; cannot auto-install icon fonts. Ensure ttf-jetbrains-mono-nerd and ttf-nerd-fonts-symbols-mono are installed."
  fi
  if command -v fc-cache >/dev/null 2>&1; then
    run fc-cache -f
  fi
}

ask(){
  [[ $YES -eq 1 || $DRY -eq 1 ]] && return 0
  read -r -p "$1 [y/N] " ans
  [[ $ans == y || $ans == Y ]]
}

prompt_default() {
  local var_name="$1" label="$2" default="$3" answer
  if [[ $YES -eq 1 || $DRY -eq 1 || $SKIP_SETUP -eq 1 ]]; then
    printf -v "$var_name" '%s' "$default"
    return 0
  fi
  read -r -p "$label [$default]: " answer
  printf -v "$var_name" '%s' "${answer:-$default}"
}

prompt_yes_no_default() {
  local var_name="$1" label="$2" default="$3" answer normalized
  if [[ $YES -eq 1 || $DRY -eq 1 || $SKIP_SETUP -eq 1 ]]; then
    [[ "$default" == y || "$default" == Y || "$default" == yes ]] && printf -v "$var_name" '1' || printf -v "$var_name" '0'
    return 0
  fi
  read -r -p "$label [$default]: " answer
  normalized="${answer:-$default}"
  [[ "$normalized" == y || "$normalized" == Y || "$normalized" == yes || "$normalized" == YES ]] && printf -v "$var_name" '1' || printf -v "$var_name" '0'
}

run_first_setup() {
  [[ $UPDATE -eq 0 ]] || return 0
  if [[ $DRY -eq 1 ]]; then
    echo "+ first setup would ask theme, accent, keyboard layout, display scale, login manager, services, CachyOS hardware, and modem defaults"
    return 0
  fi
  if [[ $YES -eq 1 || $SKIP_SETUP -eq 1 ]]; then
    [[ $AUTO_HARDWARE -eq 1 ]] && SETUP_CACHY_CHWD=1
    return 0
  fi
  cat <<MSG

==> Hyprglass first setup
Choose sane defaults now. You can change user-facing settings later with Super+I.
Driver/service setup happens here so Settings can stay simple.
MSG
  prompt_default SETUP_THEME "Theme mode: dark or light" "dark"
  prompt_default SETUP_ACCENT "Accent: graphite, blue, cyan, green, orange, red, pink, purple" "graphite"
  prompt_default SETUP_LAYOUT "Keyboard layout: us, it, es, gb, de" "us"
  prompt_default SETUP_VARIANT "Keyboard variant, empty is fine" ""
  prompt_default SETUP_SCALE "Display scale: auto, 1.25, 1.5, 1.75, 2" "auto"
  prompt_yes_no_default SETUP_LOGIN_MANAGER "Enable Hyprglass login manager at boot? greetd starts Hyprland after login" "Y"
  if [[ $SETUP_LOGIN_MANAGER -eq 1 ]]; then
    prompt_yes_no_default SETUP_AUTOLOGIN "Passwordless autologin straight into Hyprland? Use only on a private machine" "N"
  fi
  prompt_yes_no_default SETUP_ENABLE_SERVICES "Enable hardware/session services now? NetworkManager, Bluetooth, ModemManager, power profiles" "Y"
  read -r -p "LTE/5G modem APN for autoconnect, empty skips modem setup: " SETUP_MODEM_APN
  if [[ -n "$SETUP_MODEM_APN" ]]; then
    read -r -s -p "SIM PIN for autounlock, empty skips PIN storage: " SETUP_MODEM_PIN
    printf '\n'
  fi
  if is_cachyos; then
    prompt_yes_no_default SETUP_CACHY_CHWD "Run CachyOS hardware auto-configuration after install? Uses sudo chwd -a" "N"
  fi
  [[ $AUTO_HARDWARE -eq 1 ]] && SETUP_CACHY_CHWD=1
}

json_escape() {
  python3 -c 'import json,sys; print(json.dumps(sys.argv[1]))' "$1"
}

write_preferences_json() {
  local dst="$HOME/.config/hyprglass/preferences.json"
  if [[ $DRY -eq 1 ]]; then
    echo "+ write Hyprglass preferences to $dst"
    return 0
  fi
  if [[ $UPDATE -eq 1 && -f "$dst" ]]; then
    echo "Keeping existing Hyprglass preferences at $dst"
    return 0
  fi
  mkdir -p "$(dirname "$dst")"
  python3 - "$dst" "$SETUP_THEME" "$SETUP_ACCENT" "$SETUP_LAYOUT" "$SETUP_VARIANT" "$SETUP_SCALE" "$SETUP_MODEM_APN" <<'PYPREFS'
import json
from pathlib import Path
import sys
path = Path(sys.argv[1])
keys = ["themeMode", "accent", "keyboardLayout", "keyboardVariant", "monitorScale", "modemApn"]
data = dict(zip(keys, sys.argv[2:]))
data["modemPinSet"] = False
path.write_text(json.dumps(data, indent=2) + "\n")
PYPREFS
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
  elif [[ -r "$ROOT/VERSION" ]]; then
    v=$(tr -d '[:space:]' < "$ROOT/VERSION")
  elif [[ -d "$ROOT/.git" ]] && command -v git >/dev/null 2>&1; then
    v=$(git -C "$ROOT" describe --tags --always --dirty 2>/dev/null || true)
  fi
  printf '%s\n' "${v:-dev}"
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

enabled_foreign_display_manager() {
  command -v systemctl >/dev/null 2>&1 || return 1
  local svc
  for svc in sddm.service gdm.service lightdm.service ly.service lxdm.service; do
    if systemctl is-enabled --quiet "$svc" 2>/dev/null; then
      printf '%s\n' "$svc"
      return 0
    fi
  done
  return 1
}

configure_login_manager() {
  [[ $UPDATE -eq 0 ]] || return 0
  [[ $CONFIGS_ONLY -eq 0 ]] || return 0
  [[ $SETUP_LOGIN_MANAGER -eq 1 ]] || return 0
  [[ "${HYPRGLASS_SKIP_SERVICES:-0}" != 1 ]] || return 0

  local user_name existing tmp
  if [[ $DRY -eq 1 ]]; then
    echo "+ configure greetd + tuigreet so boot opens a Hyprland login session"
    [[ $SETUP_AUTOLOGIN -eq 1 ]] && echo "+ add greetd initial_session autologin for the installing user"
    return 0
  fi

  user_name=$(id -un 2>/dev/null || printf '%s' "${USER:-}")
  if [[ -z "$user_name" || "$user_name" == root ]]; then
    echo "Skipping greetd setup: could not determine a non-root target user."
    return 0
  fi

  if ! command -v systemctl >/dev/null 2>&1 || ! command -v sudo >/dev/null 2>&1; then
    echo "Skipping greetd setup: systemctl or sudo unavailable."
    return 0
  fi

  existing=$(enabled_foreign_display_manager || true)
  if [[ -n "$existing" && "${HYPRGLASS_FORCE_GREETD:-0}" != 1 ]]; then
    echo "Skipping greetd setup: $existing is already enabled. Set HYPRGLASS_FORCE_GREETD=1 to replace it."
    return 0
  fi

  if ! command -v tuigreet >/dev/null 2>&1 && ! [[ -x /usr/bin/tuigreet ]]; then
    echo "Skipping greetd setup: tuigreet is not installed. Re-run without --no-packages or install greetd-tuigreet."
    return 0
  fi

  tmp=$(mktemp "${TMPDIR:-/tmp}/hyprglass-greetd.XXXXXX")
  {
    echo '# Hyprglass greetd config.'
    echo '# Managed by install.sh. Backups are kept beside /etc/greetd/config.toml.'
    echo '[terminal]'
    echo 'vt = 1'
    echo
    if [[ $SETUP_AUTOLOGIN -eq 1 ]]; then
      echo '[initial_session]'
      echo 'command = "Hyprland"'
      printf 'user = "%s"\n' "$user_name"
      echo
    fi
    echo '[default_session]'
    echo 'command = "tuigreet --cmd Hyprland"'
    echo 'user = "greeter"'
  } >"$tmp"

  sudo install -d -m 0755 /etc/greetd
  if [[ -f /etc/greetd/config.toml ]]; then
    sudo cp -a /etc/greetd/config.toml "/etc/greetd/config.toml.hyprglass-backup-$(date +%Y%m%d-%H%M%S)"
  fi
  sudo install -m 0644 "$tmp" /etc/greetd/config.toml
  rm -f "$tmp"
  sudo systemctl enable greetd.service
  echo "greetd enabled. On next boot, tuigreet will start Hyprland after login."
}

configure_desktop_theme() {
  if [[ $DRY -eq 1 ]]; then
    echo "+ set GTK/libadwaita dark preference when gsettings is available"
    return 0
  fi
  if command -v gsettings >/dev/null 2>&1; then
    if [[ "$SETUP_THEME" == light ]]; then
      gsettings set org.gnome.desktop.interface color-scheme prefer-light 2>/dev/null || true
      gsettings set org.gnome.desktop.interface gtk-theme Adwaita 2>/dev/null || true
    else
      gsettings set org.gnome.desktop.interface color-scheme prefer-dark 2>/dev/null || true
      gsettings set org.gnome.desktop.interface gtk-theme Adwaita-dark 2>/dev/null || true
    fi
  fi
}

write_installed_hyprpaper_config() {
  local wallpaper="$HOME/.config/hypr/assets/wallpapers/hyprglass-dusk.png"
  local dst="$HOME/.config/hypr/hyprpaper.conf"
  if [[ $DRY -eq 1 ]]; then
    echo "+ write current hyprpaper wallpaper block with absolute path to $dst"
    return 0
  fi
  mkdir -p "$(dirname "$dst")"
  cat >"$dst" <<EOF
# Hyprglass hyprpaper configuration
# Current hyprpaper syntax: one fallback wallpaper block for every monitor.
wallpaper {
    monitor =
    path = $wallpaper
    fit_mode = cover
}

splash = false
ipc = true
EOF
}

write_installed_hyprlock_config() {
  local wallpaper="$HOME/.config/hypr/assets/wallpapers/hyprglass-dusk.png"
  local dst="$HOME/.config/hypr/hyprlock.conf"
  if [[ $DRY -eq 1 ]]; then
    echo "+ write absolute wallpaper path to $dst"
    return 0
  fi
  [[ -f "$dst" ]] || return 0
  python3 - "$dst" "$wallpaper" <<'PYLOCK'
from pathlib import Path
import sys
path = Path(sys.argv[1])
wallpaper = sys.argv[2]
data = path.read_text()
lines = []
for line in data.splitlines():
    if line.strip().startswith('path') and 'hyprglass-dusk.png' in line:
        indent = line[:len(line)-len(line.lstrip())]
        lines.append(f'{indent}path        = {wallpaper}')
    else:
        lines.append(line)
path.write_text('\n'.join(lines) + '\n')
PYLOCK
}

ensure_current_shell_command() {
  local target="$PREFIX/hyprglass"
  local link="/usr/local/bin/hyprglass"
  local pm_target="$PREFIX/hyprglass-powermenu"
  local pm_link="/usr/local/bin/hyprglass-powermenu"

  [[ $CONFIGS_ONLY -eq 0 ]] || return 0
  [[ -x "$target" || $DRY -eq 1 ]] || return 0
  [[ "${HYPRGLASS_NO_GLOBAL_LINK:-0}" != 1 ]] || return 0

  if [[ $DRY -eq 1 ]]; then
    echo "+ link $link -> $target so Hyprland/Waybar can find hyprglass even if ~/.local/bin is not in the session PATH"
    return 0
  fi

  if [[ -w /usr/local/bin ]]; then
    ln -sf "$target" "$link"
    echo "Linked $link -> $target"
    [[ -x "$pm_target" ]] && ln -sf "$pm_target" "$pm_link" && echo "Linked $pm_link -> $pm_target"
  elif command -v sudo >/dev/null 2>&1; then
    sudo ln -sf "$target" "$link"
    echo "Linked $link -> $target"
    if [[ -x "$pm_target" ]]; then
      sudo ln -sf "$pm_target" "$pm_link"
      echo "Linked $pm_link -> $pm_target"
    fi
  else
    cat <<MSG

Could not create $link because sudo is unavailable.
Future shells are configured, but Hyprland/Waybar may not inherit ~/.local/bin.
If commands do not work from the bar, run:
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
  if command -v hyprctl >/dev/null 2>&1; then
    sleep 0.3
    hyprctl hyprpaper wallpaper ", $HOME/.config/hypr/assets/wallpapers/hyprglass-dusk.png, cover" >/dev/null 2>&1 || true
  fi
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
  [[ -f "$HOME/.config/hypr/conf.d/theme.conf" ]] || { echo "Install failed: missing Hyprglass theme config"; exit 1; }
  grep -Fq "Hyprglass main Hyprland config" "$HOME/.config/hypr/hyprland.conf" || {
    echo "Install failed: ~/.config/hypr/hyprland.conf was not replaced with Hyprglass config"
    exit 1
  }
}

ensure_executable_bits
run_first_setup

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
    pkg_file=$(package_file)
    echo "Package profile: ${pkg_file#$ROOT/}"
    [[ -f "$pkg_file" ]] || { echo "Missing package profile: $pkg_file"; exit 1; }
    rank_cachyos_mirrors_if_requested
    mapfile -t pkgs < <(grep -vE '^\s*(#|$)' "$pkg_file" | sort -u)
    if [[ ${#pkgs[@]} -eq 0 ]]; then
      echo "No packages listed in ${pkg_file#$ROOT/}; skipping package install."
    else
      run sudo pacman -S --needed "${pkgs[@]}"
    fi
  else
    echo "pacman not found; skipping package install."
  fi
fi

# Icon fonts are required by Waybar. Updates install these even though normal package install is skipped,
# because old Hyprglass installs lacked the dedicated Symbols Nerd Font fallback.
if [[ $CONFIGS_ONLY -eq 0 && ( $NO_PACKAGES -eq 0 || $UPDATE -eq 1 ) && "${HYPRGLASS_SKIP_ICON_FONT_INSTALL:-0}" != 1 ]]; then
  ensure_icon_fonts
fi

# Build binary
run mkdir -p "$PREFIX"
if [[ $CONFIGS_ONLY -eq 0 ]]; then
  VERSION=$(build_version)
  [[ $DRY -eq 1 ]] || ensure_go_cache
  run go build -buildvcs=false -ldflags "-s -w -X main.version=$VERSION -X main.sourceRoot=$ROOT" -o "$PREFIX/hyprglass" "$ROOT/cmd/hyprglass"
  run install -m 0755 "$ROOT/scripts/hyprglass-powermenu.sh" "$PREFIX/hyprglass-powermenu"
fi

# Back up + copy configs
backup="$HOME/.config/hyprglass-backups/$(date +%Y%m%d-%H%M%S)"
run mkdir -p "$backup"
PRESERVED_DISPLAY_CONFIG=0
PRESERVED_DISPLAY_TMP=""
DISPLAY_CONFIG="$HOME/.config/hypr/conf.d/monitors.conf"

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

preserve_display_config_before_copy() {
  [[ -f "$DISPLAY_CONFIG" ]] || return 0
  if [[ $DRY -eq 1 ]]; then
    echo "+ preserve existing display config at $DISPLAY_CONFIG"
    PRESERVED_DISPLAY_CONFIG=1
    return 0
  fi
  PRESERVED_DISPLAY_TMP=$(mktemp "${TMPDIR:-/tmp}/hyprglass-monitors.XXXXXX")
  cp -a "$DISPLAY_CONFIG" "$PRESERVED_DISPLAY_TMP"
  PRESERVED_DISPLAY_CONFIG=1
}

restore_display_config_after_copy() {
  [[ $PRESERVED_DISPLAY_CONFIG -eq 1 ]] || return 0
  if [[ $DRY -eq 1 ]]; then
    echo "+ restore preserved display config after copying Hyprglass configs"
    return 0
  fi
  [[ -n "$PRESERVED_DISPLAY_TMP" && -f "$PRESERVED_DISPLAY_TMP" ]] || return 0
  mkdir -p "$(dirname "$DISPLAY_CONFIG")"
  cp -a "$PRESERVED_DISPLAY_TMP" "$DISPLAY_CONFIG"
  rm -f "$PRESERVED_DISPLAY_TMP"
  echo "Preserved existing display config: $DISPLAY_CONFIG"
}

preserve_display_config_before_copy
copy_cfg "$ROOT/config/hypr"     "$HOME/.config/hypr"
restore_display_config_after_copy
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
write_installed_hyprlock_config
write_preferences_json
if [[ $DRY -eq 0 && $CONFIGS_ONLY -eq 0 && -x "$PREFIX/hyprglass" ]]; then
  apply_args=(settings apply --no-reload)
  if [[ $UPDATE -eq 0 && $PRESERVED_DISPLAY_CONFIG -eq 0 ]]; then
    apply_args+=(--with-display)
  fi
  "$PREFIX/hyprglass" "${apply_args[@]}" >/dev/null || true
fi
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

if command -v systemctl >/dev/null 2>&1 && [[ $DRY -eq 0 && $UPDATE -eq 0 && "${HYPRGLASS_SKIP_SERVICES:-0}" != 1 && $SETUP_ENABLE_SERVICES -eq 1 ]]; then
  for svc in NetworkManager.service bluetooth.service ModemManager.service power-profiles-daemon.service; do
    enable_service_if_present "$svc"
  done
fi

configure_login_manager

if [[ $DRY -eq 0 && $UPDATE -eq 0 && -n "$SETUP_MODEM_APN" && -x "$ROOT/scripts/hyprglass-modem-autounlock-install.sh" ]]; then
  args=("$ROOT/scripts/hyprglass-modem-autounlock-install.sh" --apn "$SETUP_MODEM_APN")
  [[ -n "$SETUP_MODEM_PIN" ]] && args+=(--pin "$SETUP_MODEM_PIN")
  sudo bash "${args[@]}"
fi

if [[ $DRY -eq 0 && $UPDATE -eq 0 && $SETUP_CACHY_CHWD -eq 1 ]]; then
  if is_cachyos && command -v chwd >/dev/null 2>&1; then
    sudo chwd -a
  else
    echo "Skipping CachyOS hardware auto-configuration: chwd is unavailable or this is not CachyOS."
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
Configs refreshed. Existing display/monitor rules were preserved. Previous configs backed up to:
  $backup

If Hyprland was running, config was reloaded automatically.
If not, reload manually with: hyprctl reload
Run: hyprglass doctor
MSG
else
  cat <<MSG

Hyprglass install complete.
Boot login: greetd/tuigreet is configured when enabled in first setup.
Run: hyprglass doctor
Open Hyprglass Settings with Super+I or Super+comma inside Hyprland.
Backups: $backup
MSG
fi
