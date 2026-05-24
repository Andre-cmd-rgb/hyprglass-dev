#!/usr/bin/env bash
set -euo pipefail
ROOT=$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)
cd "$ROOT"
FAIL=0; WARN=0; SKIP=0
pass(){ echo "PASS $*"; }
warn(){ echo "WARN $*"; WARN=1; }
fail(){ echo "FAIL $*"; FAIL=1; }
skip(){ echo "SKIP $*"; SKIP=$((SKIP+1)); }
run(){ echo "+ $*"; "$@"; }

if [[ -z "${GOCACHE:-}" ]]; then
  export GOCACHE="${XDG_CACHE_HOME:-$HOME/.cache}/go-build"
fi
if ! mkdir -p "$GOCACHE" 2>/dev/null || [[ ! -w "$GOCACHE" ]]; then
  export GOCACHE="${TMPDIR:-/tmp}/hyprglass-go-cache"
  mkdir -p "$GOCACHE"
fi

tmpdir=$(mktemp -d "${TMPDIR:-/tmp}/hyprglass-check.XXXXXX")
trap 'rm -rf "$tmpdir"' EXIT
version=$(tr -d "[:space:]" < "$ROOT/VERSION" 2>/dev/null || git -C "$ROOT" describe --tags --always --dirty 2>/dev/null || printf dev)

