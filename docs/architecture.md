# Architecture

Hyprglass V0 is a repository that installs a disciplined Hyprland workstation layer on top of Arch. It is not an ISO.

The installer is idempotent, backs up existing configs, installs official Arch packages when pacman is available, builds one Go binary, and copies configs into `~/.config`.

Config layout mirrors the runtime: Hyprland is modular under `config/hypr/conf.d`; Waybar, Kitty, hyprlock, hypridle, mako, fuzzel, GTK, and Qt configs live in their own folders.

The Go binary owns user-facing checks and terminal-native control surfaces. All external commands go through `internal/command.Runner`, so tests can mock nmcli, bluetoothctl, mmcli, wpctl, and hyprctl output without requiring hardware.

V0 wraps system tools instead of owning system state because Arch already has the right primitives: NetworkManager, BlueZ, ModemManager, PipeWire/WirePlumber, and Hyprland. Hyprglass should coordinate them cleanly, not replace them.

Future ISO/image work can reuse this repository after the install/check path is proven. ISO generation is intentionally out of scope for V0.
