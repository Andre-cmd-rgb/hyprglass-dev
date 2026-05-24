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
version=$(git -C "$ROOT" describe --tags --always --dirty 2>/dev/null || printf 0.1.0)

run bash -n install.sh uninstall.sh scripts/*.sh || fail "shell syntax"
if command -v shellcheck >/dev/null 2>&1; then shellcheck install.sh uninstall.sh scripts/*.sh || fail "shellcheck"; else warn "shellcheck missing"; fi
if [[ -n $(gofmt -l .) ]]; then gofmt -l .; fail "gofmt needed"; fi
run go test ./... || fail "go test"
run go vet ./... || fail "go vet"
run go build -buildvcs=false -ldflags "-s -w -X main.version=$version -X main.sourceRoot=$ROOT" -o "$tmpdir/hyprglass" ./cmd/hyprglass || fail "go build"
run "$tmpdir/hyprglass" --help >/dev/null || fail "hyprglass help"
json=$("$tmpdir/hyprglass" doctor --json) || fail "doctor json command"
if command -v jq >/dev/null 2>&1; then echo "$json" | jq . >/dev/null || fail "invalid doctor JSON"; else warn "jq missing; JSON validation by jq skipped"; fi
while IFS= read -r line; do [[ $line =~ ^source[[:space:]]*=[[:space:]]*(.*)$ ]] || continue; p=${BASH_REMATCH[1]}; p=${p/#~\/.config\/hypr\/}; p=${p/#.config\/hypr\//config/hypr/}; p=${p/#conf.d\//config/hypr/conf.d/}; [[ -f "$p" ]] || fail "missing Hyprland source $p"; done < config/hypr/hyprland.conf
find . -xtype l -print -quit | grep -q . && fail "broken symlink found" || pass "no broken symlinks"
for s in scripts/*.sh scripts/*.py; do [[ -x $s ]] || fail "$s not executable"; done
if [[ $(sort packages/arch-core.txt | uniq -d | wc -l) -ne 0 ]]; then fail "duplicate package in arch-core"; else pass "no duplicate core packages"; fi
if command -v pacman >/dev/null 2>&1; then while read -r p; do [[ -z $p || $p =~ ^# ]] && continue; pacman -Si "$p" >/dev/null || fail "pacman cannot verify $p"; done < packages/arch-core.txt; else warn "pacman missing; Arch package verification skipped on this host"; fi
if [[ -n ${HYPRLAND_INSTANCE_SIGNATURE:-} ]] && command -v hyprctl >/dev/null 2>&1; then hyprctl monitors -j >/dev/null || fail "hyprctl monitors"; else skip "not inside Hyprland; runtime compositor checks skipped"; fi
if grep -RniIE --exclude-dir=.git --exclude-dir=build --exclude-dir=config --exclude-dir=docs --exclude=check.sh "TODO|FIXME|placeholder|maybe|probably|guess|fake|stub|not implemented" .; then fail "weak leftovers found"; else pass "no weak leftovers"; fi
if [[ $FAIL -eq 1 ]]; then echo "FINAL: FAIL"; echo "Rerun: ./scripts/check.sh"; exit 1; fi
if [[ $WARN -eq 1 || $SKIP -gt 0 ]]; then echo "FINAL: WARN"; echo "Skipped checks: $SKIP"; echo "Rerun: ./scripts/check.sh"; echo "Install: ./install.sh --dry-run"; exit 0; fi
echo "FINAL: PASS"; echo "Rerun: ./scripts/check.sh"; echo "Install: ./install.sh --dry-run"
