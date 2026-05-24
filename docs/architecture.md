# Architecture

Hyprglass V1.1 is a repository that installs a disciplined Hyprland workstation layer on top of Arch Linux or CachyOS. It is not an ISO.

The installer is idempotent, backs up existing configs, chooses an Arch or CachyOS package profile when pacman is available, builds one Go binary, and copies configs into `~/.config`.

Config layout mirrors the runtime: Hyprland is modular under `config/hypr/conf.d`; Waybar, Kitty, mako, fuzzel, GTK, and Qt configs live in their own folders. Hyprlock and hypridle source configs live under `config/hyprlock` and `config/hypridle`, but install into `~/.config/hypr/hyprlock.conf` and `~/.config/hypr/hypridle.conf`, which are the default paths used by those tools.

The Go binary owns user-facing checks and terminal-native control surfaces. All external commands go through `internal/command.Runner`, so tests can mock nmcli, bluetoothctl, mmcli, wpctl, hyprctl, pacman, chwd, and CachyOS helper output without requiring hardware.

V1.1 wraps system tools instead of owning system state because Arch/CachyOS already have the right primitives: NetworkManager, BlueZ, ModemManager, PipeWire/WirePlumber, pacman, and Hyprland. Hyprglass should coordinate them cleanly, not replace them.

Future ISO/image work can reuse this repository after the install/check path is proven. ISO generation is intentionally out of scope for V1.1.
