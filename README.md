# Hyprglass V1.5

Hyprglass is a fast, glassy Wayland desktop built on Hyprland for focused work.

Version is stored in `VERSION` and injected into the Go binary at build time.

It is not a distro, not an ISO, not a GNOME/KDE clone, not a profile-switching rice dump, and not an Electron settings app. V1.5 turns a clean Arch or CachyOS install into a polished laptop-oriented Hyprland workstation with Kitty, Waybar, fuzzel, hyprlock, hypridle, mako, and one central terminal-native Settings app.

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

The installer asks first-setup questions for theme, accent, keyboard layout, display scale, boot login manager, hardware/session services, optional LTE/5G modem autoconnect, and optional CachyOS hardware auto-configuration. For non-interactive installs:

```sh
./install.sh --yes
```

By default, a first install configures `greetd` + `tuigreet` so the machine boots to a lightweight login that starts Hyprland. Disable that with:

```sh
./install.sh --no-login-manager
```

Passwordless boot straight into Hyprland is available only when explicitly requested:

```sh
./install.sh --autologin
```

After install, open a new terminal or run:

```sh
exec $SHELL -l
```

Skip packages:

```sh
./install.sh --no-packages --yes
```

Update an existing checkout and refresh configs without overwriting existing display/monitor rules:

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
hyprglass settings apply --with-display
hyprglass doctor
hyprglass doctor --json
hyprglass icons status
hyprglass icons repair
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


## Display safety

Hyprglass does **not** rewrite monitor rules during normal Settings applies. Appearance, icon, wallpaper, network, audio, and update actions will not touch `~/.config/hypr/conf.d/monitors.conf`.

Display changes are scoped to **Settings → Display, scaling, and keyboard** or the explicit command:

```sh
hyprglass settings apply --with-display
```

The display file uses a managed block:

```text
# >>> hyprglass managed display >>>
monitor = , preferred, auto, auto
# <<< hyprglass managed display <<<
```

Hyprglass only edits scale fields inside that block. If the line inside the block is customized, such as `monitor = eDP-1, 3840x2400@60, 0x0, 1.75`, changing scale turns it into `monitor = eDP-1, 3840x2400@60, 0x0, 2` and keeps the custom resolution and position. Manual laptop/external-monitor rules outside the block are preserved. If an unmarked manual monitor config exists, Hyprglass refuses to overwrite it instead of risking a black screen. The installer also preserves any existing `monitors.conf` during updates.

## Icon/font repair

Waybar icons use Nerd Font private glyphs. Hyprglass now installs both `ttf-jetbrains-mono-nerd` and `ttf-nerd-fonts-symbols-mono`, refreshes `fc-cache`, and writes a Waybar CSS fallback stack that includes `Symbols Nerd Font Mono`.

If top-bar icons show as squares after updating an old install:

```sh
hyprglass icons repair
hyprglass settings apply
```

The same repair is available under **Settings → Developer options → Icon/font repair**.

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

System services optionally enabled by the installer: `NetworkManager.service`, `bluetooth.service`, `ModemManager.service`, `power-profiles-daemon.service`, and `greetd.service` for boot login when selected.

## CachyOS

CachyOS is supported as a first-class Arch-compatible target. The installer auto-detects `/etc/os-release`, but `--distro=cachyos` forces the CachyOS package profile. Use `--rate-mirrors` to run `cachyos-rate-mirrors` before installing packages on CachyOS. Use `--auto-hardware` only when you explicitly want Hyprglass to call `chwd -a` during setup.

Hyprglass does not install `linux-cachyos` automatically. Kernel replacement belongs to the OS installer or to an explicit user action, not to a desktop rice installer.

More detail: `docs/cachyos.md`.

## Limits

Runtime compositor checks require a real Hyprland session. Bluetooth, LTE/5G, battery, audio, and monitor state require matching hardware and services. Display changes are conservative so Settings does not black-screen users.
