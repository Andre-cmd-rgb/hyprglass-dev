# Hyprglass

A fast, glassy Wayland desktop built on Hyprland for focused work. Turns a clean Arch or CachyOS install into a polished laptop-oriented workstation with Kitty, Waybar, fuzzel, hyprlock, hypridle, mako, and a terminal-native Settings app.

## Install

Clone and run:

```sh
git clone https://github.com/Andre-cmd-rgb/hyprglass-dev && cd hyprglass-dev && ./install.sh
```

Preview what would happen first:

```sh
./install.sh --dry-run
```

CachyOS:

```sh
./install.sh --distro=cachyos
```

CachyOS with mirror ranking:

```sh
./install.sh --distro=cachyos --rate-mirrors
```

Non-interactive install with defaults:

```sh
./install.sh --yes
```

The installer asks about theme, accent, keyboard layout, display scale, boot login manager, hardware services, and optional LTE/5G modem autoconnect. Pass `--yes` to skip all prompts and accept defaults.

Boot login via `greetd` + `tuigreet` is configured by default. Disable it:

```sh
./install.sh --no-login-manager
```

Passwordless autologin (private machines only):

```sh
./install.sh --autologin
```

After install, open a new terminal or run `exec $SHELL -l` to pick up the updated PATH.

## Update

```sh
hyprglass update
```

or from the repo:

```sh
./install.sh --update
```

## Commands

```sh
hyprglass --help
hyprglass settings          # open Settings (Super+I or Super+comma in Hyprland)
hyprglass doctor [--json]   # check environment
hyprglass icons repair      # fix missing Waybar icons
hyprglass repair            # repair wallpaper and restart session components
hyprglass wifi
hyprglass bluetooth
hyprglass lte
hyprglass audio
hyprglass display
hyprglass laptop
hyprglass power
hyprglass system [--json | rate-mirrors | chwd-list]
hyprglass wallpaper apply
hyprglass touchid status
```

## Settings

`hyprglass settings` is the main control surface. Open it in Hyprland with **Super+I** or **Super+comma**.

Sections: Appearance, Display/Keyboard, Network/Bluetooth/Modem, Audio, Power, Update, Developer options.

## Display safety

Settings never rewrites `~/.config/hypr/conf.d/monitors.conf` during normal applies. Display changes require an explicit action: **Settings → Display, scaling, and keyboard** or `--with-display`.

The installer also preserves any existing `monitors.conf` during updates — it will not overwrite a custom laptop/external-monitor layout.

## Icon repair

Waybar icons use Nerd Font glyphs. If icons appear as squares after an update:

```sh
hyprglass icons repair
hyprglass settings apply
```

Or: **Settings → Developer options → Icon/font repair**.

## Uninstall

```sh
./uninstall.sh
```

## Run checks

```sh
./scripts/check.sh
```

## Required packages

`packages/arch-core.txt` (Arch) and `packages/cachyos-core.txt` (CachyOS). Optional packages in `packages/arch-optional.txt` are not installed by default.

## CachyOS

CachyOS is supported as a first-class target. The installer auto-detects `/etc/os-release`; use `--distro=cachyos` to force the CachyOS profile. Use `--rate-mirrors` to rank mirrors before installing. Use `--auto-hardware` to run `chwd -a` during setup.

Hyprglass does not replace the kernel automatically. See `docs/cachyos.md` for details.

## Version

`1.5.0` — stored in `VERSION`, injected into the binary at build time.
