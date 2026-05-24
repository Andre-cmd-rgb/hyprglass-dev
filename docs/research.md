# Developer notes

Implementation decisions that are not obvious from reading the code.

## Hyprland monitor rules

`monitor = , preferred, auto, <scale>` still works on current Hyprland; `auto` as the scale value lets the compositor choose a PPI-driven default. Hyprglass starts with `auto` and lets Settings generate fixed scales such as 1.5, 1.75, or 2.

## Hyprland window/layer rules

Hyprland 0.53+ rewrote the rule syntax. Hyprglass uses the current named-matcher style: `windowrule = float on, match:class ...` and `layerrule = blur on, match:namespace ...`. The old `windowrulev2` keyword was removed.

## Hyprpaper wallpaper syntax

Hyprpaper reads `~/.config/hypr/hyprpaper.conf`. Current hyprpaper sets wallpapers through `wallpaper { ... }` blocks where an empty `monitor =` field acts as the fallback target for all monitors. Hyprglass writes this block with an absolute path and also sends an IPC wallpaper command after restarting hyprpaper.

## Modem/SIM handling

ModemManager can unlock a SIM with `mmcli -i <SIM_INDEX> --pin=<PIN>`. NetworkManager GSM profiles support `gsm.pin`, `gsm.apn`, autoconnect, and autoconnect priority. Hyprglass stores PIN material only in `/etc/hyprglass/modem.env` with root-only permissions when the user enables modem autounlock.

## Waybar CSS

Waybar is GTK-based; the bar CSS avoids 8-digit hex colors (`#RRGGBBAA`) and uses `rgba(...)` for alpha transparency. GTK's CSS parser rejects 8-digit hex with "Junk at end of value for background". All Waybar colors go through `cssColor()` in `internal/prefs/prefs.go`.

## CachyOS

CachyOS is an Arch-compatible target, not a migration mode. Hyprglass uses a separate CachyOS package profile and exposes `cachyos-rate-mirrors`/`chwd` actions only when those tools are present. CachyOS repository migration and kernel selection are OS-level concerns that stay outside the installer.
