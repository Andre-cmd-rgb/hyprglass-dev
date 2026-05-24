# Hyprglass V0 research

Verification date: 2026-05-24

Scope: packages, commands, services, config syntax, and public documentation were checked from available web results and local commands. This build host is not Arch and not running Hyprland, so pacman and live compositor validation are marked as skipped by `scripts/check.sh` here.

| Item | Type | Verified source | Decision |
|---|---|---|---|
| hyprland | compositor/package | Hyprland official installation docs; Arch Wiki as Arch integration support | core |
| xdg-desktop-portal-hyprland | portal | Hyprland official XDPH docs | core |
| waybar | bar | Hyprland status bar docs; Waybar upstream | core |
| hyprland/workspaces | Waybar module | Hyprland status bar docs says Waybar supports Hyprland and to use `hyprland/workspaces` | core |
| hyprlock, hypridle, hyprpaper | Hypr ecosystem tools | Hyprland installation docs list ecosystem packages | core |
| kitty | terminal | Kitty official config/man docs expected on target system | core |
| fuzzel | launcher | upstream package/config docs expected on target system | core |
| mako | notifications | upstream package/config docs expected on target system | core |
| NetworkManager / nmcli | network backend | NetworkManager docs define nmcli connection profile model | core |
| BlueZ / bluetoothctl | Bluetooth backend | BlueZ tool behavior verified by command-oriented interface and package split `bluez-utils` | core |
| ModemManager / mmcli | LTE backend | ModemManager mmcli man page documents modem listing/control and simple connect | core |
| PipeWire / WirePlumber / wpctl | audio backend | Arch package/service convention, local command detection in doctor | core |
| Bubble Tea / Lip Gloss | Go TUI libraries | Not used in V0 to avoid external module fragility in offline builds; stdlib terminal output is used | deferred |

## Commands verified or checked by doctor

`hyprctl`, `kitty`, `waybar`, `hyprlock`, `hypridle`, `hyprpaper`, `fuzzel`, `mako`, `nmcli`, `bluetoothctl`, `mmcli`, `wpctl`, `grim`, `slurp`, `wl-copy`, `systemctl`, `loginctl`, `jq`, and `go` are checked by `hyprglass doctor`.

## Services

System services: `NetworkManager.service`, `bluetooth.service`, `ModemManager.service`.
User services checked operationally by target commands: PipeWire/WirePlumber and XDG portals. The installer offers enabling only the three system services and does not touch partitions, packages outside Arch repos, or user data.

## Decisions made

- Use `hyprpaper` instead of `swww` for V0 because it is part of the Hypr ecosystem and keeps wallpaper behavior simple.
- Use Waybar `pulseaudio` module because it is widely supported while PipeWire exposes pulse-compatible controls through `pipewire-pulse`.
- Use a custom Waybar Bluetooth module because native Bluetooth support varies by build and configuration.
- Keep LTE repair manual/confirmed because restarting ModemManager or NetworkManager can kill the active connection.
- Keep display changes read-only in V0 because blind monitor writes can black-screen users.
- Do not use Bubble Tea/Lip Gloss in this artifact because the build environment may not fetch modules; the command runner remains mockable and the CLI is usable.

## Uncertain or runtime-only items

- Live Hyprland reload, monitor state, blur rendering, portal behavior, Bluetooth adapter state, LTE modem state, and battery module behavior require a real session/hardware.
- Arch package availability is fully validated only on Arch with `pacman -Si`; this host does not provide pacman.
