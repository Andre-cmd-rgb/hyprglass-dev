#!/usr/bin/env bash
set -euo pipefail

ROOT=$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)
DISTRO=${HYPRGLASS_DISTRO:-auto}

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

install_packages() {
  local pkg_file
  pkg_file=$(package_file)
  [[ -f "$pkg_file" ]] || { echo "Missing package profile: $pkg_file" >&2; exit 1; }
  echo "Package profile: ${pkg_file#$ROOT/}"
  mapfile -t pkgs < <(grep -vE '^\s*(#|$)' "$pkg_file" | sort -u)
  if [[ ${#pkgs[@]} -eq 0 ]]; then
    echo "No packages listed in ${pkg_file#$ROOT/}; skipping package install."
    return 0
  fi
  exec sudo pacman -S --needed "${pkgs[@]}"
}

backup_configs() {
  local dst
  dst="$HOME/.config/hyprglass-backups/manual-$(date +%Y%m%d-%H%M%S)"
  mkdir -p "$dst"
  for d in hypr kitty waybar mako fuzzel gtk-3.0 gtk-4.0 qt6ct hyprglass hyprlock hypridle; do
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
  HYPRGLASS_DISTRO=auto|arch|cachyos scripts/install-packages.sh
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