run bash -n install.sh uninstall.sh scripts/*.sh || fail "shell syntax"
[[ -r VERSION ]] || fail "VERSION file missing"
[[ $version =~ ^[0-9]+\.[0-9]+\.[0-9]+(-[A-Za-z0-9._-]+)?$|^dev$ ]] || fail "VERSION must be semver or dev"
if command -v shellcheck >/dev/null 2>&1; then shellcheck install.sh uninstall.sh scripts/*.sh || fail "shellcheck"; else warn "shellcheck missing"; fi
if [[ -n $(gofmt -l .) ]]; then gofmt -l .; fail "gofmt needed"; fi
run go test ./... || fail "go test"
run go vet ./... || fail "go vet"
run go build -buildvcs=false -ldflags "-s -w -X main.version=$version -X main.sourceRoot=$ROOT" -o "$tmpdir/hyprglass" ./cmd/hyprglass || fail "go build"
run "$tmpdir/hyprglass" --help >/dev/null || fail "hyprglass help"
run "$tmpdir/hyprglass" system --json >/dev/null || fail "hyprglass system json"
printf 'q\n' | "$tmpdir/hyprglass" power >/dev/null || fail "hyprglass power menu"
printf 'q\n' | "$tmpdir/hyprglass" settings >/dev/null || fail "hyprglass settings close path"
printf 'q\n' | "$tmpdir/hyprglass" laptop >/dev/null || fail "hyprglass laptop menu"
json=$("$tmpdir/hyprglass" doctor --json) || fail "doctor json command"
if command -v jq >/dev/null 2>&1; then echo "$json" | jq . >/dev/null || fail "invalid doctor JSON"; else warn "jq missing; JSON validation by jq skipped"; fi
if command -v jq >/dev/null 2>&1; then jq -e . config/waybar/config.jsonc >/dev/null || fail "invalid Waybar JSONC"; else python3 -m json.tool config/waybar/config.jsonc >/dev/null || fail "invalid Waybar JSONC"; fi
while IFS= read -r line; do [[ $line =~ ^source[[:space:]]*=[[:space:]]*(.*)$ ]] || continue; p=${BASH_REMATCH[1]}; p=${p/#~\/.config\/hypr\/}; p=${p/#.config\/hypr\//config/hypr/}; p=${p/#conf.d\//config/hypr/conf.d/}; [[ -f "$p" ]] || fail "missing Hyprland source $p"; done < config/hypr/hyprland.conf
if grep -RniE '^[[:space:]]*windowrule[[:space:]]*=[[:space:]]*(float|tile|fullscreen|maximize|center|pseudo|pin|no_initial_focus|stay_focused|no_blur|no_shadow|no_anim|opaque|force_rgbx)([[:space:]]*,|[[:space:]]*$)|^[[:space:]]*windowrule[[:space:]]*=.*,([[:space:]]*)(float|tile|fullscreen|maximize|center|pseudo|pin|no_initial_focus|stay_focused|no_blur|no_shadow|no_anim|opaque|force_rgbx)([[:space:]]*,|[[:space:]]*$)' config/hypr; then fail "windowrule boolean effects need explicit values"; else pass "windowrule boolean effects have explicit values"; fi
if grep -RniE '^[[:space:]]*layerrule[[:space:]]*=[[:space:]]*(blur|blur_popups|no_anim|dim_around|no_screen_share)([[:space:]]*,|[[:space:]]*$)|^[[:space:]]*layerrule[[:space:]]*=.*,([[:space:]]*)(blur|blur_popups|no_anim|dim_around|no_screen_share)([[:space:]]*,|[[:space:]]*$)' config/hypr; then fail "layerrule boolean effects need explicit values"; else pass "layerrule boolean effects have explicit values"; fi
for svc in hyprpaper waybar mako hypridle; do grep -Eq "^exec-once[[:space:]]*=[[:space:]]*$svc([[:space:]]|\$)" config/hypr/conf.d/autostart.conf || fail "missing autostart for $svc"; done
grep -Fq "hyprglass-dusk.png" config/hypr/hyprpaper.conf || fail "hyprpaper wallpaper path missing"
grep -Fq "wallpaper {" config/hypr/hyprpaper.conf || fail "hyprpaper config must use current wallpaper block syntax"
grep -Fq "hyprglass managed display" config/hypr/conf.d/monitors.conf || fail "monitor config must use a Hyprglass managed display block"
grep -Fq "ApplyDisplayAndInput" internal/appsettings/settings.go || fail "Settings display page must call display-scoped apply logic"
grep -Fq "PRESERVED_DISPLAY_CONFIG" install.sh || fail "installer must preserve existing display config during update/install refresh"
grep -Fq "configure_login_manager" install.sh || fail "installer must configure boot login manager"
grep -Fq "tuigreet --cmd Hyprland" install.sh || fail "greetd config must start Hyprland through tuigreet"
if grep -RniE '^[[:space:]]*(preload[[:space:]]*=|wallpaper[[:space:]]*=)' config/hypr/hyprpaper.conf; then fail "hyprpaper config uses removed legacy syntax"; fi
[[ -f assets/wallpapers/hyprglass-dusk.png ]] || fail "wallpaper asset missing"
[[ -f config/waybar/config.jsonc && -f config/waybar/style.css ]] || fail "waybar config/style missing"
grep -Fq "setsid -f kitty" config/waybar/config.jsonc || fail "Waybar click commands must detach Settings from Waybar"
if grep -n 'pkill.*-x.*waybar' internal/appsettings/settings.go | grep -v SIGUSR2; then fail "Settings must not hard-kill Waybar; it can kill the bar when launched from the bar"; fi
grep -Fq 'SIGUSR2' internal/appsettings/settings.go || fail "Settings must reload Waybar with SIGUSR2"
if grep -RniE '#[0-9a-fA-F]{8}([[:space:];]|$)' config/waybar/style.css; then fail "Waybar GTK CSS must not use 8-digit hex colors"; fi
grep -Fq 'Symbols Nerd Font Mono' config/waybar/style.css || fail "Waybar CSS must include Symbols Nerd Font Mono fallback"
grep -Fq 'JetBrainsMonoNL Nerd Font' config/waybar/style.css || fail "Waybar CSS must include JetBrainsMonoNL Nerd Font fallback"
grep -Fq 'ttf-nerd-fonts-symbols-mono' packages/arch-core.txt || fail "Arch package profile missing Symbols Nerd Font package"
grep -Fq 'ttf-nerd-fonts-symbols-mono' packages/cachyos-core.txt || fail "CachyOS package profile missing Symbols Nerd Font package"
grep -Fq 'fontconfig' packages/arch-core.txt || fail "Arch package profile missing fontconfig"
grep -Fq 'fontconfig' packages/cachyos-core.txt || fail "CachyOS package profile missing fontconfig"
grep -Fq 'greetd' packages/arch-core.txt || fail "Arch package profile missing greetd"
grep -Fq 'greetd-tuigreet' packages/arch-core.txt || fail "Arch package profile missing greetd-tuigreet"
grep -Fq 'greetd' packages/cachyos-core.txt || fail "CachyOS package profile missing greetd"
grep -Fq 'greetd-tuigreet' packages/cachyos-core.txt || fail "CachyOS package profile missing greetd-tuigreet"
grep -Fq 'ensure_icon_fonts' install.sh || fail "installer must repair/install icon fonts"
grep -Fq 'hyprglass icons [status|repair]' cmd/hyprglass/main.go || fail "CLI help missing icon repair command"
pass "wallpaper, icon font, and top bar config chain present"
install_home="$tmpdir/home"
mkdir -p "$install_home/.config/hypr"
printf '# autogenerated by Hyprland\n' >"$install_home/.config/hypr/hyprland.conf"
env HOME="$install_home" HYPRGLASS_SKIP_SERVICES=1 HYPRGLASS_ALLOW_ROOT=1 "$ROOT/install.sh" --configs-only --yes >/dev/null || fail "temp configs-only install"
grep -Fq "Hyprglass main Hyprland config" "$install_home/.config/hypr/hyprland.conf" || fail "installer did not replace generated hyprland.conf"
[[ -f "$install_home/.config/hypr/hyprpaper.conf" ]] || fail "installer did not copy hyprpaper.conf"
grep -Fq "$install_home/.config/hypr/assets/wallpapers/hyprglass-dusk.png" "$install_home/.config/hypr/hyprpaper.conf" || fail "hyprpaper config does not use absolute wallpaper path"
grep -Fq "wallpaper {" "$install_home/.config/hypr/hyprpaper.conf" || fail "installed hyprpaper config does not use wallpaper block syntax"
if grep -RniE '^[[:space:]]*(preload[[:space:]]*=|wallpaper[[:space:]]*=)' "$install_home/.config/hypr/hyprpaper.conf"; then fail "installed hyprpaper config uses removed legacy syntax"; fi
[[ -f "$install_home/.config/hypr/assets/wallpapers/hyprglass-dusk.png" ]] || fail "installer did not copy wallpaper asset"
[[ -f "$install_home/.config/waybar/config.jsonc" && -f "$install_home/.config/waybar/style.css" ]] || fail "installer did not copy waybar config/style"
[[ -f "$install_home/.config/gtk-4.0/settings.ini" ]] || fail "installer did not copy GTK4 settings"
HOME="$install_home" "$tmpdir/hyprglass" wallpaper apply >/dev/null || fail "wallpaper apply command"
[[ -f "$install_home/.config/hypr/assets/wallpapers/hyprglass-dusk.png" ]] || fail "wallpaper apply did not install wallpaper"
pass "temp install replaces generated config and copies wallpaper/top bar files"

preserve_home="$tmpdir/home-preserve"
mkdir -p "$preserve_home/.config/hypr/conf.d"
cat >"$preserve_home/.config/hypr/conf.d/monitors.conf" <<'MONITOR_EOF'
# user display layout must survive installer refresh
monitor = eDP-1, 3840x2400@60, 0x0, 1.75
MONITOR_EOF
env HOME="$preserve_home" HYPRGLASS_SKIP_SERVICES=1 HYPRGLASS_ALLOW_ROOT=1 "$ROOT/install.sh" --configs-only --yes >/dev/null || fail "temp configs-only install with existing display config"
grep -Fq "monitor = eDP-1, 3840x2400@60, 0x0, 1.75" "$preserve_home/.config/hypr/conf.d/monitors.conf" || fail "installer overwrote existing display config"
if grep -Fq "monitor = , preferred, auto, auto" "$preserve_home/.config/hypr/conf.d/monitors.conf"; then fail "installer injected generic display rule into preserved config"; fi
pass "installer preserves existing display config"

managed_home="$tmpdir/home-managed-scale"
mkdir -p "$managed_home/.config/hypr/conf.d" "$managed_home/.config/hyprglass"
cat >"$managed_home/.config/hypr/conf.d/monitors.conf" <<'MANAGED_SCALE_EOF'
# >>> hyprglass managed display >>>
monitor = eDP-1, 3840x2400@60, 0x0, 1.75
# <<< hyprglass managed display <<<
MANAGED_SCALE_EOF
cat >"$managed_home/.config/hyprglass/preferences.json" <<'PREF_SCALE_EOF'
{
  "themeMode": "dark",
  "accent": "graphite",
  "keyboardLayout": "us",
  "keyboardVariant": "",
  "monitorScale": "2",
  "modemApn": "",
  "modemPinSet": false
}
PREF_SCALE_EOF
HOME="$managed_home" "$tmpdir/hyprglass" settings apply --with-display --no-reload >/dev/null || fail "display managed scale apply command"
grep -Fq "monitor = eDP-1, 3840x2400@60, 0x0, 2" "$managed_home/.config/hypr/conf.d/monitors.conf" || fail "display scale apply did not preserve custom resolution"
if grep -Fq "monitor = , preferred, auto" "$managed_home/.config/hypr/conf.d/monitors.conf"; then fail "display scale apply replaced custom display rule"; fi
pass "display scale updates preserve custom resolution and placement"

find . -xtype l -print -quit | grep -q . && fail "broken symlink found" || pass "no broken symlinks"
for s in scripts/*.sh scripts/*.py; do [[ -x $s ]] || fail "$s not executable"; done
for pkg_file in packages/arch-core.txt packages/cachyos-core.txt; do
  [[ -f "$pkg_file" ]] || { fail "missing package profile $pkg_file"; continue; }
  if [[ $(grep -vE '^\s*(#|$)' "$pkg_file" | sort | uniq -d | wc -l) -ne 0 ]]; then
    fail "duplicate package in $pkg_file"
  else
    pass "no duplicate packages in $pkg_file"
  fi
done
if command -v pacman >/dev/null 2>&1; then
  while read -r p; do [[ -z $p || $p =~ ^# ]] && continue; pacman -Si "$p" >/dev/null || fail "pacman cannot verify Arch package $p"; done < packages/arch-core.txt
  if grep -qi '^ID=cachyos' /etc/os-release 2>/dev/null; then
    while read -r p; do [[ -z $p || $p =~ ^# ]] && continue; pacman -Si "$p" >/dev/null || fail "pacman cannot verify CachyOS package $p"; done < packages/cachyos-core.txt
  fi
else
  warn "pacman missing; Arch/CachyOS package verification skipped on this host"
fi
if [[ -n ${HYPRLAND_INSTANCE_SIGNATURE:-} ]] && command -v hyprctl >/dev/null 2>&1; then hyprctl monitors -j >/dev/null || fail "hyprctl monitors"; else skip "not inside Hyprland; runtime compositor checks skipped"; fi
if grep -RniIE --exclude-dir=.git --exclude-dir=build --exclude-dir=config --exclude-dir=docs --exclude=check.sh "TODO|FIXME|placeholder|maybe|probably|guess|fake|stub|not implemented" .; then fail "weak leftovers found"; else pass "no weak leftovers"; fi
if [[ $FAIL -eq 1 ]]; then echo "FINAL: FAIL"; echo "Rerun: ./scripts/check.sh"; exit 1; fi
if [[ $WARN -eq 1 || $SKIP -gt 0 ]]; then echo "FINAL: WARN"; echo "Skipped checks: $SKIP"; echo "Rerun: ./scripts/check.sh"; echo "Install: ./install.sh --dry-run"; exit 0; fi
echo "FINAL: PASS"; echo "Rerun: ./scripts/check.sh"; echo "Install: ./install.sh --dry-run"
