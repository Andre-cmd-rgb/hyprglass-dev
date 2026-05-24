# Hyprglass V1

Hyprglass is a fast, glassy Wayland desktop built on Hyprland for focused work.

It is not a distro, not an ISO, not a GNOME/KDE clone, not a profile-switching rice dump, and not an Electron settings app. V1 turns a clean Arch install into a polished laptop-oriented Hyprland workstation with Kitty, Waybar, fuzzel, hyprlock, hypridle, mako, and one central terminal-native Settings app.

## Install

Dry run first:

```sh
./install.sh --dry-run
```

Install on a clean Arch system:

```sh
./install.sh
```

The installer asks first-setup questions for theme, accent, keyboard layout, display scale, and optional LTE/5G modem defaults. For non-interactive installs:

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

`hyprglass settings` is now the main control surface. It covers appearance, accent color, display scale, keyboard layout, wallpaper repair, Wi-Fi, Bluetooth, modem status, modem autounlock/autoconnect, audio, power, services, updates, and doctor checks.

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

Then choose **Wallpaper repair**. It copies the bundled wallpaper and rewrites `~/.config/hypr/hyprpaper.conf` using the current `wallpaper { ... }` block syntax with an absolute path. From a terminal you can also run:

```sh
hyprglass repair
```

## Required packages/services

Core packages are listed in `packages/arch-core.txt`. Optional packages are listed in `packages/arch-optional.txt` and are not installed by default.

System services optionally enabled by the installer: `NetworkManager.service`, `bluetooth.service`, `ModemManager.service`, and `power-profiles-daemon.service`.

## Limits

Runtime compositor checks require a real Hyprland session. Bluetooth, LTE/5G, battery, audio, and monitor state require matching hardware and services. Display changes are conservative so Settings does not black-screen users.
