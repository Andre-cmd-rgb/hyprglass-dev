# Hyprglass V1.1 research notes

- Hyprland monitor rules still support `monitor = , preferred, auto, <scale>`; current docs also support `auto` scale for PPI-driven defaults. Hyprglass now starts with `auto` and lets Settings generate fixed scales such as 1.5, 1.75, or 2.
- Hyprland 0.53+ rewrote window/layer rules. Hyprglass keeps the newer named matcher style already used by the repo: `windowrule = float on, match:class ...` and `layerrule = blur on, match:namespace ...`.
- Hyprpaper reads `~/.config/hypr/hyprpaper.conf`. Current hyprpaper sets wallpapers through `wallpaper { ... }` blocks where an empty `monitor =` acts as the fallback target. Hyprglass writes this block with an absolute path and also sends an IPC wallpaper command after restarting hyprpaper.
- ModemManager can unlock a SIM with `mmcli -i <SIM_INDEX> --pin=<PIN>`. NetworkManager GSM profiles support `gsm.pin`, `gsm.apn`, autoconnect, and autoconnect priority. Hyprglass stores PIN material only in `/etc/hyprglass/modem.env` with root-only permissions when the user enables modem autounlock.

- Waybar is GTK-based, so the bar CSS avoids 8-digit hex colors and uses `rgba(...)` for alpha transparency. This prevents GTK parser errors such as `Junk at end of value for background`.

- CachyOS is treated as an Arch-compatible target, not as a repository migration mode. Hyprglass uses a separate CachyOS package profile and exposes `cachyos-rate-mirrors`/`chwd` actions only when those tools are available.
- CachyOS repository and kernel migration stay outside the installer because they change pacman repositories and kernel packages. Users should do that through CachyOS itself or explicit OS-level documentation.
