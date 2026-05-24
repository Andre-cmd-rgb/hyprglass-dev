# Architecture

Hyprglass is a repository that installs a disciplined Hyprland workstation layer on top of Arch Linux or CachyOS. It is not an ISO.

The installer is idempotent, backs up existing configs, chooses an Arch or CachyOS package profile when pacman is available, builds one Go binary, and copies configs into `~/.config`.

Config layout mirrors the runtime: Hyprland is modular under `config/hypr/conf.d`; Waybar, Kitty, mako, fuzzel, GTK, and Qt configs live in their own folders. Hyprlock and hypridle source configs live under `config/hyprlock` and `config/hypridle`, but install into `~/.config/hypr/hyprlock.conf` and `~/.config/hypr/hypridle.conf`, which are the default paths used by those tools.

The Go binary owns user-facing checks and terminal-native control surfaces. All external commands go through `internal/command.Runner`, so tests can mock nmcli, bluetoothctl, mmcli, wpctl, hyprctl, pacman, chwd, and CachyOS helper output without requiring hardware.

Hyprglass wraps system tools instead of owning system state because Arch/CachyOS already have the right primitives: NetworkManager, BlueZ, ModemManager, PipeWire/WirePlumber, pacman, and Hyprland. Hyprglass coordinates them cleanly, not replace them.

## Package layout

```
cmd/hyprglass/        — CLI entry point; dispatches to internal packages
internal/
  appsettings/        — Settings TUI menu
  audio/              — PipeWire/WirePlumber status
  bluetooth/          — bluetoothctl wrapper
  command/            — Runner interface + timeout-aware RealRunner + MockRunner
  display/            — hyprctl monitor info
  doctor/             — environment health checks
  fileutil/           — shared file copy helper
  icons/              — Nerd Font icon/font repair
  laptop/             — battery, thermals, power profiles
  lte/                — ModemManager modem status
  platform/           — /etc/os-release detection
  prefs/              — user preferences JSON + config generation
  srcroot/            — source checkout finder (env, ldflags, file, cwd, exe)
  system/             — system menu (pacman, chwd, cachyos-rate-mirrors)
  tui/                — shared terminal UI helpers
  wifi/               — NetworkManager Wi-Fi status
config/               — Hyprland, Waybar, Kitty, mako, fuzzel, GTK, Qt configs
packages/             — Arch and CachyOS package lists
scripts/              — build, check, install helpers
```
