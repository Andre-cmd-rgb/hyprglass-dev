#!/usr/bin/env bash
set -euo pipefail

cfg=$(mktemp /tmp/hyprglass-power.XXXXXX.ini)
trap 'rm -f "$cfg"' EXIT
cat >"$cfg" <<'FUZZELCFG'
[main]
font=JetBrainsMono Nerd Font:size=13
prompt=⏻
lines=5
width=20
line-height=32

[colors]
background=0f1115ee
text=f2f2ecff
match=8ea8ffff
selection=1a1f2dee
selection-text=f2f2ecff
border=ffffff24

[border]
width=1
radius=16
FUZZELCFG

selected=$(printf '%s\n' \
    '  Lock' \
    '⏸  Suspend' \
    '  Log out' \
    '  Reboot' \
    '⏻  Power off' | fuzzel --config="$cfg" --dmenu) || exit 0

case "$selected" in
    *Lock*)        loginctl lock-session ;;
    *Suspend*)     systemctl suspend ;;
    *"Log out"*)   hyprctl dispatch exit ;;
    *Reboot*)      systemctl reboot ;;
    *"Power off"*) systemctl poweroff ;;
esac
