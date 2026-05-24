# Hyprglass V1.2

Hyprglass is a fast, glassy Wayland desktop built on Hyprland for focused work.

Version is stored in `VERSION` and injected into the Go binary at build time.

It is not a distro, not an ISO, not a GNOME/KDE clone, not a profile-switching rice dump, and not an Electron settings app. V1.2 turns a clean Arch or CachyOS install into a polished laptop-oriented Hyprland workstation with Kitty, Waybar, fuzzel, hyprlock, hypridle, mako, and one central terminal-native Settings app.

## Install

Dry run first:

```sh
./install.sh --dry-run
```

Install on a clean Arch system:

```sh
./install.sh
```

Install on CachyOS:

```sh
./install.sh --distro=cachyos
```

CachyOS with mirror ranking:

```sh
./install.sh --distro=cachyos --rate-mirrors
```

CachyOS with opt-in hardware auto-configuration:

```sh
./install.sh --distro=cachyos --auto-hardware
```

For CachyOS, Hyprglass uses `packages/cachyos-core.txt`, keeps the same clean Hyprland stack, and adds CachyOS support hooks for mirror ranking and hardware detection. It does not convert Arch into CachyOS and it does not silently replace your kernel.

The installer asks first-setup questions for theme, accent, keyboard layout, display scale, hardware/session services, optional LTE/5G modem autoconnect, and optional CachyOS hardware auto-configuration. For non-interactive installs:

```sh
./install.sh --yes
```

After install, open a new terminal or run:

```sh
exec $SHELL -l
```

Skip packages:

```sh
./install.sh --no-packages --yes
```

Update an existing checkout and refresh configs:

```sh
hyprglass update
```

or from the repo:

```sh
./install.sh --update
```

Run checks:

```sh
./scripts/check.sh
```

System/CachyOS tools:

```sh
hyprglass system
hyprglass system --json
hyprglass system rate-mirrors
hyprglass system chwd-list
```

`hyprglass system chwd-auto` exists for CachyOS machines, but it is confirmation-gated because driver changes are real system changes.

Uninstall configs:

```sh
./uninstall.sh
```

## Commands

```sh
hyprglass --help
hyprglass settings
hyprglass settings apply
hyprglass doctor
hyprglass doctor --json
hyprglass repair
hyprglass wifi
hyprglass bluetooth
hyprglass lte
hyprglass audio
hyprglass display
hyprglass laptop
hyprglass power
hyprglass wallpaper apply
hyprglass touchid status
```

## Settings

`hyprglass settings` is the main control surface. The user-facing screen is intentionally simple: appearance, display/keyboard, network/modem, audio, power, and update. Repairs, doctor checks, services, and CachyOS system actions live under Developer options.

Open it in Hyprland with:

```text
Super + I
Super + comma
```

The old direct command panels still exist for scripting and fallback use, but normal users should live in the Settings app.

## Wallpaper repair

If the wallpaper does not load, run:

```sh
hyprglass settings
```

Open **Developer options → Wallpaper repair**. It copies the bundled wallpaper and rewrites `~/.config/hypr/hyprpaper.conf` using the current `wallpaper { ... }` block syntax with an absolute path. From a terminal you can also run:

```sh
hyprglass repair
```

## Required packages/services

Core Arch packages are listed in `packages/arch-core.txt`. CachyOS packages are listed in `packages/cachyos-core.txt`. Optional packages are listed in `packages/arch-optional.txt` and are not installed by default.

System services optionally enabled by the installer: `NetworkManager.service`, `bluetooth.service`, `ModemManager.service`, and `power-profiles-daemon.service`.

## CachyOS

CachyOS is supported as a first-class Arch-compatible target. The installer auto-detects `/etc/os-release`, but `--distro=cachyos` forces the CachyOS package profile. Use `--rate-mirrors` to run `cachyos-rate-mirrors` before installing packages on CachyOS. Use `--auto-hardware` only when you explicitly want Hyprglass to call `chwd -a` during setup.

Hyprglass does not install `linux-cachyos` automatically. Kernel replacement belongs to the OS installer or to an explicit user action, not to a desktop rice installer.

More detail: `docs/cachyos.md`.

## Limits

Runtime compositor checks require a real Hyprland session. Bluetooth, LTE/5G, battery, audio, and monitor state require matching hardware and services. Display changes are conservative so Settings does not black-screen users.
