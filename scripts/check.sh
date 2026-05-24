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
run bash -n install.sh uninstall.sh scripts/*.sh || fail "shell syntax"
if command -v shellcheck >/dev/null 2>&1; then shellcheck install.sh uninstall.sh scripts/*.sh || fail "shellcheck"; else warn "shellcheck missing"; fi
if [[ -n $(gofmt -l .) ]]; then gofmt -w .; fail "gofmt changed files; rerun check"; fi
run go test ./... || fail "go test"
run go vet ./... || fail "go vet"
mkdir -p build
run go build -buildvcs=false -o build/hyprglass ./cmd/hyprglass || fail "go build"
run build/hyprglass --help >/dev/null || fail "hyprglass help"
json=$(build/hyprglass doctor --json) || fail "doctor json command"
if command -v jq >/dev/null 2>&1; then echo "$json" | jq . >/dev/null || fail "invalid doctor JSON"; else warn "jq missing; JSON validation by jq skipped"; fi
while IFS= read -r line; do [[ $line =~ ^source[[:space:]]*=[[:space:]]*(.*)$ ]] || continue; p=${BASH_REMATCH[1]}; p=${p/#~\/.config\/hypr\/}; p=${p/#.config\/hypr\//config/hypr/}; p=${p/#conf.d\//config/hypr/conf.d/}; [[ -f "$p" ]] || fail "missing Hyprland source $p"; done < config/hypr/hyprland.conf
find . -xtype l -print -quit | grep -q . && fail "broken symlink found" || pass "no broken symlinks"
for s in scripts/*.sh; do [[ -x $s ]] || fail "$s not executable"; done
if [[ $(sort packages/arch-core.txt | uniq -d | wc -l) -ne 0 ]]; then fail "duplicate package in arch-core"; else pass "no duplicate core packages"; fi
if command -v pacman >/dev/null 2>&1; then while read -r p; do [[ -z $p || $p =~ ^# ]] && continue; pacman -Si "$p" >/dev/null || fail "pacman cannot verify $p"; done < packages/arch-core.txt; else warn "pacman missing; Arch package verification skipped on this host"; fi
if [[ -n ${HYPRLAND_INSTANCE_SIGNATURE:-} ]] && command -v hyprctl >/dev/null 2>&1; then hyprctl monitors -j >/dev/null || fail "hyprctl monitors"; else skip "not inside Hyprland; runtime compositor checks skipped"; fi
if grep -RniIE --exclude-dir=.git --exclude-dir=build --exclude=check.sh "TODO|FIXME|placeholder|maybe|probably|guess|fake|stub|not implemented" .; then fail "weak leftovers found"; else pass "no weak leftovers"; fi
if [[ $FAIL -eq 1 ]]; then echo "FINAL: FAIL"; echo "Rerun: ./scripts/check.sh"; exit 1; fi
if [[ $WARN -eq 1 || $SKIP -gt 0 ]]; then echo "FINAL: WARN"; echo "Skipped checks: $SKIP"; echo "Rerun: ./scripts/check.sh"; echo "Install: ./install.sh --dry-run"; exit 0; fi
echo "FINAL: PASS"; echo "Rerun: ./scripts/check.sh"; echo "Install: ./install.sh --dry-run"
