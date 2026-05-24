# Hyprglass

Hyprglass is a fast, glassy Wayland desktop built on Hyprland for focused work.

It is not a distro, not an ISO, not a GNOME/KDE clone, not a profile-switching rice dump, and not an Electron settings app. V0 turns a clean Arch install into a polished Hyprland workstation with Kitty, Waybar, fuzzel, hyprlock, hypridle, mako, and terminal-native Hyprglass tools.

## Install

Dry run first:

```sh
./install.sh --dry-run
```

Install:

```sh
./install.sh --yes
```

After install, open a new terminal or run:

```sh
exec $SHELL -l
```

The installer adds `~/.local/bin` to your shell startup files so `hyprglass` is available on `PATH`.

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

The updater stashes local repo changes before pulling, re-runs the installer after
pulling, refreshes configs, and restarts Hyprland session components when run
inside Hyprland. You do not need to run `git commit` for normal updates.

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
hyprglass doctor
hyprglass doctor --json
hyprglass wifi
hyprglass bluetooth
hyprglass lte
hyprglass audio
hyprglass display
hyprglass laptop
hyprglass settings
hyprglass power
hyprglass wallpaper apply
hyprglass touchid status
```

`hyprglass laptop` is the laptop control surface: battery, power profile,
thermal/fan readings, sleep state, LTE, and fingerprint status. It can set
`power-saver`, `balanced`, or `performance` through `powerprofilesctl` when the
hardware/driver stack exposes those profiles.

`hyprglass wallpaper apply` installs the Hyprglass wallpaper into
`~/.config/hypr/assets/wallpapers/`, refreshes `hyprpaper.conf`, and restarts
hyprpaper inside a Hyprland session. `hyprglass touchid` checks and runs fprintd
enrollment/verification, but it does not edit PAM automatically.

## Shortcuts

See `docs/shortcuts.md`.

## Required packages/services

Core packages are listed in `packages/arch-core.txt`. Optional packages are listed in `packages/arch-optional.txt` and are not installed by default.

System services optionally enabled by the installer: `NetworkManager.service`, `bluetooth.service`, `ModemManager.service`.

## Known limitations

Runtime compositor checks require a real Hyprland session. Bluetooth, LTE, battery, audio, and monitor state require matching hardware and services. Display writes are intentionally conservative in V0 to avoid black-screening users.
